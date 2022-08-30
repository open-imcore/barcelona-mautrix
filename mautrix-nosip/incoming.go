package main

import (
	"fmt"
	"html"
	"strings"

	"go.mau.fi/imessage-nosip/imessage"
	pb "go.mau.fi/imessage-nosip/protobuf"
	log "maunium.net/go/maulogger/v2"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/appservice"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type ReplyInfo struct {
	MessageID string
	Sender    string
}

type PreparedEvent struct {
	GUID string

	Intent  *appservice.IntentAPI
	Type    event.Type
	Content interface{}
	Extra   map[string]interface{}
	Caption *event.MessageEventContent

	/// When present, the event cannot be backfilled until the designated event is bridged
	ReplyTo          *ReplyInfo
	RelatesToGUID    string
	TapbackRedaction *pb.Message
	AvatarChange     *pb.Message
	NameChange       *pb.Message
	MemberChange     *pb.Message
	ExpiresIn        uint32
	Error            string
	MediaKey         []byte
	TapbackType      *pb.TapbackType
}

type IncomingMessageIngress struct {
	Bridge *IMBridge
	Portal *Portal

	PreparedEvents []PreparedEvent
	log            log.Logger
}

func NewIncomingMessageIngress(portal *Portal) *IncomingMessageIngress {
	return &IncomingMessageIngress{
		Bridge:         portal.bridge,
		Portal:         portal,
		PreparedEvents: make([]PreparedEvent, 0),
		log:            portal.log.Sub("Incoming"),
	}
}

func (mt *IncomingMessageIngress) ConsumeMessage(msg *pb.Message) {
	intent := mt.Portal.getIntentForMessage(msg, nil)
	for _, item := range msg.Items {
		prepared := PreparedEvent{
			GUID:   item.GUID,
			Intent: intent,
			Type:   event.EventMessage,
			Extra:  make(map[string]interface{}),
		}
		switch item.X.(type) {
		case *pb.Item_Text:
			content := &event.MessageEventContent{
				MsgType: event.MsgText,
			}

			item := item.GetText()
			body := item.GetText()
			body = strings.ReplaceAll(body, "\ufffc", "")
			subject := item.GetSubject()

			if len(subject) > 0 {
				content.Body = fmt.Sprintf("**%s**\n%s", subject, body)
				content.Format = event.FormatHTML
				content.FormattedBody = fmt.Sprintf("<strong>%s</strong><br>%s", html.EscapeString(subject), html.EscapeString(body))
			} else {
				content.Body = body
			}
			prepared.Content = content
			break
		case *pb.Item_Attachment:
			var err error
			prepared.Content, prepared.Extra, err = mt.Portal.handleIMAttachment(msg, item.GetAttachment().Attachment, intent)
			if err != nil {
				prepared.Content = &event.MessageEventContent{
					MsgType: event.MsgNotice,
					Body:    err.Error(),
				}
				prepared.Extra = make(map[string]interface{})
			}
		case *pb.Item_Typing:
			continue
		case *pb.Item_Plugin:
			plugin := item.GetPlugin()
			if richLink := plugin.GetRichLink(); richLink != nil {
				mt.log.Debugfln("Handling rich link in iMessage %s", msg.GUID)
				linkPreview := mt.Portal.convertRichLinkToBeeper(richLink)
				fallbackText := plugin.GetFallbackText()
				if linkPreview != nil {
					prepared.Extra["com.beeper.linkpreviews"] = []*BeeperLinkPreview{linkPreview}
					mt.log.Debugfln("Link preview metadata converted for %s", msg.GUID)
					if len(fallbackText) == 0 {
						fallbackText = linkPreview.MatchedURL
					}
					if len(fallbackText) == 0 {
						fallbackText = linkPreview.CanonicalURL
					}
				}
				prepared.Content = &event.MessageEventContent{
					MsgType: event.MsgText,
					Body:    fallbackText,
				}
			} else {
				mt.log.Debugfln("Dropping message %s because it has no rich link", msg.GUID)
			}
		case *pb.Item_GroupNameChange:
			prepared.NameChange = msg
		case *pb.Item_GroupAvatarChange:
			prepared.AvatarChange = msg
		case *pb.Item_GroupParticipantChange:
			prepared.MemberChange = msg
		case *pb.Item_Tapback:
			tapback := item.GetTapback()

			if tapback.Tapback.IsRemove() {
				// check if we have a staged tapback that we can remove instead
				dropped := false
				for index, evt := range mt.PreparedEvents {
					if evt.Type != event.EventReaction {
						continue
					}
					if evt.TapbackType.RemovingVariant() != tapback.Tapback.Type {
						continue
					}
					if evt.ReplyTo == nil || evt.ReplyTo.MessageID != tapback.GetTapback().Target {
						continue
					}
					if evt.Extra["com.beeper.proto.network_message_sender"] != msg.Sender.ToString() {
						continue
					}
					mt.log.Debugfln("Dropping tapback %s as we never bridged it and it was later redacted by %s", item.GUID, evt.GUID)
					mt.PreparedEvents = append(mt.PreparedEvents[:index], mt.PreparedEvents[index+1:]...)
					dropped = true
					break
				}
				if !dropped {
					// actually redact
					prepared.TapbackRedaction = msg
				}
			} else {
				prepared.Type = event.EventReaction
				prepared.Content = &event.ReactionEventContent{
					RelatesTo: event.RelatesTo{
						// EventID: target.MXID,
						Type: event.RelAnnotation,
						Key:  (imessage.TapbackType)(tapback.Tapback.Type).Emoji(),
					},
				}
				prepared.RelatesToGUID = tapback.Tapback.Target
			}
		case *pb.Item_Phantom:
			mt.log.Debugfln("Witnessed phantom item %s/%s", item.GetPhantom().GetTypeString(), item.GetPhantom().GetDebugDescription())
		}
		if len(msg.Service) > 0 {
			prepared.Extra[bridgeInfoService] = msg.Service
		}
		if msg.MessageMetadata != nil {
			prepared.Extra["com.beeper.message_metadata"] = msg.GetMessageMetadata().ConvertToMap()
		}
		prepared.Extra["com.beeper.proto.network_message_origin"] = msg.ChatGUID.ToString()
		prepared.Extra["com.beeper.proto.network_message_sender"] = msg.Sender.ToString()
		prepared.Extra["com.beeper.proto.network_message_id"] = msg.GUID
		mt.PreparedEvents = append(mt.PreparedEvents, prepared)
	}
}

