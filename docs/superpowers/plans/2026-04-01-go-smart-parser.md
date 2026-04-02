# Go 版本智能解析器与 AppleScript 重构 执行计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 引入智能时间解析库并修复 AppleScript 核心缺陷。

**Architecture:** 实现一个新的 `SmartParser` 并作为通用兜底，同时更新 `calendar_darwin.go` 模板。

**Tech Stack:** Go, github.com/olebedev/when, AppleScript.

---

### Task 1: 环境准备

**Files:**
- Modify: `CalendarAssistant-Go/go.mod`

- [ ] **Step 1: 添加 `when` 库依赖**

Run: `go get github.com/olebedev/when`
Run: `go get github.com/olebedev/when/rules/zh` (中文支持)

- [ ] **Step 2: 确认依赖成功**

Run: `go mod tidy`
Expected: `go.sum` 已更新，包含 `when` 相关条目。

---

### Task 2: 实现 `SmartParser`

**Files:**
- Create: `CalendarAssistant-Go/internal/parser/smart.go`
- Create: `CalendarAssistant-Go/internal/parser/smart_test.go`

- [ ] **Step 1: 编写 `SmartParser` 核心逻辑**

```go
package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/zh"
	"github.com/olebedev/when/rules/en"
)

type smartParser struct {
	w *when.Parser
}

func NewSmartParser() MeetingParser {
	w := when.New(nil)
	w.Add(zh.All...)
	w.Add(en.All...)
	return &smartParser{w: w}
}

func (p *smartParser) CanParse(text string) bool {
	// 任何包含 URL 或 能识别出时间的文本都可以尝试
	return true
}

func (p *smartParser) Parse(text string) (*MeetingEvent, error) {
	// 1. 解析时间
	r, err := p.w.Parse(text, time.Now())
	if err != nil || r == nil {
		return nil, err
	}
	startTime := r.Time
	endTime := startTime.Add(1 * time.Hour)

	// 2. 提取链接
	linkRegex := regexp.MustCompile(`https?://[^\s]+`)
	link := linkRegex.FindString(text)

	// 3. 提取标题
	subject := p.extractSubject(text)

	return &MeetingEvent{
		Subject:   subject,
		StartTime: startTime,
		EndTime:   endTime,
		Location:  link,
	}, nil
}

func (p *smartParser) extractSubject(text string) string {
	lines := strings.Split(text, "\n")
	keywords := []string{"分享", "评审", "会议", "PPT", "讨论", "需求", "优化"}
	
	// 优先寻找关键词行
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" { continue }
		for _, kw := range keywords {
			if strings.Contains(trimmed, kw) {
				return trimmed
			}
		}
	}
	
	// 兜底：取第一行非空
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			if len(trimmed) > 50 { return trimmed[:50] + "..." }
			return trimmed
		}
	}
	
	return "新建日程"
}
```

- [ ] **Step 2: 编写测试用例**

```go
package parser

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSmartParser(t *testing.T) {
	p := NewSmartParser()
	text := "明天上午10点评审V1.19.4需求 https://kdocs.cn/l/123"
	event, err := p.Parse(text)
	assert.NoError(t, err)
	assert.Contains(t, event.Subject, "评审")
	assert.Equal(t, 10, event.StartTime.Hour())
}
```

- [ ] **Step 3: 运行测试验证**

Run: `go test -v ./internal/parser/...`
Expected: 测试通过。

---

### Task 3: 重构 `calendar_darwin.go`

**Files:**
- Modify: `CalendarAssistant-Go/internal/sync/calendar_darwin.go`

- [ ] **Step 1: 修正 AppleScript 逻辑**

```go
// 增加注入防护函数
func escapeAppleScriptString(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

// 修改 SyncEvent 中的模板，确保日期顺序：Year -> Month -> Day
// 并增加转义调用
```

- [ ] **Step 2: 修正提醒设置逻辑**

确保 `buildAlarmScript` 生成正确的 `sound alarm` 指令。

---

### Task 4: 更新主逻辑集成

**Files:**
- Modify: `CalendarAssistant-Go/internal/parser/parser.go`

- [ ] **Step 1: 更新 `GetParsers` 列表**

在 `GetParsers` (或同等函数) 中将 `SmartParser` 添加到末尾作为通用兜底。
