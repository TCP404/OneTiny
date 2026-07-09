VERSION ?= v0.6.0

APP_NAME := OneTiny
CLI_NAME := onetiny-cli
BIN_DIR := build/bin
FRONTEND_DIR := frontend
LOGO_PNG := resource/logo/logo.png
APPICON := build/appicon.png
RUNTIME_ICON := internal/gui/assets/appicon.png
MAC_ICON := build/darwin/icons.icns
WINDOWS_ICON := build/windows/icon.ico
WINDOWS_MANIFEST := internal/gui/assets/windows/wails.exe.manifest
WINDOWS_INFO := internal/gui/assets/windows/info.json
WINDOWS_SYSO = cmd/gui/rsrc_windows_$(GOARCH).syso
MAC_INFO := internal/gui/assets/darwin/Info.plist
MAC_APP := $(BIN_DIR)/$(APP_NAME).app

GUI_MAIN := ./cmd/gui
CLI_MAIN := ./cmd/cli

GOARCH ?= $(shell go env GOARCH)
HOST_GOOS := $(shell go env GOOS)

GO_LDFLAGS := -s -w -X github.com/tcp404/OneTiny/internal/version.Version=$(VERSION)
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
	DEFAULT_BUILD_TARGET := package-mac
else
	DEFAULT_BUILD_TARGET := build-gui
endif

.DEFAULT_GOAL := help

.PHONY: help all dev build release build-gui build-cli package-mac package-windows package-linux compress clean
.PHONY: gui cli package frontend icons windows-resource

help:
	@printf "OneTiny Makefile\n"
	@printf "\nUsage:\n"
	@printf "  make <target>\n"
	@printf "\nMain targets:\n"
	@printf "  %-18s %s\n" "help" "显示这份命令说明"
	@printf "  %-18s %s\n" "dev" "日常开发构建：只生成 GUI 二进制，不打包"
	@printf "  %-18s %s\n" "build" "当前平台默认桌面产物；macOS 会生成 .app"
	@printf "  %-18s %s\n" "release" "当前平台发布产物；Windows/Linux 暂时只生成 GUI 二进制"
	@printf "\nBuild targets:\n"
	@printf "  %-18s %s\n" "build-gui" "构建 GUI 二进制到 build/bin"
	@printf "  %-18s %s\n" "build-cli" "构建 CLI 二进制到 build/bin"
	@printf "\nPackage targets:\n"
	@printf "  %-18s %s\n" "package-mac" "macOS: 生成 build/bin/OneTiny.app"
	@printf "  %-18s %s\n" "package-windows" "Windows: 暂未实现安装包/压缩包"
	@printf "  %-18s %s\n" "package-linux" "Linux: 暂未实现 AppImage/deb/rpm/tar.gz"
	@printf "\nMaintenance:\n"
	@printf "  %-18s %s\n" "compress" "用 UPX 压缩 GUI 二进制"
	@printf "  %-18s %s\n" "clean" "清理 build/bin 和 Windows .syso 文件"
	@printf "\nCompatibility aliases:\n"
	@printf "  %-18s %s\n" "all" "等同 build"
	@printf "  %-18s %s\n" "gui" "等同 build-gui"
	@printf "  %-18s %s\n" "cli" "等同 build-cli"
	@printf "  %-18s %s\n" "package" "等同 package-mac"
	@printf "\nInternal targets:\n"
	@printf "  %-18s %s\n" "frontend" "安装并构建前端资源"
	@printf "  %-18s %s\n" "icons" "生成 macOS/Windows 图标资源"
	@printf "  %-18s %s\n" "windows-resource" "生成 Windows .syso 资源"

all: build

dev: build-gui

build: $(DEFAULT_BUILD_TARGET)

ifeq ($(HOST_GOOS),darwin)
release: package-mac
else
release: build-gui
	@echo "release: $(HOST_GOOS) package target is not implemented; produced $(GUI_BINARY)"
endif

frontend:
	npm install --prefix $(FRONTEND_DIR)
	npm run build --prefix $(FRONTEND_DIR)

icons: $(RUNTIME_ICON)
	mkdir -p build build/darwin build/windows
	cp $(RUNTIME_ICON) $(APPICON)
	wails3 generate icons -input $(APPICON) -macfilename $(MAC_ICON) -windowsfilename $(WINDOWS_ICON)

$(RUNTIME_ICON): $(LOGO_PNG)
	mkdir -p internal/gui/assets
	cp $(LOGO_PNG) $(RUNTIME_ICON)

windows-resource: icons
	wails3 generate syso -arch $(GOARCH) -icon $(WINDOWS_ICON) -manifest $(WINDOWS_MANIFEST) -info $(WINDOWS_INFO) -out $(WINDOWS_SYSO)

build-gui: frontend icons $(GUI_DEPS)
	mkdir -p $(BIN_DIR)
	go build $(GO_BUILD_FLAGS) -o $(GUI_BINARY) $(GUI_MAIN)

build-cli:
	mkdir -p $(BIN_DIR)
	go build $(GO_BUILD_FLAGS) -o $(CLI_BINARY) $(CLI_MAIN)

ifeq ($(HOST_GOOS),darwin)
package-mac: build-gui icons
	rm -rf $(MAC_APP)
	mkdir -p $(MAC_APP)/Contents/MacOS $(MAC_APP)/Contents/Resources
	cp $(GUI_BINARY) $(MAC_APP)/Contents/MacOS/$(APP_NAME)
	cp $(MAC_ICON) $(MAC_APP)/Contents/Resources/icons.icns
	cp $(MAC_INFO) $(MAC_APP)/Contents/Info.plist
	codesign --force --deep --sign - $(MAC_APP)
else
package-mac:
	@echo "package-mac must be run on macOS"
	@exit 1
endif

package-windows:
	@echo "package-windows is not implemented yet; use make build-gui to create $(GUI_BINARY)"
	@exit 1

package-linux:
	@echo "package-linux is not implemented yet; use make build-gui to create $(GUI_BINARY)"
	@exit 1

gui: build-gui

cli: build-cli

package: package-mac

compress: build-gui
	$(UPX) $(UPX_FLAGS) $(GUI_BINARY)

clean:
	rm -rf $(BIN_DIR)
	rm -f cmd/gui/*.syso
