package parser

import (
	"testing"
	"time"
)

func TestDingTalkMeetingParser(t *testing.T) {
	parser := NewDingTalkMeetingParser()
	
	t.Run("DingTalk Format 1", func(t *testing.T) {
		text := `许磊 邀你参加钉钉会议 
主题：许磊发起的视频会议_0331
开始时间：2026-03-31 13:41
会议号：463 185 3104

入会链接：
https://meeting.dingtalk.com/j/TyCaMl5LEgB
可通过浏览器入会，无需下载钉钉`
		
		event, err := parser.Parse(text)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if event.Subject != "许磊发起的视频会议_0331" {
			t.Errorf("Expected subject '许磊发起的视频会议_0331', got '%s'", event.Subject)
		}
		
		expectedStart := time.Date(2026, 3, 31, 13, 41, 0, 0, time.Local)
		if !event.StartTime.Equal(expectedStart) {
			t.Errorf("Expected start time %v, got %v", expectedStart, event.StartTime)
		}
		
		// 默认结束时间（如果文中没有明确结束时间，暂定开始后 1 小时）
		expectedEnd := expectedStart.Add(time.Hour)
		if !event.EndTime.Equal(expectedEnd) {
			t.Errorf("Expected end time %v, got %v", expectedEnd, event.EndTime)
		}
		
		if event.Location != "https://meeting.dingtalk.com/j/TyCaMl5LEgB" {
			t.Errorf("Expected location 'https://meeting.dingtalk.com/j/TyCaMl5LEgB', got '%s'", event.Location)
		}
	})
}
