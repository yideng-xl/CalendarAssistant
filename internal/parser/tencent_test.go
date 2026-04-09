package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTencentMeetingParser(t *testing.T) {
	p := NewTencentMeetingParser()

	t.Run("Standard Format", func(t *testing.T) {
		text := `会议主题：[项目] 需求评审
时间：2026-04-01 10:00-11:30
腾讯会议：https://meeting.tencent.com/dm/abc123xyz
会议号：#腾讯会议：123-456-789`

		event, err := p.Parse(text)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if event.Subject != "[项目] 需求评审" {
			t.Errorf("Expected subject '[项目] 需求评审', got '%s'", event.Subject)
		}

		expectedStart := time.Date(2026, 4, 1, 10, 0, 0, 0, time.Local)
		if !event.StartTime.Equal(expectedStart) {
			t.Errorf("Expected start time %v, got %v", expectedStart, event.StartTime)
		}

		expectedEnd := time.Date(2026, 4, 1, 11, 30, 0, 0, time.Local)
		if !event.EndTime.Equal(expectedEnd) {
			t.Errorf("Expected end time %v, got %v", expectedEnd, event.EndTime)
		}

		if event.Location != "https://meeting.tencent.com/dm/abc123xyz" {
			t.Errorf("Expected location 'https://meeting.tencent.com/dm/abc123xyz', got '%s'", event.Location)
		}
	})

	t.Run("Informal Chat Style", func(t *testing.T) {
		// 口语化格式：没有标准模板，只有会议号和中文数字时间
		text := `四点半网管有个优化项需求的评审（已有功能基础上的优化），可以听下，腾讯会议号：925-9449-7964，资料可以先看看【金山文档 | WPS云文档】 V1.17.4-配置基线对比`

		assert.True(t, p.CanParse(text), "应能识别包含'腾讯会议'的口语化文本")
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.NotNil(t, event)
		// 四点半 → 16:30
		assert.Equal(t, 16, event.StartTime.Hour())
		assert.Equal(t, 30, event.StartTime.Minute())
		// 会议号应在描述中
		assert.Contains(t, event.Description, "925-9449-7964")
	})

	t.Run("Informal With Only Meeting ID", func(t *testing.T) {
		text := `明天下午三点 腾讯会议号：123-456-789`

		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		tomorrow := time.Now().AddDate(0, 0, 1)
		assert.Equal(t, tomorrow.Day(), event.StartTime.Day())
		assert.Equal(t, 15, event.StartTime.Hour())
		assert.Contains(t, event.Description, "123-456-789")
	})
}
