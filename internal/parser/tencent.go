package parser

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type tencentMeetingParser struct {
	subjectRegex   *regexp.Regexp
	timeRegex      *regexp.Regexp
	linkRegex      *regexp.Regexp
	meetingIDRegex *regexp.Regexp
}

func NewTencentMeetingParser() MeetingParser {
	return &tencentMeetingParser{
		subjectRegex:   regexp.MustCompile(`(?:会议主题|主题)[：:](.+)`),
		timeRegex:      regexp.MustCompile(`(?:会议时间|时间)[：:]\s*(\d{4}[-/]\d{2}[-/]\d{2})\s+(\d{2}:\d{2})-(\d{2}:\d{2})`),
		linkRegex:      regexp.MustCompile(`https://meeting\.tencent\.com/dm/\w+`),
		meetingIDRegex: regexp.MustCompile(`(?:腾讯会议号|会议号|会议 ID|Meeting ID)[：:]\s*([\d][\d\s-]+[\d])`),
	}
}

func (p *tencentMeetingParser) CanParse(text string) bool {
	return strings.Contains(text, "腾讯会议") || strings.Contains(text, "meeting.tencent.com")
}

func (p *tencentMeetingParser) Parse(text string) (*MeetingEvent, error) {
	subjectMatch := p.subjectRegex.FindStringSubmatch(text)
	timeMatch := p.timeRegex.FindStringSubmatch(text)
	linkMatch := p.linkRegex.FindString(text)

	// 标准格式：有主题行 + 标准时间行
	if len(subjectMatch) >= 2 && len(timeMatch) >= 4 {
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
			Subject:   strings.TrimSpace(subjectMatch[1]),
			StartTime: startTime,
			EndTime:   endTime,
			Location:  linkMatch,
		}, nil
	}

	// 口语化格式：没有标准模板，但包含"腾讯会议"关键词 + 会议号
	// 例如："四点半有个评审，腾讯会议号：925-9449-7964"
	meetingIDMatch := p.meetingIDRegex.FindStringSubmatch(text)
	if len(meetingIDMatch) < 2 && linkMatch == "" {
		return nil, errors.New("could not parse tencent meeting details")
	}

	// 委托 SmartParser 来解析时间和主题
	smart := &smartParser{}
	startTime, err := smart.parseTime(text)
	if err != nil {
		return nil, errors.New("腾讯会议号已识别，但无法解析时间: " + err.Error())
	}
	endTime := startTime.Add(time.Hour)

	// 尝试提取结束时间
	if et, ok := smart.parseEndTime(text, startTime); ok {
		endTime = et
	}

	subject := smart.extractSubject(text)
	if subject == "新建日程" {
		subject = "腾讯会议"
	}

	// 构建描述
	description := ""
	if len(meetingIDMatch) >= 2 {
		description = "腾讯会议号：" + strings.TrimSpace(meetingIDMatch[1])
	}

	return &MeetingEvent{
		Subject:     subject,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    linkMatch,
		Description: description,
	}, nil
}
