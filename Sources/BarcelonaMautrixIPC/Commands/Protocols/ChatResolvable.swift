//
//  ChatResolvable.swift
//  BarcelonaMautrixIPC
//
//  Created by Eric Rabil on 5/24/21.
//  Copyright © 2021 Eric Rabil. All rights reserved.
//

import Foundation
import Barcelona
import IMCore

public protocol ChatResolvable {
    var chat_guid: String { get set }
}

public extension ChatResolvable {
    var chat: IMChat? {
        if let chat = IMChatRegistry.shared.existingChat(withGUID: chat_guid) {
            return chat
        } else {
            var parsed = ParsedGUID(rawValue: chat_guid)
            
            let service = parsed.service == "iMessage" ? IMServiceStyle.iMessage : .SMS
            let id = parsed.last
            
            if id.isPhoneNumber || id.isEmail || id.isBusinessID {
                return Chat.directMessage(withHandleID: id, service: service).imChat
            } else {
                parsed.service = service == .iMessage ? "SMS" : "iMessage"
                return IMChatRegistry.shared.existingChat(withGUID: parsed.description)
            }
        }
    }
    
    var cbChat2: CBChat? {
        CBChatRegistry.shared.chats[.guid(chat_guid)]
    }
    
    var cbChat: Chat? {
        guard let chat = chat else {
            return nil
        }
        
        return Chat(chat)
    }
    
    var blChat: BLChat? {
        chat?.blChat
    }
}
