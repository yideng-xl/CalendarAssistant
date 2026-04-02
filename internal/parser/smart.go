package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type smartParser struct{}

func NewSmartParser() MeetingParser {
	return &smartParser{}
}

func (p *smartParser) CanParse(text string) bool {
	return strings.TrimSpace(text) != ""
}

func (p *smartParser) Parse(text string) (*MeetingEvent, error) {
	// 增加日志过滤逻辑：如果包含明显的日志特征，直接忽略
	if p.isLogData(text) {
		return nil, errors.New("detected as log data, skipping")
	}

	// 提取链接
	linkRegex := regexp.MustCompile(`https?://[^\s\n]+`)
	location := linkRegex.FindString(text)

	// 如果没有链接，也没有日程相关的核心词，则视为普通文本，忽略
	if location == "" && !p.hasCalendarKeywords(text) {
		return nil, errors.New("no calendar keywords or links found, skipping")
	}

	// 1. 提取时间
	startTime, err := p.parseTime(text)
	if err != nil {
		return nil, err
	}
	endTime := startTime.Add(time.Hour)

	// 3. 提取标题
	subject := p.extractSubject(text)

	return &MeetingEvent{
		Subject:   subject,
		StartTime: startTime,
		EndTime:   endTime,
		Location:  location,
	}, nil
}

// isLogData 识别并过滤开发日志。
// 如果文本包含 INFO/ERROR 等日志级别，且不含明确的会议链接，则视为日志。
func (p *smartParser) isLogData(text string) bool {
	// 常见的日志级别和堆栈特征
	logPatterns := []string{
		"INFO ", "ERROR ", "WARNING ", "DEBUG ", "TRACE ",
		"at ", "java.lang.", "Exception", "Stacktrace",
		"0x", "RuntimeError",
	}

	upperText := strings.ToUpper(text)
	for _, pattern := range logPatterns {
		if strings.Contains(upperText, strings.ToUpper(pattern)) {
			// 如果包含日志特征且没有明确的会议链接，则判定为日志
			linkRegex := regexp.MustCompile(`https?://[^\s\n]+`)
			if !linkRegex.MatchString(text) {
				return true
			}
		}
	}
	return false
}

func (p *smartParser) hasCalendarKeywords(text string) bool {
	keywords := []string{"分享", "评审", "会议", "PPT", "讨论", "需求", "优化", "同步", "周会", "Subject", "主题"}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func (p *smartParser) parseTime(text string) (time.Time, error) {
	now := time.Now()

	// 尝试匹配标准格式: 2026-04-02 10:00
	stdRegex := regexp.MustCompile(`(\d{4})[-/](\d{1,2})[-/](\d{1,2})\s+(\d{1,2}):(\d{2})`)
	if m := stdRegex.FindStringSubmatch(text); m != nil {
		year, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])
		hour, _ := strconv.Atoi(m[4])
		min, _ := strconv.Atoi(m[5])
		return time.Date(year, time.Month(month), day, hour, min, 0, 0, time.Local), nil
	}

	// 尝试匹配相对日期: 明天/后天/周五 + 时间
	var targetDate time.Time
	foundDate := false

	if strings.Contains(text, "今天") {
		targetDate = now
		foundDate = true
	} else if strings.Contains(text, "明天") {
		targetDate = now.AddDate(0, 0, 1)
		foundDate = true
	} else if strings.Contains(text, "后天") {
		targetDate = now.AddDate(0, 0, 2)
		foundDate = true
	}

	// 匹配 "周几"
	weekdays := map[string]time.Weekday{
		"周一": time.Monday, "周二": time.Tuesday, "周三": time.Wednesday,
		"周四": time.Thursday, "周五": time.Friday, "周六": time.Saturday, "周日": time.Sunday,
		"星期一": time.Monday, "星期二": time.Tuesday, "星期三": time.Wednesday,
		"星期四": time.Thursday, "星期五": time.Friday, "星期六": time.Saturday, "星期日": time.Sunday,
	}
	for kw, w := range weekdays {
		if strings.Contains(text, kw) {
			diff := int(w - now.Weekday())
			if diff <= 0 {
				diff += 7
			}
			targetDate = now.AddDate(0, 0, diff)
			foundDate = true
			break
		}
	}

	if !foundDate {
		return time.Time{}, errors.New("could not find date in text")
	}

	// 匹配时间点: 10:00 或 10点
	timeRegex := regexp.MustCompile(`(?:上午|下午)?\s*(\d{1,2})[:点](\d{2})?`)
	if m := timeRegex.FindStringSubmatch(text); m != nil {
		hour, _ := strconv.Atoi(m[1])
		min := 0
		if m[2] != "" {
			min, _ = strconv.Atoi(strings.Trim(m[2], "点"))
		}
		
		// 处理上午/下午
		if strings.Contains(m[0], "下午") && hour < 12 {
			hour += 12
		} else if strings.Contains(m[0], "上午") && hour == 12 {
			hour = 0
		}

		return time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), hour, min, 0, 0, time.Local), nil
	}

	return time.Time{}, errors.New("could not find time point in text")
}

func (p *smartParser) extractSubject(text string) string {
	lines := strings.Split(text, "\n")
	keywords := []string{"分享", "评审", "会议", "PPT", "讨论", "需求", "优化", "同步", "周会"}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" { continue }
		for _, kw := range keywords {
			if strings.Contains(trimmed, kw) {
				return trimmed
			}
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			runes := []rune(trimmed)
			if len(runes) > 50 {
				return string(runes[:50]) + "..."
			}
			return trimmed
		}
	}

	return "新建日程"
}
