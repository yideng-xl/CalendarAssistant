package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type zoomMeetingParser struct {
	topicRegex     *regexp.Regexp
	timeRegex      *regexp.Regexp
	durationRegex  *regexp.Regexp
	linkRegex      *regexp.Regexp
	meetingIDRegex *regexp.Regexp
}

func NewZoomMeetingParser() MeetingParser {
	return &zoomMeetingParser{
		// Zoom 邀请常见格式（中英文混合）
		// Topic: xxx / 主题：xxx
		topicRegex: regexp.MustCompile(`(?:Topic|主题)[：:]\s*(.+)`),
		// Time: Apr 10, 2026 02:00 PM Beijing / 时间：2026年4月10日 14:00
		timeRegex: regexp.MustCompile(`(?:Time|时间)[：:]\s*(.+?)(?:\n|$)`),
		// Duration: 1 hr 30 min
		durationRegex: regexp.MustCompile(`(?:Duration|时长)[：:]\s*(.+?)(?:\n|$)`),
		// Zoom 会议链接
		linkRegex: regexp.MustCompile(`https://[a-z0-9]*\.?zoom\.us/j/\S+`),
		// Meeting ID: 123 4567 8901
		meetingIDRegex: regexp.MustCompile(`(?:Meeting\s*ID|会议\s*ID|会议号)[：:]\s*(\S[\d\s]+\d)`),
	}
}

func (p *zoomMeetingParser) CanParse(text string) bool {
	return strings.Contains(text, "zoom.us") ||
		strings.Contains(text, "Zoom Meeting") ||
		strings.Contains(text, "Zoom 会议") ||
		(strings.Contains(text, "Zoom") && strings.Contains(text, "Meeting ID"))
}

func (p *zoomMeetingParser) Parse(text string) (*MeetingEvent, error) {
	topicMatch := p.topicRegex.FindStringSubmatch(text)
	linkMatch := p.linkRegex.FindString(text)
	meetingIDMatch := p.meetingIDRegex.FindStringSubmatch(text)

	subject := ""
	if len(topicMatch) >= 2 {
		subject = strings.TrimSpace(topicMatch[1])
	}
	if subject == "" {
		return nil, errors.New("could not parse zoom meeting topic")
	}

	// 解析时间
	startTime, err := p.parseZoomTime(text)
	if err != nil {
		return nil, err
	}

	// 解析时长
	duration := time.Hour // 默认 1 小时
	if d := p.parseDuration(text); d > 0 {
		duration = d
	}
	endTime := startTime.Add(duration)

	description := ""
	if len(meetingIDMatch) >= 2 {
		description = "Meeting ID: " + strings.TrimSpace(meetingIDMatch[1])
	}

	return &MeetingEvent{
		Subject:     subject,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    linkMatch,
		Description: description,
	}, nil
}

// parseZoomTime 解析 Zoom 邀请中的时间，支持多种格式
func (p *zoomMeetingParser) parseZoomTime(text string) (time.Time, error) {
	timeMatch := p.timeRegex.FindStringSubmatch(text)
	if len(timeMatch) < 2 {
		return time.Time{}, errors.New("could not find time in zoom invite")
	}
	timeStr := strings.TrimSpace(timeMatch[1])

	// 格式1: 中文标准格式 — 2026年4月10日 14:00 或 2026-04-10 14:00
	cnRegexes := []*regexp.Regexp{
		regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})日\s+(\d{1,2}):(\d{2})`),
		regexp.MustCompile(`(\d{4})[-/](\d{1,2})[-/](\d{1,2})\s+(\d{1,2}):(\d{2})`),
	}
	for _, re := range cnRegexes {
		if m := re.FindStringSubmatch(timeStr); m != nil {
			year, _ := strconv.Atoi(m[1])
			month, _ := strconv.Atoi(m[2])
			day, _ := strconv.Atoi(m[3])
			hour, _ := strconv.Atoi(m[4])
			min, _ := strconv.Atoi(m[5])
			return time.Date(year, time.Month(month), day, hour, min, 0, 0, time.Local), nil
		}
	}

	// 格式2: 英文格式 — Apr 10, 2026 02:00 PM 或 April 10, 2026 2:00 PM
	enLayouts := []string{
		"Jan 2, 2006 3:04 PM",
		"Jan 2, 2006 03:04 PM",
		"January 2, 2006 3:04 PM",
		"January 2, 2006 03:04 PM",
		"Jan 02, 2006 3:04 PM",
		"Jan 02, 2006 03:04 PM",
	}
	// 清除时区标签文本（如 "Beijing", "Shanghai", "Pacific Time (US and Canada)" 等）
	cleanedTime := regexp.MustCompile(`\s+(Beijing|Shanghai|UTC|GMT|PST|EST|CST|PT|ET|CT|MT|Pacific|Eastern|Central|Mountain).*$`).ReplaceAllString(timeStr, "")
	cleanedTime = strings.TrimSpace(cleanedTime)

	for _, layout := range enLayouts {
		if t, err := time.ParseInLocation(layout, cleanedTime, time.Local); err == nil {
			// 修正年份（Go 的 time.Parse 可能默认为 0000）
			if t.Year() == 0 {
				now := time.Now()
				t = time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
			}
			return t, nil
		}
	}

	return time.Time{}, errors.New("could not parse zoom meeting time: " + timeStr)
}

// parseDuration 解析时长字段
func (p *zoomMeetingParser) parseDuration(text string) time.Duration {
	durationMatch := p.durationRegex.FindStringSubmatch(text)
	if len(durationMatch) < 2 {
		return 0
	}
	dStr := strings.TrimSpace(durationMatch[1])

	var total time.Duration

	// 匹配小时
	hrRegex := regexp.MustCompile(`(\d+)\s*(?:hr|hour|小时)`)
	if m := hrRegex.FindStringSubmatch(dStr); m != nil {
		h, _ := strconv.Atoi(m[1])
		total += time.Duration(h) * time.Hour
	}

	// 匹配分钟
	minRegex := regexp.MustCompile(`(\d+)\s*(?:min|minute|分钟|分)`)
	if m := minRegex.FindStringSubmatch(dStr); m != nil {
		mi, _ := strconv.Atoi(m[1])
		total += time.Duration(mi) * time.Minute
	}

	return total
}