func (pm *PreparedEvent) realizeAvatarChange(portal *Portal) *id.EventID {
	change := pm.AvatarChange.GetItems()[0].GetGroupAvatarChange()
	if avatar := change.GetNewAvatar(); avatar != nil && change.GetAction() == pb.GroupActionType_GroupActionAdd {
		return portal.UpdateAvatar(avatar, pm.Intent)
	} else if change.GetAction() == pb.GroupActionType_GroupActionRemove {
		// TODO
	} else {
		portal.log.Warnfln("Unexpected group action type %d in avatar change item", change.GetAction())
	}
	return nil
}

func (pm *PreparedEvent) realizeMemberChange(portal *Portal) *id.EventID {
	change := pm.MemberChange.GetItems()[0].GetGroupParticipantChange()
	target := change.GetTarget()
	if len(target) == 0 {
		return nil
	}
	puppet := portal.bridge.GetPuppetByLocalID(target)
	puppet.Sync()
	if change.Action == pb.GroupActionType_GroupActionAdd {
		return portal.setMembership(pm.Intent, puppet, event.MembershipJoin, pm.MemberChange.Time.AsTime().UnixMilli())
	} else if change.Action == pb.GroupActionType_GroupActionRemove {
		// TODO make sure this won't break anything and enable it
		//return portal.setMembership(intent, puppet, event.MembershipLeave, dbMessage.Timestamp)
	} else {
		portal.log.Warnfln("Unexpected group action type %d in member change item", change.Action)
	}
	return nil
}

var (
	PortalCreationDummyEvent = event.Type{Type: "fi.mau.dummy.portal_created", Class: event.MessageEventType}
)

const firstEventIDPrefix = "__mautrix_nosip_first_event_id_"
const nextBatchGUIDStartPrefix = "__mautrix_nosip_next_batch_guid_start_"
const nextBatchIDPrefix = "__mautrix_nosip_next_batch_id_"

func (mt *IncomingMessageIngress) firstEventIDKey() string {
	return firstEventIDPrefix + mt.Portal.MXID.String()
}

func (mt *IncomingMessageIngress) loadFirstEventID() id.EventID {
	return id.EventID(mt.Portal.bridge.DB.KV.Get(mt.firstEventIDKey()))
}

func (mt *IncomingMessageIngress) createFirstEvent() {
	firstEventResp, err := mt.Portal.MainIntent().SendMessageEvent(mt.Portal.MXID, PortalCreationDummyEvent, struct{}{})
	if err != nil {
		mt.log.Errorln("Failed to send dummy event to mark portal creation:", err)
	} else {
		mt.Portal.bridge.DB.KV.Set(mt.firstEventIDKey(), firstEventResp.EventID.String())
	}
}

func (mt *IncomingMessageIngress) loadNextBatchGUID() string {
	return ""
}

func (mt *IncomingMessageIngress) nextBatchIDKey() string {
	return nextBatchIDPrefix + mt.Portal.MXID.String()
}

func (mt *IncomingMessageIngress) loadNextBatchID() id.BatchID {
	return id.BatchID(mt.Portal.bridge.DB.KV.Get(mt.nextBatchIDKey()))
}

func (mt *IncomingMessageIngress) saveNextBatchID(bid id.BatchID) {
	mt.Portal.bridge.DB.KV.Set(mt.nextBatchIDKey(), bid.String())
}

func (mt *IncomingMessageIngress) prepareBatch(isForward, isLatest bool, prevEventID id.EventID) {
	var req mautrix.ReqBatchSend

	if !isForward {
		firstEventID := mt.loadFirstEventID()
		nextBatchID := mt.loadNextBatchID()
		if firstEventID != "" || nextBatchID != "" {
			req.PrevEventID = firstEventID
			req.BatchID = nextBatchID
		} else {
			mt.log.Warnfln("Can't backfill %d messages through %s to chat: first event ID not known", 0, mt.Portal.MXID)
			return // nil
		}
	} else {
		req.PrevEventID = prevEventID
	}

	messages := make([]PreparedEvent, 0)
	deferred := make([]PreparedEvent, 0)

	/**
	- you may not backfill tapbacks to a message in the same batch
	-
	*/

	for _, msg := range messages {

	}
}
