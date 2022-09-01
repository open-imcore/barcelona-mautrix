// swift-tools-version: 5.4
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

extension Target.Dependency {
    static func paris(_ name: String) -> Target.Dependency {
        .product(name: name, package: "Paris")
    }
    
    static func barcelona(_ name: String) -> Target.Dependency {
        .product(name: name, package: "Barcelona")
    }
}

extension Package {
    func addingLibrary(name: String, dependencies: [Target.Dependency] = []) -> Package {
        products.append(.library(name: name, targets: [name]))
        targets.append(.target(name: name, dependencies: dependencies))
        return self
    }
}

extension Array {
    static func paris(_ names: String...) -> [Target.Dependency] {
        names.map { .paris($0) }
    }
    
    static func barcelona(_ names: String...) -> [Target.Dependency] {
        names.map { .barcelona($0) }
    }
}

let package = Package(
    name: "barcelona-mautrix",
    platforms: [
        .iOS(.v13), .macOS(.v10_15)
    ],
    products: [
        .executable(name: "barcelona-mautrix", targets: ["barcelona-mautrix"])
    ],
    dependencies: [
//         Dependencies declare other packages that this package depends on.
//         .package(url: /* package url */, from: "1.0.0"),
        .package(name: "Barcelona", url: "https://github.com/open-imcore/barcelona", .revisionItem("b1a39aad6fac795a72bb3cb464ba900bc9b34225")),
//        .package(name: "Barcelona", path: "../../barcelona"),
        .package(url: "https://github.com/SwiftyContacts/SwiftyContacts", .upToNextMajor(from: "4.0.0")),
        .package(url: "https://github.com/EricRabil/ERBufferedStream", .upToNextMajor(from: "1.0.4")),
        .package(name: "SwiftProtobuf", url: "https://github.com/apple/swift-protobuf", from: "1.20.0"),
        .package(name: "GRPC", url: "https://github.com/grpc/grpc-swift.git", from: "1.9.0"),
        .package(url: "https://github.com/EricRabil/Yammit", .upToNextMajor(from: "1.0.2"))
    ],
    targets: [
        .executableTarget(name: "barcelona-mautrix", dependencies: [
            "BarcelonaMautrixIPC"
        ], resources: [.copy("barcelona-mautrix.entitlements")])
    ]
)
.addingLibrary(name: "BarcelonaMautrixIPC", dependencies: [
    .barcelona("Barcelona"), .barcelona("BarcelonaDB"), .barcelona("BarcelonaJS"), "SwiftyContacts", "ERBufferedStream", "Yammit"
])
