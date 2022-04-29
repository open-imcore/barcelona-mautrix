//
//  BLMessageStatus.swift
//  BarcelonaMautrixIPC
//
//  Created by Eric Rabil on 5/24/21.
//  Copyright © 2021 Eric Rabil. All rights reserved.
//

import Foundation
import Barcelona

public struct BLMessageStatus: Codable {
    public init(guid: String, chatGUID: String, status: BLMessageStatus.StatusEvent, message: String? = nil, statusCode: String? = nil) {
        self.guid = guid
        self.chatGUID = chatGUID
        self.status = status
        self.message = message
        self.statusCode = statusCode
    }
    
    public init?(event: CBMessageStatusChange) {
        switch event.type {
        case .notDelivered:
            guid = event.messageID
            chatGUID = event.chat.senderCorrelatableGUID
            status = .failed
            message = event.message.errorCode.localizedDescription
            statusCode = event.message.errorCode.description
        default:
            return nil
        }
    }
    
    public init(sentMessageGUID: String, forChatGUID: String) {
        guid = sentMessageGUID
        chatGUID = forChatGUID
        status = .sent
        message = nil
        statusCode = nil
    }
    
    public enum StatusEvent: String, Codable {
        case sent, failed
    }
    
    public var guid: String
    public var chatGUID: String
    public var status: StatusEvent
    public var message: String?
    public var statusCode: String?
    
    public enum CodingKeys: String, CodingKey {
        case guid, chatGUID = "chat_guid", status, message, statusCode = "status_code"
    }
}
