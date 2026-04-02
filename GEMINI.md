# CalendarAssistant 项目指令上下文 (Instructional Context)

## 1. 项目概览 (Project Overview)
- **定位**：系统级剪贴板助手（Tray App），旨在解决协同工具碎片化（腾讯会议、钉钉等）导致的日程同步难题。
- **当前状态**：Python MVP 版本。通过监听剪贴板，利用正则解析会议信息，并调用 AppleScript 同步至 macOS 日历。
- **技术路线**：从 Python MVP 向 **Go 语言重构** 演进，追求零依赖、高性能及跨平台支持。

## 2. 核心架构与逻辑 (Architecture & Logic)
- **监听器**：无间断读取系统剪贴板（当前 macOS 使用 `pbpaste`）。
- **解析引擎**：采用高度容错的正则表达式库，支持 YYYY-MM-DD、中英文冒号、腾讯会议链接等格式。
- **校验逻辑**：
    - **重复检测**：防止同一会议多次同步。
    - **冲突提醒**：检测时间交叉冲突，并提供通知。
- **通知系统**：利用 `osascript` 发送 macOS 原生弹窗。
- **自动化提醒**：同步时自动设置 -15m, -5m, 0m 三级声音闹钟。

## 3. 构建与运行 (Building and Running)
### 当前 Python MVP
- **主程序**：`CalendarAssistant/calendar_sync_prototype.py`
- **直接运行**：`python3 CalendarAssistant/calendar_sync_prototype.py`
- **打包 (PyInstaller)**：使用 `CalendarAssistant/CalendarAssistant.spec` 进行打包。
    - 命令：`pyinstaller CalendarAssistant/CalendarAssistant.spec`
- **产物**：`dist/CalendarAssistant`

### 未来 Go 版本 (TODO)
- **构建命令**：待重构启动后定义。
- **测试命令**：待重构启动后定义。

## 4. 开发约定 (Development Conventions)
- **中文优先**：所有交流、文档、代码注释及反馈必须使用 **中文**。
- **时间标准**：默认采用 **北京时间 (CST/UTC+8)**。
- **TDD (测试驱动开发)**：在编写实现代码前，必须先编写并验证测试用例。
- **设计优先**：在进行重大重构或功能开发前，需激活 `brainstorming` 技能进行方案确认。
- **工具隔离**：每个工具应位于其独立的文件夹下（如当前的 `CalendarAssistant/`）。

## 5. 关键文件索引 (Key Files)
- `AGENT.md`：核心配置与偏好设定。
- `PROJECT_HANDOFF.md`：项目现状与历史交接信息。
- `CalendarAssistant/calendar_sync_prototype.py`：当前核心逻辑实现。
- `CalendarAssistant/CalendarAssistant.spec`：PyInstaller 打包配置文件。
