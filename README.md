# CalendarAssistant (日历助手) 📅

[English](#english) | [中文](#中文)

---

## 中文

### 简介
**CalendarAssistant** 是一款专为解决协同工具碎片化而设计的 macOS 状态栏应用。它能自动监听剪贴板，智能解析来自腾讯会议、钉钉、Zoom 等工具的会议邀请，并一键同步至系统日历，同时自动设置多级声音闹钟提醒。

### 核心特性
- **智能解析 (Smart Parser)**：支持自然语言识别（如“明天上午10点”、“周五下午2:30”），自动提取标题和会议链接。
- **系统级集成**：利用 AppleScript 深度集成 macOS 日历，支持重复检测和时间冲突提醒。
- **多级闹钟**：同步时自动设置 15分钟前、5分钟前、开始时三级声音闹钟。
- **智能过滤**：自动识别并排除日志数据（如 ERROR/INFO 堆栈），减少干扰。
- **原生体验**：提供原生状态栏图标、macOS 原生通知。
- **双架构支持**：Universal Binary，同时支持 Apple Silicon (M1/M2/M3) 和 Intel 芯片。

### 安装与使用
1. 在 [Releases](https://github.com/yideng-xl/CalendarAssistant/releases) 页面下载最新的 `CalendarAssistant-Installer.dmg`。
2. 双击打开并将应用拖入 `Applications` 文件夹。
3. 运行应用，它将静默运行在系统状态栏。

### 开发指南
项目基于 Go 语言开发。使用以下命令构建：
```bash
# 生成 macOS Universal DMG 和 Windows EXE
make release
```

---

<a name="english"></a>
## English

### Introduction
**CalendarAssistant** is a macOS tray application designed to solve the fragmentation of collaboration tools. It monitors the clipboard, intelligently parses meeting invitations from tools like Tencent Meeting, DingTalk, and Zoom, and synchronizes them to the system calendar with automatic multi-level sound alarms.

### Key Features
- **Smart Parser**: Supports natural language processing (e.g., "Tomorrow 10 AM", "Friday 2:30 PM") and automatically extracts subjects and meeting links.
- **System Integration**: Deep integration with macOS Calendar via AppleScript, including duplicate detection and conflict alerts.
- **Multi-level Alarms**: Automatically sets sound alarms at -15m, -5m, and start time.
- **Log Filtering**: Intelligently identifies and excludes log data (e.g., ERROR/INFO stacks) to minimize false positives.
- **Native Experience**: Provides a native menu bar icon and system notifications.
- **Universal Binary**: Supports both Apple Silicon (M1/M2/M3) and Intel-based Macs.

### Installation
1. Download the latest `CalendarAssistant-Installer.dmg` from the [Releases](https://github.com/yideng-xl/CalendarAssistant/releases) page.
2. Open the DMG and drag the application to the `Applications` folder.
3. Launch the app; it runs silently in the system tray.

### Development
Developed with Go. Build the project using:
```bash
# Generate macOS Universal DMG and Windows EXE
make release
```

---

## License
MIT License.
