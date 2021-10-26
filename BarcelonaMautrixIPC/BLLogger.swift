//
//  BLLogger.swift
//  BarcelonaMautrixIPC
//
//  Created by Eric Rabil on 5/28/21.
//  Copyright © 2021 Eric Rabil. All rights reserved.
//

import Foundation
import BarcelonaFoundation

private let BLDefaultModule = ""

private extension LoggingLevel {
    var ipcLevel: IPCLoggingLevel {
        switch self {
        case .info:
            return .info
        case .warn:
            return .warn
        case .debug:
            return .debug
        case .fault:
            return .fatal
        case .error:
            return .error
        }
    }
}

public class BLMautrixSTDOutDriver: LoggingDriver {
    public static let shared = BLMautrixSTDOutDriver()
    
    private init() {}
    
    public func log(level: LoggingLevel, fileID: StaticString, line: Int, function: StaticString, dso: UnsafeRawPointer, category: StaticString, message: StaticString, args: [CVarArg]) {
        BLWritePayload(.init(id: nil, command: .log(LogCommand(level: level.ipcLevel, module: String(category), message: String(format: String(message), arguments: args), metadata: [
            "fileID": fileID.description,
            "function": function.description,
            "line": line.description
        ]))), log: false)
    }
    
    public func log(level: LoggingLevel, module: String, message: BackportedOSLogMessage, metadata: [String: Encodable]) {
        BLWritePayload(.init(id: nil, command: .log(LogCommand(level: level.ipcLevel, module: module, message: message.render(level: BLRuntimeConfiguration.privacyLevel), metadata: metadata))))
    }
    
    public func log(level: LoggingLevel, fileID: StaticString, line: Int, function: StaticString, dso: UnsafeRawPointer, category: StaticString, message: BackportedOSLogMessage) {
        BLWritePayload(.init(id: nil, command: .log(LogCommand(level: level.ipcLevel, module: String(category), message: message.render(level: BLRuntimeConfiguration.privacyLevel), metadata: [
            "fileID": fileID.description,
            "function": function.description,
            "line": line.description
        ]))), log: false)
    }
}
