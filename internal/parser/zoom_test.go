package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZoomMeetingParser(t *testing.T) {
	p := NewZoomMeetingParser()

	t.Run("English Zoom Invite", func(t *testing.T) {
		text := `Join Zoom Meeting
Topic: Weekly Team Standup
Time: Apr 10, 2026 10:00 AM Beijing

Duration: 1 hr 30 min

https://us02web.zoom.us/j/12345678901?pwd=abc123
Meeting ID: 123 4567 8901
Passcode: 654321`

		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, "Weekly Team Standup", event.Subject)
		assert.Equal(t, time.Date(2026, 4, 10, 10, 0, 0, 0, time.Local), event.StartTime)
		// 1hr 30min
		assert.Equal(t, time.Date(2026, 4, 10, 11, 30, 0, 0, time.Local), event.EndTime)
		assert.Contains(t, event.Location, "zoom.us")
		assert.Contains(t, event.Description, "123 4567 8901")
	})

	t.Run("Chinese Zoom Invite", func(t *testing.T) {
		text := `加入 Zoom 会议
主题：产品月度总结
时间：2026-04-15 14:00
时长：2 小时

https://zoom.us/j/98765432101
会议号：987 6543 2101`

		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, "产品月度总结", event.Subject)
		assert.Equal(t, time.Date(2026, 4, 15, 14, 0, 0, 0, time.Local), event.StartTime)
		assert.Equal(t, time.Date(2026, 4, 15, 16, 0, 0, 0, time.Local), event.EndTime)
	})

	t.Run("Zoom PM Time Format", func(t *testing.T) {
		text := `Zoom Meeting Invitation
Topic: Design Review
Time: April 10, 2026 2:30 PM

https://us04web.zoom.us/j/77788899900`

		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, "Design Review", event.Subject)
		assert.Equal(t, time.Date(2026, 4, 10, 14, 30, 0, 0, time.Local), event.StartTime)
		// 无时长，默认 1 小时
		assert.Equal(t, time.Date(2026, 4, 10, 15, 30, 0, 0, time.Local), event.EndTime)
	})

	t.Run("CanParse Rejects Non-Zoom", func(t *testing.T) {
		assert.False(t, p.CanParse("飞书会议邀请"))
		assert.False(t, p.CanParse("这是一段普通文字"))
	})
}
