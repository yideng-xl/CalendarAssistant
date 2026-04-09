package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFeishuMeetingParser(t *testing.T) {
	p := NewFeishuMeetingParser()

	t.Run("Standard Feishu Invite", func(t *testing.T) {
		text := `张三 邀请你加入飞书会议
主题：产品需求评审会
时间：2026-04-10 14:00 - 15:30

飞书会议：https://vc.feishu.cn/j/abc123def
会议号：123-456-789

入会方式：点击链接入会或搜索会议号加入`

		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, "产品需求评审会", event.Subject)
		assert.Equal(t, time.Date(2026, 4, 10, 14, 0, 0, 0, time.Local), event.StartTime)
		assert.Equal(t, time.Date(2026, 4, 10, 15, 30, 0, 0, time.Local), event.EndTime)
		assert.Equal(t, "https://vc.feishu.cn/j/abc123def", event.Location)
		assert.Contains(t, event.Description, "123-456-789")
	})

	t.Run("Feishu Without End Time", func(t *testing.T) {
		text := `飞书会议邀请
主题：周一晨会
时间：2026-04-13 09:00

https://vc.feishu.cn/j/xyz789`

		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, "周一晨会", event.Subject)
		assert.Equal(t, time.Date(2026, 4, 13, 9, 0, 0, 0, time.Local), event.StartTime)
		// 无结束时间，默认 1 小时
		assert.Equal(t, time.Date(2026, 4, 13, 10, 0, 0, 0, time.Local), event.EndTime)
	})

	t.Run("Lark Suite Format", func(t *testing.T) {
		text := `Join Lark Meeting
话题：Q2 Planning Review
时间：2026/04/15 10:00 - 11:00

https://vc.larksuite.com/j/meeting123
Meeting ID: 987-654-321`

		assert.True(t, p.CanParse(text))
		event, err := p.Parse(text)
		assert.NoError(t, err)
		assert.Equal(t, "Q2 Planning Review", event.Subject)
		assert.Equal(t, time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local), event.StartTime)
		assert.Equal(t, time.Date(2026, 4, 15, 11, 0, 0, 0, time.Local), event.EndTime)
	})

	t.Run("CanParse Rejects Non-Feishu", func(t *testing.T) {
		assert.False(t, p.CanParse("腾讯会议邀请"))
		assert.False(t, p.CanParse("这是一段普通文本"))
	})
}
