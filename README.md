<div align="center">

# 💎 CalendarAssistant

**Intelligent Calendar Synchronization Tool for macOS & Windows**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Platform: macOS / Windows](https://img.shields.io/badge/platform-macOS%20%7C%20Windows-blue.svg)](https://github.com/yideng-xl/CalendarAssistant)
[![Version: 0.1.0](https://img.shields.io/badge/version-0.5.1-green.svg)](https://github.com/yideng-xl/CalendarAssistant/releases)

[🇨🇳 中文说明 (Chinese)](README_zh.md)

---

</div>

**CalendarAssistant** is a cross-platform tray application designed to solve the fragmentation of collaboration tools. It monitors the clipboard, uses a custom smart parsing engine to extract meeting invitation details, and synchronizes them to the system calendar.

### ✨ Key Features

- 🧠 **Smart Parser**: Supports natural language processing (e.g., "Tomorrow 10 AM", "Friday 2:30 PM") and automatically extracts subjects and meeting links.
- 🔗 **System Integration**: Deep integration with macOS Calendar via AppleScript, including duplicate detection and conflict alerts.
- 🔔 **Multi-level Alarms**: Automatically sets sound alarms at -15m, -5m, and start time.
- 🛡️ **Log Filtering**: Intelligently identifies and excludes log data (e.g., ERROR/INFO stacks) to minimize false positives.
- 🍎 **Native Experience**: Native menu bar icon and system notifications. Universal Binary support for both Apple Silicon and Intel Macs.

### 🚀 Installation

1.  Download the latest `CalendarAssistant-v0.5.1-mac.dmg` from the [Releases](https://github.com/yideng-xl/CalendarAssistant/releases) page.
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

MIT License. See [LICENSE](LICENSE) for details.
