package sync

import (
	"time"
	"github.com/yideng/calendar-assistant/internal/parser"
)

// SyncOptions 同步偏好设置
type SyncOptions struct {
	Reminders []time.Duration // 提醒档位（如：-15m, -5m, 0m）
	Silent    bool            // 是否静默通知
}

// CalendarProvider 定义跨平台日历操作接口
type CalendarProvider interface {
	// SyncEvent 同步会议到系统日历
	SyncEvent(event *parser.MeetingEvent, options SyncOptions) error
	// HasEvent 检测是否已存在该会议，防止重复同步
	HasEvent(event *parser.MeetingEvent) (bool, error)
	// GetConflicts 获取冲突的会议标题
	GetConflicts(event *parser.MeetingEvent) ([]string, error)
}
