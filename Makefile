RELEASE_BUILD_ARGS=--configuration release --arch arm64 --arch x86_64
DEBUG_BUILD_ARGS=--configuration debug
BUILD_ARGS?=$(DEBUG_BUILD_ARGS)
BINARY_PATH=$(shell swift build $(BUILD_ARGS) --show-bin-path)

CODESIGN_IDENTITY?="-"
ENTITLEMENTS_PATH=$(PWD)/Sources/barcelona-mautrix/barcelona-mautrix.entitlements
KC_PASSWORD?=""
KC_PATH?=login.keychain

IPC_BUILD_DIR = ./ipc-built

sign:
	codesign -fs $(CODESIGN_IDENTITY) --keychain "$(KC_PATH)" --entitlements $(ENTITLEMENTS_PATH) $(BINARY_PATH)/barcelona-mautrix

ci-sign:
	security unlock-keychain -p $(KC_PASSWORD) ci.keychain
	make sign

build: ipc-v1-swift
	swift build -Xcc -Wno-nullability-completeness -Xcc -Wno-incomplete-umbrella -Xcc -Wno-objc-protocol-method-implementation -Xcc -Wno-arc-performSelector-leaks -Xcc -Wno-strict-prototypes -Xcc -Wno-property-attribute-mismatch -Xcc -Wno-visibility $(BUILD_ARGS)

product-path:
	echo $(BINARY_PATH)/barcelona-mautrix
	
all: build sign

rstebuild:
	make build sign CODESIGN_IDENTITY='"Eric Rabil Self-Trusted Execution"'

# Mautrix

mautrix-arm: ipc-v1-go
	cd mautrix-nosip && CGO_FLAGS='-I/opt/homebrew/include/' CGO_LDFLAGS='-L/opt/homebrew/lib' go build

mautrix-intel-on-arm: ipc-v1-go
	cd mautrix-nosip && arch -x86_64 -e CGO_CFLAGS='-I/usr/local/include/' -e CGO_LDFLAGS='-L/usr/local/lib/' /usr/local/bin/go build

mautrix-intel: ipc-v1-go
	cd mautrix-nosip && CGO_CFLAGS='-I/usr/local/include/' CGO_LDFLAGS='-L/usr/local/lib/' /usr/local/bin/go build

mautrix-run:
	cd mautrix-nosip && ./mautrix-imessage

# IPC

ipc-mac-deps:
	brew install swift-protobuf protoc-gen-go

swift_opts=--swift_opt=Visibility=Public
swift_out=Sources/BarcelonaMautrixIPCProtobuf
go_out=$(PWD)

ipc-v1-%:
	$(eval OUT:= $(or $($*_out), $(IPC_BUILD_DIR)/v1/$*))
	mkdir -p $(OUT)
	protoc '--$*_out=$(OUT)' $($*_opts) ipc/v1/v1.proto

ipc-v1: ipc-v1-go ipc-v1-swift