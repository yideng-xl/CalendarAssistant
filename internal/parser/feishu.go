package parser

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type feishuMeetingParser struct {
	subjectRegex   *regexp.Regexp
	timeRangeRegex *regexp.Regexp
	dateTimeRegex  *regexp.Regexp
	linkRegex      *regexp.Regexp
	meetingIDRegex *regexp.Regexp
}

func NewFeishuMeetingParser() MeetingParser {
	return &feishuMeetingParser{
		// 飞书邀请常见格式：
		// 主题：xxx / 会议主题：xxx
		subjectRegex: regexp.MustCompile(`(?:会议主题|主题|话题)[：:](.+)`),
		// 时间范围格式：2026-04-10 10:00 - 11:00 或 2026/04/10 10:00-11:00
		timeRangeRegex: regexp.MustCompile(`(\d{4}[-/]\d{1,2}[-/]\d{1,2})\s+(\d{1,2}:\d{2})\s*[-~]\s*(\d{1,2}:\d{2})`),
		// 单独的日期时间格式：2026-04-10 10:00
		dateTimeRegex: regexp.MustCompile(`(\d{4}[-/]\d{1,2}[-/]\d{1,2})\s+(\d{1,2}:\d{2})`),
		// 飞书会议链接
		linkRegex: regexp.MustCompile(`https://(?:vc\.feishu\.cn|meetings\.feishu\.cn|vc\.larksuite\.com)/j/\S+`),
		// 会议号：xxx-xxx-xxx 或 会议 ID：xxx
		meetingIDRegex: regexp.MustCompile(`(?:会议号|会议ID|Meeting\s*ID)[：:]\s*(\S+)`),
	}
}

func (p *feishuMeetingParser) CanParse(text string) bool {
	return strings.Contains(text, "飞书会议") ||
		strings.Contains(text, "feishu.cn") ||
		strings.Contains(text, "larksuite.com") ||
		strings.Contains(text, "Lark 会议") ||
		strings.Contains(text, "Lark Meeting")
}

func (p *feishuMeetingParser) Parse(text string) (*MeetingEvent, error) {
	subjectMatch := p.subjectRegex.FindStringSubmatch(text)
	linkMatch := p.linkRegex.FindString(text)
	meetingIDMatch := p.meetingIDRegex.FindStringSubmatch(text)

	// 主题：优先使用正则匹配，否则尝试从第一行非空行提取
	subject := ""
	if len(subjectMatch) >= 2 {
		subject = strings.TrimSpace(subjectMatch[1])
	} else {
		subject = p.extractSubjectFromLines(text)
	}
	if subject == "" {
		return nil, errors.New("could not parse feishu meeting subject")
	}

	// 解析时间
	layout := "2006-01-02 15:04"
	var startTime, endTime time.Time

	// 优先匹配时间范围格式
	if m := p.timeRangeRegex.FindStringSubmatch(text); m != nil {
		dateStr := strings.ReplaceAll(m[1], "/", "-")
		var err error
		startTime, err = time.ParseInLocation(layout, dateStr+" "+m[2], time.Local)
		if err != nil {
			return nil, err
		}
		endTime, err = time.ParseInLocation(layout, dateStr+" "+m[3], time.Local)
		if err != nil {
			return nil, err
		}
	} else if m := p.dateTimeRegex.FindStringSubmatch(text); m != nil {
		dateStr := strings.ReplaceAll(m[1], "/", "-")
		var err error
		startTime, err = time.ParseInLocation(layout, dateStr+" "+m[2], time.Local)
		if err != nil {
			return nil, err
		}
		endTime = startTime.Add(time.Hour) // 默认 1 小时
	} else {
		return nil, errors.New("could not parse feishu meeting time")
	}

	description := ""
	if len(meetingIDMatch) >= 2 {
		description = "会议号：" + strings.TrimSpace(meetingIDMatch[1])
	}

	return &MeetingEvent{
		Subject:     subject,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    linkMatch,
		Description: description,
	}, nil
}

// extractSubjectFromLines 从邀请文本中提取可能的主题行
func (p *feishuMeetingParser) extractSubjectFromLines(text string) string {
	lines := strings.Split(text, "\n")
	// 飞书邀请格式常为 "xxx 邀请你加入飞书会议" 或者第一行含有主题信息
	inviteRegex := regexp.MustCompile(`(.+?)邀请你(?:加入|参加)`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if m := inviteRegex.FindStringSubmatch(line); m != nil {
			// 返回邀请者信息作为辅助，实际主题在后面
			continue
		}
		// 跳过包含链接、会议号等信息行
		if strings.HasPrefix(line, "http") || strings.Contains(line, "会议号") ||
			strings.Contains(line, "Meeting ID") || strings.Contains(line, "入会") {
			continue
		}
		// 取第一个有意义的行作为主题
		runes := []rune(line)
		if len(runes) > 50 {
			return string(runes[:50]) + "..."
		}
		return line
	}
	return ""
}
