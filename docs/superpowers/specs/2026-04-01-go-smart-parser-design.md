# 设计文档：Go 版本智能解析器与 AppleScript 重构 (2026-04-01)

## 1. 目标
通过引入自然语言处理库 `github.com/olebedev/when`，为 Go 版本 CalendarAssistant 提供智能的中文时间解析能力，并修复现有 AppleScript 的安全与逻辑缺陷。

## 2. 需求分析
- **智能时间提取**: 自动识别“明天上午10点”、“周五下午两点”等自然语言。
- **智能标题提取**: 通过关键词识别（分享、评审、会议等）或首行兜底来确定日程标题。
- **AppleScript 安全性**: 防止双引号注入导致脚本执行失败。
- **AppleScript 稳定性**: 调整日期设置顺序（年->月->日），防止月份天数溢出。

## 3. 系统架构 (Go)
### 3.1 核心解析器 (`internal/parser/smart.go`)
- **时间解析**: 引入 `github.com/olebedev/when` 及其中文插件。
- **逻辑流程**:
    1. 提取第一个 URL 作为 `Location`。
    2. 使用 `when.Parse` 提取第一个时间段。
    3. 默认 `EndTime = StartTime + 1h`。
    4. 标题识别：正则匹配关键词行 -> 首行非空内容 -> 兜底标题。
- **接口匹配**: 实现 `MeetingParser` 接口。

### 3.2 同步逻辑重构 (`internal/sync/calendar_darwin.go`)
- **注入防护**: 增加 `escapeAppleScriptString` 函数。
- **日期设置**: 调整 AppleScript 模板，顺序改为 `year -> month -> day -> hours -> minutes -> seconds`。
- **多级提醒**: 修正 `buildAlarmScript` 中硬编码的报警类型，支持 `-15, -5, 0` 分钟提醒。

## 4. 依赖项
- `github.com/olebedev/when`: MIT 协议，开源免费。

## 5. 验证计划 (TDD)
- **单元测试**: 在 `internal/parser/smart_test.go` 中测试多种中文自然语言。
- **功能测试**: 运行 `go run cmd/calendar-assistant/main.go` 并粘贴用户提供的测试文本。
