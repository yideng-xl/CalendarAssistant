package parser

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSmartParser(t *testing.T) {
	p := NewSmartParser()

	t.Run("Standard Date Format", func(t *testing.T) {
		text := "2026-04-02 10:00 PPT分享经验"
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, 2026, event.StartTime.Year())
		assert.Equal(t, 4, int(event.StartTime.Month()))
		assert.Equal(t, 2, event.StartTime.Day())
		assert.Equal(t, 10, event.StartTime.Hour())
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
		// 普通包含时间的文本（但不含日程特征）不应解析
		text := "今天是 2026-04-01 10:00"
		event, err := p.Parse(text)
		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "no calendar keywords")
	})

	t.Run("Meeting Link Inclusion", func(t *testing.T) {
		// 即使包含时间且不含关键词，但包含链接，也应通过
		text := "2026-04-01 10:00 https://meeting.tencent.com/dm/abc"
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.NotNil(t, event)
	})
}
