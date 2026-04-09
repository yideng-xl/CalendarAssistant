<div align="center">

# 💎 CalendarAssistant

**Intelligent Calendar Synchronization Tool for macOS & Windows**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform: macOS / Windows](https://img.shields.io/badge/platform-macOS%20%7C%20Windows-blue.svg)](https://github.com/yideng-xl/CalendarAssistant)
[![Version: 0.1.0](https://img.shields.io/badge/version-0.5.1-green.svg)](https://github.com/yideng-xl/CalendarAssistant/releases)

[🇬🇧 English Version](README.md)

---

</div>

**CalendarAssistant** 是一款专为解决协同工具碎片化而设计的跨平台状态栏应用。它能自动监听剪贴板，通过自研的智能解析引擎提取会议邀请信息，并一键同步至系统日历。

### ✨ 核心特性

- 🧠 **智能解析 (Smart Parser)**：支持自然语言识别（如“明天上午10点”、“周五下午2:30”），自动提取标题和会议链接。
- 🔗 **系统级集成**：利用 AppleScript 深度集成 macOS 日历，支持重复检测和时间冲突提醒。
- 🔔 **多级闹钟**：同步时自动设置 15分钟前、5分钟前、开始时三级声音闹钟，确保不遗漏任何会议。
- 🛡️ **智能过滤**：自动识别并排除日志数据（如 ERROR/INFO 堆栈），减少干扰。
- 🍎 **原生体验**：提供原生状态栏图标、macOS 原生通知，支持 Apple Silicon (M1/M2/M3) 和 Intel 芯片。

### 🚀 安装与使用

1.  在 [Releases](https://github.com/yideng-xl/CalendarAssistant/releases) 页面下载最新的 `CalendarAssistant-v0.5.1-mac.dmg`。
2.  双击打开并将应用拖入 `Applications` 文件夹。
3.  运行应用，它将静默运行在系统状态栏。

### 🛠 开发者指南

项目基于 Go 语言开发。使用以下命令构建：

```bash
# 生成 macOS Universal DMG 和 Windows EXE
make release
```

---

## 👥 贡献者

*   **一簦 (The Developer)** - 创建者与维护者。
*   **Gemini CLI (AI Co-architect)** - 负责智能解析引擎与系统集成逻辑。
    - GitHub Email: `gemini-cli@google.com`

## 📄 开源协议

本项目采用 [MIT License](LICENSE) 开源协议。
