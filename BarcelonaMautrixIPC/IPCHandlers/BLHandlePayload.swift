//
//  BLHandlePayload.swift
//  BarcelonaFoundation
//
//  Created by Eric Rabil on 8/23/21.
//  Copyright © 2021 Eric Rabil. All rights reserved.
//

import Foundation
import Barcelona
import IMCore
import Swog

internal let IPCLog = Logger(category: "BLIPC")

private extension IPCPayload {
    var runnable: Runnable? {
        switch command {
        case .send_message(let req):
            return req
        case .send_media(let req):
            return req
        case .send_tapback(let req):
            return req
        case .send_read_receipt(let req):
            return req
        case .set_typing(let req):
            return req
        case .get_chats(let req):
            return req
        case .get_chat(let req):
            return req
        case .get_chat_avatar(let req):
            return req
        case .get_contact(let req):
            return req
        case .get_messages_after(let req):
            return req
        case .get_recent_messages(let req):
            return req
        default:
            return nil
        }
    }
}

public func BLHandlePayload(_ payload: IPCPayload) {
    if payload.command.name == .ping {
        payload.reply(withResponse: .ack)
        return
    }
    
    guard let runnable = payload.runnable else {
        return IPCLog.warn("Received unhandleable payload type \(payload.command.name)")
    }
    
    runnable.run(payload: payload)
}
