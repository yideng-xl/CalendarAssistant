package parser

import "time"

// MeetingEvent 代表解析出的会议事件
type MeetingEvent struct {
	Subject     string    // 会议主题
	StartTime   time.Time // 开始时间
	EndTime     time.Time // 结束时间
	Location    string    // 会议链接/地点
	Description string    // 备注信息（如会议号）
}

// MeetingParser 定义了不同会议格式的解析器接口
type MeetingParser interface {
	Parse(text string) (*MeetingEvent, error)
	CanParse(text string) bool
}
