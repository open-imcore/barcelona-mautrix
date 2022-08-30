package protobuf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PayloadCommand = isPayload_Command
type SendMessagePartType = isSendMessagePart_Part

func (mapping *Array) ConvertToSlice() (converted []interface{}) {
	if len(converted) < len(mapping.Values) {
		converted = make([]interface{}, len(mapping.Values))
	}
	for index, value := range mapping.Values {
		converted[index] = value.ConvertToNative()
	}
	return
}

func (mapping *Mapping) ConvertToMap() (converted map[string]interface{}) {
	if len(converted) < len(mapping.Mapping) {
		converted = make(map[string]interface{}, len(mapping.Mapping))
	}
	for key, value := range mapping.Mapping {
		converted[key] = value.ConvertToNative()
	}
	return
}

func (value *MetadataValue) ConvertToNative() (converted interface{}) {
	switch value.Value.(type) {
	case *MetadataValue_Array:
		converted = value.GetArray().ConvertToSlice()
	case *MetadataValue_Mapping:
		converted = value.GetMapping().ConvertToMap()
	case *MetadataValue_String_:
		converted = value.GetString_()
	case *MetadataValue_Double:
		converted = value.GetDouble()
	case *MetadataValue_Float:
		converted = value.GetFloat()
	case *MetadataValue_Int32:
		converted = value.GetInt32()
	case *MetadataValue_Int64:
		converted = value.GetInt64()
	case *MetadataValue_UInt32:
		converted = value.GetUInt32()
	case *MetadataValue_UInt64:
		converted = value.GetUInt64()
	case *MetadataValue_SInt32:
		converted = value.GetSInt32()
	case *MetadataValue_SInt64:
		converted = value.GetSInt64()
	case *MetadataValue_Fixed32:
		converted = value.GetFixed32()
	case *MetadataValue_Fixed64:
		converted = value.GetFixed64()
	case *MetadataValue_SFIxed32:
		converted = value.GetSFIxed32()
	case *MetadataValue_SFixed64:
		converted = value.GetSFixed64()
	case *MetadataValue_Bool:
		converted = value.GetBool()
	case *MetadataValue_Bytes:
		converted = value.GetBytes()
	default:
		converted = nil
	}
	return
}

func CreateMetadataValue(value interface{}) *MetadataValue {
	var result MetadataValue
	switch value.(type) {
	case *string:
		result = MetadataValue{Value: &MetadataValue_String_{value.(string)}}
	case *float32:
		result = MetadataValue{Value: &MetadataValue_Float{value.(float32)}}
	case *float64:
		result = MetadataValue{Value: &MetadataValue_Double{value.(float64)}}
	case *int32:
		result = MetadataValue{Value: &MetadataValue_Int32{value.(int32)}}
	case *int64:
		result = MetadataValue{Value: &MetadataValue_Int64{value.(int64)}}
	case *uint32:
		result = MetadataValue{Value: &MetadataValue_UInt32{value.(uint32)}}
	case *uint64:
		result = MetadataValue{Value: &MetadataValue_UInt64{value.(uint64)}}
	case *bool:
		result = MetadataValue{Value: &MetadataValue_Bool{value.(bool)}}
	case *[]byte:
		result = MetadataValue{Value: &MetadataValue_Bytes{value.([]byte)}}
	case *map[string]interface{}:
		dict := value.(map[string]interface{})
		convertedDict := make(map[string]*MetadataValue, len(dict))
		for key, value := range dict {
			convertedDict[key] = CreateMetadataValue(value)
		}
		result = MetadataValue{Value: &MetadataValue_Mapping{&Mapping{Mapping: convertedDict}}}
	case *[]interface{}:
		slice := value.([]interface{})
		convertedSlice := make([]*MetadataValue, len(slice))
		for index, value := range slice {
			convertedSlice[index] = CreateMetadataValue(value)
		}
		result = MetadataValue{Value: &MetadataValue_Array{&Array{Values: convertedSlice}}}
	}
	return &result
}

func CreateMetadataValueFromMap(mapping map[string]interface{}) *Mapping {
	if len(mapping) == 0 {
		return nil
	}
	return CreateMetadataValue(mapping).GetMapping()
}

func (tapback *Tapback) IsRemove() bool {
	switch tapback.GetType() {
	case TapbackType_TapbackRemoveDislike:
		return true
	case TapbackType_TapbackRemoveEmphasis:
		return true
	case TapbackType_TapbackRemoveLaugh:
		return true
	case TapbackType_TapbackRemoveLike:
		return true
	case TapbackType_TapbackRemoveLove:
		return true
	case TapbackType_TapbackRemoveQuestion:
		return true
	default:
		return false
	}
}

func (msg *Message) SenderText() string {
	if msg.GetIsFromMe() {
		return "self"
	}
	return msg.GetSender().GetLocalID()
}

func (tapback *Tapback) RemovingVariant() *Tapback {
	if tapback == nil {
		return nil
	} else if tapback.IsRemove() {
		return tapback
	} else {
		newTapback := Tapback{Target: tapback.Target, Type: tapback.Type}
		newTapback.Type += 1000
		return &newTapback
	}
}

func (tapback TapbackType) RemovingVariant() TapbackType {
	number := tapback.Number()
	if number > 2999 || number < 2000 {
		return TapbackType(number)
	}
	number += 1000
	return TapbackType(number)
}

func (contact *Contact) HasName() bool {
	return contact.FirstName != nil || contact.LastName != nil || contact.Nickname != nil
}

func (contact *Contact) Name() string {
	if contact.FirstName != nil {
		if contact.LastName != nil {
			return fmt.Sprintf("%s %s", contact.GetFirstName(), contact.GetLastName())
		} else {
			return contact.GetFirstName()
		}
	} else if contact.LastName != nil {
		return contact.GetLastName()
	} else if contact.Nickname != nil {
		return contact.GetNickname()
	} else if len(contact.Emails) > 0 {
		return contact.Emails[0]
	} else if len(contact.Phones) > 0 {
		return contact.Phones[0]
	}
	return ""
}

func (attachment *Attachment) Read() ([]byte, error) {
	if strings.HasPrefix(attachment.PathOnDisk, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		attachment.PathOnDisk = filepath.Join(home, attachment.PathOnDisk[2:])
	}
	return os.ReadFile(attachment.PathOnDisk)
}

func ParseIdentifier(guid string) GUID {
	if len(guid) == 0 {
		return GUID{}
	}
	parts := strings.Split(guid, ";")
	return GUID{
		Service: parts[0],
		IsGroup: parts[1] == "+",
		LocalID: parts[2],
	}
}

func (id *GUID) ToString() string {
	if len(id.GetLocalID()) == 0 {
		return ""
	}
	typeChar := '-'
	if id.GetIsGroup() {
		typeChar = '+'
	}
	return fmt.Sprintf("%s;%c;%s", id.GetService(), typeChar, id.GetLocalID())
}

func NewGUID(service string, localID string, isGroup bool) *GUID {
	guid := GUID{}
	guid.Service = service
	guid.IsGroup = isGroup
	guid.LocalID = localID
	return &guid
}
