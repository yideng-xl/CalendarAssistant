# Makefile for CalendarAssistant

GO_BIN=/opt/homebrew/bin/go
BINARY_NAME=calendar-assistant
VERSION=0.2.0
DIST_DIR=dist
APP_NAME=CalendarAssistant.app
PLIST_NAME=com.yideng.calendar-assistant.plist
DMG_NAME=CalendarAssistant-v$(VERSION)-mac.dmg
WIN_EXE=calendar-assistant-v$(VERSION)-win.exe

.PHONY: all build-mac build-win bundle dmg clean setup install uninstall release

all: release

setup:
	@mkdir -p $(DIST_DIR)
	@$(GO_BIN) mod tidy

# 构建 macOS Universal Binary (Intel + Apple Silicon)
build-mac: setup
	@echo "🍎 Building macOS Apple Silicon binary..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GO_BIN) build -o $(DIST_DIR)/$(BINARY_NAME)-arm64 cmd/calendar-assistant/main.go
	@echo "💻 Building macOS Intel binary..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GO_BIN) build -o $(DIST_DIR)/$(BINARY_NAME)-amd64 cmd/calendar-assistant/main.go
	@echo "🔗 Creating Universal Binary using lipo..."
	@lipo -create $(DIST_DIR)/$(BINARY_NAME)-arm64 $(DIST_DIR)/$(BINARY_NAME)-amd64 -output $(DIST_DIR)/$(BINARY_NAME)-universal
	@chmod +x $(DIST_DIR)/$(BINARY_NAME)-universal

# 封装并压缩 macOS .app
bundle: build-mac
	@echo "📦 Creating macOS App Bundle..."
	@rm -rf $(DIST_DIR)/$(APP_NAME)
	@mkdir -p $(DIST_DIR)/$(APP_NAME)/Contents/MacOS
	@mkdir -p $(DIST_DIR)/$(APP_NAME)/Contents/Resources
	@cp $(DIST_DIR)/$(BINARY_NAME)-universal $(DIST_DIR)/$(APP_NAME)/Contents/MacOS/$(BINARY_NAME)
	@if [ -f assets/AppIcon.icns ]; then cp assets/AppIcon.icns $(DIST_DIR)/$(APP_NAME)/Contents/Resources/AppIcon.icns; fi
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '<plist version="1.0"><dict>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '    <key>CFBundleExecutable</key><string>$(BINARY_NAME)</string>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '    <key>CFBundleIdentifier</key><string>com.yideng.calendar-assistant.v$(VERSION)</string>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '    <key>CFBundleShortVersionString</key><string>$(VERSION)</string>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '    <key>CFBundleName</key><string>CalendarAssistant</string>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '    <key>CFBundleIconFile</key><string>AppIcon</string>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '    <key>CFBundlePackageType</key><string>APPL</string>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '    <key>LSUIElement</key><string>1</string>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '    <key>NSCalendarUsageDescription</key><string>需要访问日历以同步会议安排。</string>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '    <key>NSPrincipalClass</key><string>NSApplication</string>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo '</dict></plist>' >> $(DIST_DIR)/$(APP_NAME)/Contents/Info.plist
	@echo "🔏 Signing App Bundle with entitlements..."
	@codesign --deep --force --sign - --entitlements assets/entitlements.plist $(DIST_DIR)/$(APP_NAME)

# 制作 DMG 镜像
dmg: bundle
	@echo "📀 Creating DMG Installer v$(VERSION)..."
	@rm -f $(DIST_DIR)/$(DMG_NAME)
	@rm -rf $(DIST_DIR)/dmg_temp
	@mkdir -p $(DIST_DIR)/dmg_temp/.background
	@cp -R $(DIST_DIR)/$(APP_NAME) $(DIST_DIR)/dmg_temp/
	@ln -s /Applications $(DIST_DIR)/dmg_temp/Applications
	@if [ -f assets/background.png ]; then cp assets/background.png $(DIST_DIR)/dmg_temp/.background/background.png; fi
	@codesign --deep --force --sign - --entitlements assets/entitlements.plist $(DIST_DIR)/dmg_temp/$(APP_NAME)
	@hdiutil create -volname "CalendarAssistant v$(VERSION)" -srcfolder $(DIST_DIR)/dmg_temp -ov -format UDZO $(DIST_DIR)/$(DMG_NAME)
	@rm -rf $(DIST_DIR)/dmg_temp
	@echo "✅ DMG created: $(DIST_DIR)/$(DMG_NAME)"

# 构建 Windows 版
build-win: setup
	@echo "🪟 Building Windows binary v$(VERSION)..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO_BIN) build -ldflags="-H=windowsgui" -o $(DIST_DIR)/$(WIN_EXE) cmd/calendar-assistant/main.go
	@echo "✅ Windows Release ready: $(DIST_DIR)/$(WIN_EXE)"

# 生成最终发布文件夹
release: dmg build-win
	@echo "🧹 Cleaning up intermediate files..."
	@rm -f $(DIST_DIR)/$(BINARY_NAME)-arm64 $(DIST_DIR)/$(BINARY_NAME)-amd64 $(DIST_DIR)/$(BINARY_NAME)-universal
	@echo "✨ Release v$(VERSION) files are ready in the '$(DIST_DIR)' folder:"
	@ls -lh $(DIST_DIR)/$(DMG_NAME) $(DIST_DIR)/$(WIN_EXE)

install: bundle
	@cp -R $(DIST_DIR)/$(APP_NAME) /Applications/
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > ~/Library/LaunchAgents/$(PLIST_NAME)
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> ~/Library/LaunchAgents/$(PLIST_NAME)
	@echo '<plist version="1.0"><dict>' >> ~/Library/LaunchAgents/$(PLIST_NAME)
	@echo '    <key>Label</key><string>com.yideng.calendar-assistant</string>' >> ~/Library/LaunchAgents/$(PLIST_NAME)
	@echo '    <key>ProgramArguments</key><array><string>/Applications/$(APP_NAME)/Contents/MacOS/$(BINARY_NAME)</string></array>' >> ~/Library/LaunchAgents/$(PLIST_NAME)
	@echo '    <key>RunAtLoad</key><true/>' >> ~/Library/LaunchAgents/$(PLIST_NAME)
	@echo '    <key>KeepAlive</key><false/>' >> ~/Library/LaunchAgents/$(PLIST_NAME)
	@echo '</dict></plist>' >> ~/Library/LaunchAgents/$(PLIST_NAME)
	@launchctl load ~/Library/LaunchAgents/$(PLIST_NAME) 2>/dev/null || echo "Already loaded"

clean:
	@rm -rf $(DIST_DIR)
