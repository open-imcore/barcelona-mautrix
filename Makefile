RELEASE_BUILD_ARGS=--configuration release --arch arm64 --arch x86_64
DEBUG_BUILD_ARGS=--configuration debug
BUILD_ARGS?=$(DEBUG_BUILD_ARGS)
BINARY_PATH=$(shell swift build $(BUILD_ARGS) --show-bin-path)

CODESIGN_IDENTITY?="-"
ENTITLEMENTS_PATH=$(PWD)/Sources/barcelona-mautrix/barcelona-mautrix.entitlements
KC_PASSWORD?=""
KC_PATH?=login.keychain

sign:
	codesign -fs $(CODESIGN_IDENTITY) --keychain "$(KC_PATH)" --entitlements $(ENTITLEMENTS_PATH) $(BINARY_PATH)/barcelona-mautrix

ci-sign:
	security unlock-keychain -p $(KC_PASSWORD) ci.keychain
	make sign

build:
	swift build $(BUILD_ARGS) -Xlinker -F/System/Library/PrivateFrameworks
	
product-path:
	echo $(BINARY_PATH)/barcelona-mautrix
	
all: build sign
