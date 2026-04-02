# CalendarAssistant-Go 项目交接手册 (2026-03-31)

## 1. 项目概览
- **定位**：系统级剪贴板日程同步助手。
- **技术栈**：Go 1.21+, AppleScript (macOS 同步), Windows WinRT (待补全), Systray (UI)。
- **核心价值**：解决 JG 公司多平台会议同步难题，实现无感、零依赖、高性能同步。

## 2. 核心架构
- **parser**: TDD 驱动的多引擎正则解析（支持腾讯会议、钉钉会议及其变体）。
- **sync**: 跨平台适配层。macOS 采用 AppleScript 以确保权限弹窗成功触发。
- **monitor**: 具备 MD5 去重能力的 2s 轮询监听器。
- **ui**: 基于 `systray` 的托盘菜单，包含 Nano Banana 状态图标嵌入。

## 3. 构建与分发
在 `CalendarAssistant-Go` 目录下运行：
- `make build-mac`: 编译纯二进制。
- `make bundle`: 封装为标准的 macOS `.app` (带图标、隐藏 Dock、权限声明)。
- `make build-win`: 交叉编译 Windows 版。
- `make all`: 一键构建全平台。

## 4. 常见问题 (FAQ)
- **图标不刷新**：如果修改了图标但 `.app` 还是空白，将 App 拖入 `/Applications` 文件夹即可强制系统刷新。
- **同步无反应**：确认是否点击了菜单切换到 `🟢 监听中`。
- **权限问题**：首次运行复制内容时，必须在系统弹窗中点击“允许”。若未弹出，运行 `tccutil reset Calendar`。

## 5. 后续扩展建议 (Roadmap)
- **开机自启动**：在 Makefile 中增加 `install` 目标，写入 LaunchAgents。
- **格式库**：增加飞书、企业微信的正则支持。
- **Windows 实装**：在 Windows 环境下使用 `go-ole` 实现完整的日历写入逻辑。
