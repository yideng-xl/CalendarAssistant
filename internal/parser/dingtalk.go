package parser

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type dingTalkMeetingParser struct {
	subjectRegex   *regexp.Regexp
	startTimeRegex *regexp.Regexp
	linkRegex      *regexp.Regexp
	meetingIDRegex *regexp.Regexp
}

func NewDingTalkMeetingParser() MeetingParser {
	return &dingTalkMeetingParser{
		subjectRegex:   regexp.MustCompile(`主题[：:](.+)`),
		startTimeRegex: regexp.MustCompile(`开始时间[：:](.+)`),
		linkRegex:      regexp.MustCompile(`https://meeting\.dingtalk\.com/j/\w+`),
		meetingIDRegex: regexp.MustCompile(`会议号[：:](.+)`),
	}
}

func (p *dingTalkMeetingParser) CanParse(text string) bool {
	return strings.Contains(text, "钉钉会议")
}

func (p *dingTalkMeetingParser) Parse(text string) (*MeetingEvent, error) {
	subjectMatch := p.subjectRegex.FindStringSubmatch(text)
	startTimeMatch := p.startTimeRegex.FindStringSubmatch(text)
	linkMatch := p.linkRegex.FindString(text)
	meetingIDMatch := p.meetingIDRegex.FindStringSubmatch(text)

	if len(subjectMatch) < 2 || len(startTimeMatch) < 2 {
		return nil, errors.New("could not parse dingtalk meeting details")
	}

	layout := "2006-01-02 15:04"
	startTimeStr := strings.TrimSpace(startTimeMatch[1])
	startTime, err := time.ParseInLocation(layout, startTimeStr, time.Local)
	if err != nil {
		return nil, err
	}

	// 如果没有明确结束时间，默认为 1 小时
	endTime := startTime.Add(time.Hour)

	description := ""
	if len(meetingIDMatch) >= 2 {
		description = "会议号：" + strings.TrimSpace(meetingIDMatch[1])
	}

	return &MeetingEvent{
		Subject:     strings.TrimSpace(subjectMatch[1]),
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    linkMatch,
		Description: description,
	}, nil
}
