package parser

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type tencentMeetingParser struct {
	subjectRegex *regexp.Regexp
	timeRegex    *regexp.Regexp
	linkRegex    *regexp.Regexp
}

func NewTencentMeetingParser() MeetingParser {
	return &tencentMeetingParser{
		subjectRegex: regexp.MustCompile(`(?:会议主题|主题)[：:](.+)`),
		timeRegex:    regexp.MustCompile(`(?:会议时间|时间)[：:]\s*(\d{4}[-/]\d{2}[-/]\d{2})\s+(\d{2}:\d{2})-(\d{2}:\d{2})`),
		linkRegex:    regexp.MustCompile(`https://meeting\.tencent\.com/dm/\w+`),
	}
}

func (p *tencentMeetingParser) CanParse(text string) bool {
	return strings.Contains(text, "腾讯会议") || strings.Contains(text, "meeting.tencent.com")
}

func (p *tencentMeetingParser) Parse(text string) (*MeetingEvent, error) {
	subjectMatch := p.subjectRegex.FindStringSubmatch(text)
	timeMatch := p.timeRegex.FindStringSubmatch(text)
	linkMatch := p.linkRegex.FindString(text)

	if len(subjectMatch) < 2 || len(timeMatch) < 4 {
		return nil, errors.New("could not parse tencent meeting details")
	}

	dateStr := strings.ReplaceAll(timeMatch[1], "/", "-")
	startTimeStr := timeMatch[2]
	endTimeStr := timeMatch[3]

	layout := "2006-01-02 15:04"
	startTime, err := time.ParseInLocation(layout, dateStr+" "+startTimeStr, time.Local)
	if err != nil {
		return nil, err
	}

	endTime, err := time.ParseInLocation(layout, dateStr+" "+endTimeStr, time.Local)
	if err != nil {
		return nil, err
	}

	return &MeetingEvent{
		Subject:  strings.TrimSpace(subjectMatch[1]),
		StartTime: startTime,
		EndTime:   endTime,
		Location:  linkMatch,
	}, nil
}
