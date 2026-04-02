package parser

import (
	"testing"
	"time"
)

func TestTencentMeetingParser(t *testing.T) {
	parser := NewTencentMeetingParser()
	
	t.Run("Standard Format", func(t *testing.T) {
		text := `会议主题：[项目] 需求评审
时间：2026-04-01 10:00-11:30
腾讯会议：https://meeting.tencent.com/dm/abc123xyz
会议号：#腾讯会议：123-456-789`
		
		event, err := parser.Parse(text)
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
}
