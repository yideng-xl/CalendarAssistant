package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSmartParser(t *testing.T) {
	p := NewSmartParser()

	t.Run("Standard Date Format", func(t *testing.T) {
		text := "2026-04-02 10:00 技术分享会"
		assert.True(t, p.CanParse(text), "应能识别包含会议关键词和时间的文本")
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 2026, event.StartTime.Year())
		assert.Equal(t, 4, int(event.StartTime.Month()))
		assert.Equal(t, 2, event.StartTime.Day())
		assert.Equal(t, 10, event.StartTime.Hour())
	})

	t.Run("CanParse Rejects Plain Text", func(t *testing.T) {
		// 普通文本不含会议关键词和时间，CanParse 应返回 false
		assert.False(t, p.CanParse("这是一段普通的文字"))
		assert.False(t, p.CanParse("今天天气不错"))
		assert.False(t, p.CanParse("需求文档已经更新"))
		assert.False(t, p.CanParse("讨论一下方案"))
	})

	t.Run("CanParse Accepts Meeting Text", func(t *testing.T) {
		assert.True(t, p.CanParse("明天 10:00 周会"))
		assert.True(t, p.CanParse("2026-04-05 14:00 评审会议"))
	})

	t.Run("Log Data Filter", func(t *testing.T) {
		// 日志文本不应解析为日程
		logText := "2026-04-01 10:00:01 INFO [main] c.y.c.Monitor: Clipboard updated"
		event, err := p.Parse(logText)
		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "log data")
	})

	t.Run("Normal Text Filter", func(t *testing.T) {
		// 有时间但无会议关键词，CanParse 应直接返回 false
		text := "今天是 2026-04-01 10:00"
		assert.False(t, p.CanParse(text), "仅含时间不含会议关键词，不应触发解析")
	})

	t.Run("Meeting Link Inclusion", func(t *testing.T) {
		// 包含会议链接+关键词+时间，应能通过
		text := "2026-04-01 10:00 会议邀请 https://meeting.tencent.com/dm/abc"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.NotNil(t, event)
	})

	// === 新增：更多中文日期格式测试 ===

	t.Run("Chinese Date - Year Month Day", func(t *testing.T) {
		text := "2026年4月15日 14:00 产品评审会议"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 2026, event.StartTime.Year())
		assert.Equal(t, 4, int(event.StartTime.Month()))
		assert.Equal(t, 15, event.StartTime.Day())
		assert.Equal(t, 14, event.StartTime.Hour())
	})

	t.Run("Chinese Date - Short Month Day Hao", func(t *testing.T) {
		text := "4月10号 下午2:30 周会"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 4, int(event.StartTime.Month()))
		assert.Equal(t, 10, event.StartTime.Day())
		assert.Equal(t, 14, event.StartTime.Hour())
		assert.Equal(t, 30, event.StartTime.Minute())
	})

	t.Run("Chinese Date - Day After Tomorrow", func(t *testing.T) {
		text := "大后天 上午10:00 面试"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		expected := time.Now().AddDate(0, 0, 3)
		assert.Equal(t, expected.Day(), event.StartTime.Day())
		assert.Equal(t, 10, event.StartTime.Hour())
	})

	t.Run("Chinese Date - Next Week", func(t *testing.T) {
		text := "下周三 14:00 评审会议"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, time.Wednesday, event.StartTime.Weekday())
		// 下周三应该至少7天后
		assert.True(t, event.StartTime.After(time.Now().AddDate(0, 0, 6)))
	})

	t.Run("Chinese Date - Evening", func(t *testing.T) {
		text := "明天 晚上8:00 分享会"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 20, event.StartTime.Hour())
	})

	// === 新增：时间范围解析 ===

	t.Run("Time Range Detection", func(t *testing.T) {
		text := "2026-04-10 10:00-11:30 评审会议"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 10, event.StartTime.Hour())
		assert.Equal(t, 11, event.EndTime.Hour())
		assert.Equal(t, 30, event.EndTime.Minute())
	})

	// === 新增：时区处理测试 ===

	t.Run("Timezone - UTC", func(t *testing.T) {
		text := "2026-04-10 02:00 UTC 会议邀请"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		// UTC 02:00 转换为本地时间（北京时间 UTC+8 = 10:00）
		_, localOffset := time.Now().In(time.Local).Zone()
		expectedHour := 2 + localOffset/3600
		assert.Equal(t, expectedHour, event.StartTime.Hour())
	})

	t.Run("Timezone - UTC+Offset", func(t *testing.T) {
		text := "2026-04-10 09:00 UTC+9 会议邀请"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		// UTC+9 09:00 = UTC 00:00 → 转换为本地时间
		_, localOffset := time.Now().In(time.Local).Zone()
		expectedHour := (9 - 9 + localOffset/3600 + 24) % 24
		assert.Equal(t, expectedHour, event.StartTime.Hour())
	})
}

// === 新增：中文数字时间测试 ===

func TestSmartParserChineseNumberTime(t *testing.T) {
	p := NewSmartParser()

	t.Run("Chinese Number - Si Dian Ban", func(t *testing.T) {
		text := "四点半有个评审会议"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 16, event.StartTime.Hour()) // 四点半 → 16:30（无前缀默认下午）
		assert.Equal(t, 30, event.StartTime.Minute())
	})

	t.Run("Chinese Number - Morning", func(t *testing.T) {
		text := "上午九点 周会"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 9, event.StartTime.Hour())
		assert.Equal(t, 0, event.StartTime.Minute())
	})

	t.Run("Chinese Number - Afternoon Explicit", func(t *testing.T) {
		text := "下午两点四十五 评审会议"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 14, event.StartTime.Hour())
		assert.Equal(t, 45, event.StartTime.Minute())
	})

	t.Run("Chinese Number - Ten O Clock", func(t *testing.T) {
		text := "十点 站会"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 10, event.StartTime.Hour())
	})

	t.Run("Chinese Number With Relative Date", func(t *testing.T) {
		text := "明天三点半 评审会议"
		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		tomorrow := time.Now().AddDate(0, 0, 1)
		assert.Equal(t, tomorrow.Day(), event.StartTime.Day())
		assert.Equal(t, 15, event.StartTime.Hour()) // 三点半 → 15:30
		assert.Equal(t, 30, event.StartTime.Minute())
	})
}

// === 新增：多事件解析测试 ===

func TestSmartParserMultiple(t *testing.T) {
	p := NewSmartParser().(*smartParser)

	t.Run("Numbered Events", func(t *testing.T) {
		text := `今天会议安排：
1. 2026-04-10 10:00 产品评审会议
2. 2026-04-10 14:00 技术分享会
3. 2026-04-10 16:00 周会`

		events, err := p.ParseMultiple(text)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(events))
		assert.Equal(t, 10, events[0].StartTime.Hour())
		assert.Equal(t, 14, events[1].StartTime.Hour())
		assert.Equal(t, 16, events[2].StartTime.Hour())
	})

	t.Run("Single Event Fallback", func(t *testing.T) {
		text := "2026-04-10 10:00 评审会议"
		events, err := p.ParseMultiple(text)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
	})
}
