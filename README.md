<div align="center">

# 💎 CalendarAssistant

**Intelligent Calendar Synchronization Tool for macOS & Windows**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Platform: macOS / Windows](https://img.shields.io/badge/platform-macOS%20%7C%20Windows-blue.svg)](https://github.com/yideng-xl/CalendarAssistant)
[![Version: 0.1.0](https://img.shields.io/badge/version-0.1.0-green.svg)](https://github.com/yideng-xl/CalendarAssistant/releases)

[🇬🇧 English](#english) | [🇨🇳 中文](#中文)

---

</div>

## 中文

**CalendarAssistant** 是一款专为解决协同工具碎片化而设计的跨平台状态栏应用。它能自动监听剪贴板，通过自研的智能解析引擎提取会议邀请信息，并一键同步至系统日历。

### ✨ 核心特性

- 🧠 **智能解析 (Smart Parser)**：支持自然语言识别（如“明天上午10点”、“周五下午2:30”），自动提取标题和会议链接。
- 🔗 **系统级集成**：利用 AppleScript 深度集成 macOS 日历，支持重复检测和时间冲突提醒。
- 🔔 **多级闹钟**：同步时自动设置 15分钟前、5分钟前、开始时三级声音闹钟，确保不遗漏任何会议。
- 🛡️ **智能过滤**：自动识别并排除日志数据（如 ERROR/INFO 堆栈），减少干扰。
- 🍎 **原生体验**：提供原生状态栏图标、macOS 原生通知，支持 Apple Silicon (M1/M2/M3) 和 Intel 芯片。

### 🚀 安装与使用

1.  在 [Releases](https://github.com/yideng-xl/CalendarAssistant/releases) 页面下载最新的 `CalendarAssistant-v0.1.0-mac.dmg`。
2.  双击打开并将应用拖入 `Applications` 文件夹。
3.  运行应用，它将静默运行在系统状态栏。

### 🛠 开发者指南

项目基于 Go 语言开发。使用以下命令构建：

```bash
# 生成 macOS Universal DMG 和 Windows EXE
make release
```

---

<a name="english"></a>
## English

**CalendarAssistant** is a cross-platform tray application designed to solve the fragmentation of collaboration tools. It monitors the clipboard, uses a custom smart parsing engine to extract meeting invitation details, and synchronizes them to the system calendar.

### ✨ Key Features

- 🧠 **Smart Parser**: Supports natural language processing (e.g., "Tomorrow 10 AM", "Friday 2:30 PM") and automatically extracts subjects and meeting links.
- 🔗 **System Integration**: Deep integration with macOS Calendar via AppleScript, including duplicate detection and conflict alerts.
- 🔔 **Multi-level Alarms**: Automatically sets sound alarms at -15m, -5m, and start time.
- 🛡️ **Log Filtering**: Intelligently identifies and excludes log data (e.g., ERROR/INFO stacks) to minimize false positives.
- 🍎 **Native Experience**: Native menu bar icon and system notifications. Universal Binary support for both Apple Silicon and Intel Macs.

### 🚀 Installation

1.  Download the latest `CalendarAssistant-v0.1.0-mac.dmg` from the [Releases](https://github.com/yideng-xl/CalendarAssistant/releases) page.
2.  Open the DMG and drag the application to the `Applications` folder.
3.  Launch the app; it runs silently in the system tray.

### 🛠 Development

Developed with Go. Build the project using:

```bash
# Generate macOS Universal DMG and Windows EXE
make release
```

---

## 👥 Contributors

*   **一簦 (The Developer)** - Creator and maintainer.
*   **Gemini CLI (AI Co-architect)** - AI pair programmer responsible for the smart parsing engine and system integration logic.

## 📄 License

MIT License.
