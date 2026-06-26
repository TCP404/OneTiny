VERSION ?= v0.6.0

APP_NAME := OneTiny
CLI_NAME := onetiny-cli
BIN_DIR := build/bin
FRONTEND_DIR := frontend
LOGO_SVG := README/logo.svg
APPICON := build/appicon.png
RUNTIME_ICON := internal/gui/assets/appicon.png
MAC_ICON := build/darwin/icons.icns
WINDOWS_ICON := build/windows/icon.ico
WINDOWS_MANIFEST := build/windows/wails.exe.manifest
WINDOWS_INFO := build/windows/info.json
WINDOWS_SYSO = cmd/onetiny-gui/rsrc_windows_$(GOARCH).syso
MAC_APP := $(BIN_DIR)/$(APP_NAME).app

GUI_MAIN := ./cmd/onetiny-gui
CLI_MAIN := .

GOARCH ?= $(shell go env GOARCH)
HOST_GOOS := $(shell go env GOOS)

GO_LDFLAGS := -s -w -X github.com/TCP404/OneTiny-cli/internal/constant.VERSION=$(VERSION)
UPX ?= upx
UPX_FLAGS ?= --best

ifeq ($(OS),Windows_NT)
	GUI_BINARY := $(BIN_DIR)/$(APP_NAME).exe
	CLI_BINARY := $(BIN_DIR)/$(CLI_NAME).exe
	GUI_DEPS := windows-resource
	GO_LDFLAGS += -H windowsgui
else
	GUI_BINARY := $(BIN_DIR)/$(APP_NAME)
	CLI_BINARY := $(BIN_DIR)/$(CLI_NAME)
	GUI_DEPS :=
endif

GO_BUILD_FLAGS := -ldflags "$(GO_LDFLAGS)"

ifeq ($(HOST_GOOS),darwin)
	BUILD_TARGETS := frontend icons gui package
else
	BUILD_TARGETS := frontend icons gui
endif

.PHONY: all build frontend icons windows-resource gui cli package compress clean

all: build

build: $(BUILD_TARGETS)

frontend:
	npm install --prefix $(FRONTEND_DIR)
	npm run build --prefix $(FRONTEND_DIR)

icons: $(RUNTIME_ICON)
	mkdir -p build build/darwin build/windows
	cp $(RUNTIME_ICON) $(APPICON)
	wails3 generate icons -input $(APPICON) -macfilename $(MAC_ICON) -windowsfilename $(WINDOWS_ICON)

$(RUNTIME_ICON): $(LOGO_SVG)
	command -v rsvg-convert >/dev/null 2>&1 || (echo "rsvg-convert is required to regenerate $(RUNTIME_ICON) from $(LOGO_SVG)" && exit 1)
	mkdir -p internal/gui/assets
	rsvg-convert -w 1024 -h 1024 -o $(RUNTIME_ICON) $(LOGO_SVG)

windows-resource: icons
	wails3 generate syso -arch $(GOARCH) -icon $(WINDOWS_ICON) -manifest $(WINDOWS_MANIFEST) -info $(WINDOWS_INFO) -out $(WINDOWS_SYSO)

gui: icons $(GUI_DEPS)
	mkdir -p $(BIN_DIR)
	go build $(GO_BUILD_FLAGS) -o $(GUI_BINARY) $(GUI_MAIN)

cli:
	mkdir -p $(BIN_DIR)
	go build $(GO_BUILD_FLAGS) -o $(CLI_BINARY) $(CLI_MAIN)

ifeq ($(HOST_GOOS),darwin)
package: gui icons
	rm -rf $(MAC_APP)
	mkdir -p $(MAC_APP)/Contents/MacOS $(MAC_APP)/Contents/Resources
	cp $(GUI_BINARY) $(MAC_APP)/Contents/MacOS/$(APP_NAME)
	cp $(MAC_ICON) $(MAC_APP)/Contents/Resources/icons.icns
	cp build/darwin/Info.plist $(MAC_APP)/Contents/Info.plist
	codesign --force --deep --sign - $(MAC_APP)
else
package:
	@echo "package currently creates a macOS .app bundle and must be run on macOS"
endif

compress: gui
	$(UPX) $(UPX_FLAGS) $(GUI_BINARY)

clean:
	rm -rf $(BIN_DIR)
	rm -f cmd/onetiny-gui/*.syso
