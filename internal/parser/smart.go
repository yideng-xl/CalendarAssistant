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
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	// SmartParser 作为兜底，仍需要基本的日程信号才触发：
	// 必须同时包含"日程关键词"和"时间线索"，否则不尝试解析
	return p.hasCalendarKeywords(text) && p.hasTimeHint(text)
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

	// 1. 提取时间（支持时区）
	startTime, err := p.parseTime(text)
	if err != nil {
		return nil, err
	}
	endTime := startTime.Add(time.Hour)

	// 2. 尝试提取结束时间（时间范围格式）
	if et, ok := p.parseEndTime(text, startTime); ok {
		endTime = et
	}

	// 3. 提取标题
	subject := p.extractSubject(text)

	return &MeetingEvent{
		Subject:   subject,
		StartTime: startTime,
		EndTime:   endTime,
		Location:  location,
	}, nil
}

// ParseMultiple 解析文本中的多个日程事件
// 当一段文本包含多个时间点和主题时，分别提取为独立事件
func (p *smartParser) ParseMultiple(text string) ([]*MeetingEvent, error) {
	if p.isLogData(text) {
		return nil, errors.New("detected as log data, skipping")
	}

	// 按常见分隔模式拆分事件块
	blocks := p.splitEventBlocks(text)

	if len(blocks) <= 1 {
		// 只有一个事件块，退回到单事件解析
		event, err := p.Parse(text)
		if err != nil {
			return nil, err
		}
		return []*MeetingEvent{event}, nil
	}

	var events []*MeetingEvent
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		// 每个块必须包含时间线索
		if !p.hasTimeHint(block) {
			continue
		}
		event, err := p.Parse(block)
		if err != nil {
			continue
		}
		events = append(events, event)
	}

	if len(events) == 0 {
		return nil, errors.New("no events found in text")
	}
	return events, nil
}

// splitEventBlocks 将文本按事件分隔符拆分
func (p *smartParser) splitEventBlocks(text string) []string {
	// 策略1: 按编号分隔 "1. xxx", "2. xxx" 或 "① ② ③"
	numberedRegex := regexp.MustCompile(`(?m)^\s*(?:\d+[.、)）]|[①②③④⑤⑥⑦⑧⑨⑩])\s*`)
	if numberedRegex.MatchString(text) {
		parts := numberedRegex.Split(text, -1)
		var result []string
		for _, part := range parts {
			if strings.TrimSpace(part) != "" {
				result = append(result, part)
			}
		}
		if len(result) > 1 {
			return result
		}
	}

	// 策略2: 按空行+时间模式分隔
	lines := strings.Split(text, "\n")
	var blocks []string
	var current strings.Builder
	timePattern := regexp.MustCompile(`\d{4}[-/]\d{1,2}[-/]\d{1,2}|(?:今天|明天|后天|下?周[一二三四五六日]|下?星期[一二三四五六日])`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 空行后跟有时间模式的行，视为新事件开始
		if trimmed == "" && current.Len() > 0 {
			// 先把当前块暂存，待判断下一行是否为新事件
			blocks = append(blocks, current.String())
			current.Reset()
			continue
		}
		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)
	}
	if current.Len() > 0 {
		blocks = append(blocks, current.String())
	}

	// 合并没有时间信息的块到前一个块
	var merged []string
	for _, block := range blocks {
		if len(merged) > 0 && !timePattern.MatchString(block) {
			merged[len(merged)-1] += "\n" + block
		} else {
			merged = append(merged, block)
		}
	}

	return merged
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
	// 只保留真正的日程/会议相关关键词，去掉"同步"、"讨论"、"需求"、"优化"等日常用语
	keywords := []string{"会议", "日程", "评审", "周会", "晨会", "站会", "分享会", "培训", "面试", "约会", "Subject", "主题", "邀请"}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

// hasTimeHint 检查文本中是否包含时间线索
func (p *smartParser) hasTimeHint(text string) bool {
	timePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\d{4}[-/]\d{1,2}[-/]\d{1,2}\s+\d{1,2}:\d{2}`),                               // 2026-04-02 10:00
		regexp.MustCompile(`\d{4}年\d{1,2}月\d{1,2}[日号]`),                                                 // 2026年4月10日/号
		regexp.MustCompile(`\d{1,2}月\d{1,2}[日号]`),                                                        // 4月10日/号
		regexp.MustCompile(`(?:今天|明天|后天|大后天|下?周[一二三四五六日天]|下?星期[一二三四五六日天]|本周[一二三四五六日天])`), // 相对日期
		regexp.MustCompile(`(?:上午|下午|早上|晚上)\s*\d{1,2}[:点]`),                                          // 上午10点
		regexp.MustCompile(`\d{1,2}[:点]\d{2}`),                                                             // 10:00 或 10点30
	}
	for _, re := range timePatterns {
		if re.MatchString(text) {
			return true
		}
	}

	// 检查中文数字时间：四点半、两点、十点四十五 等
	if p.parseChineseNumberTime(text) != nil {
		return true
	}

	return false
}

// chineseNumMap 中文数字到阿拉伯数字的映射
var chineseNumMap = map[rune]int{
	'零': 0, '〇': 0, '一': 1, '二': 2, '两': 2, '三': 3,
	'四': 4, '五': 5, '六': 6, '七': 7, '八': 8, '九': 9,
	'十': 10,
}

// parseChineseNum 将中文数字字符串转换为阿拉伯数字（支持 0-59 范围）
// 例如："四" → 4, "十" → 10, "十二" → 12, "二十五" → 25, "三十" → 30
func parseChineseNum(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	runes := []rune(s)
	result := 0
	hasValue := false

	if len(runes) == 1 {
		if v, ok := chineseNumMap[runes[0]]; ok {
			return v, true
		}
		return 0, false
	}

	for i, r := range runes {
		if r == '十' {
			hasValue = true
			if i == 0 {
				result = 10 // "十二" = 12
			} else {
				// result 已有十位前缀，乘 10
				result *= 10
			}
		} else if v, ok := chineseNumMap[r]; ok {
			hasValue = true
			if i > 0 && runes[i-1] == '十' {
				result += v // "十二" → 10 + 2
			} else {
				result = v // 首位数字
			}
		}
	}
	return result, hasValue
}

// chineseTimeResult 中文时间解析结果
type chineseTimeResult struct {
	hour int
	min  int
}

// parseChineseNumberTime 解析中文数字时间表达
// 支持格式：四点半、三点、两点四十五、下午四点半、早上九点 等
func (p *smartParser) parseChineseNumberTime(text string) *chineseTimeResult {
	// 匹配 "[上午|下午|早上|晚上] + 中文数字 + 点 + [半|中文数字分钟]"
	cnTimeRegex := regexp.MustCompile(`(上午|下午|早上|晚上)?([零〇一二两三四五六七八九十]+)点(半|([零〇一二两三四五六七八九十]+))?`)
	m := cnTimeRegex.FindStringSubmatch(text)
	if m == nil {
		return nil
	}

	period := m[1]  // 上午/下午/早上/晚上（可能为空）
	hourStr := m[2] // 中文数字小时
	halfOrMin := m[3] // "半" 或中文数字分钟

	hour, ok := parseChineseNum(hourStr)
	if !ok || hour > 24 {
		return nil
	}

	min := 0
	if halfOrMin == "半" {
		min = 30
	} else if m[4] != "" {
		min, _ = parseChineseNum(m[4])
	}

	// 处理上午/下午
	if period == "下午" || period == "晚上" {
		if hour < 12 {
			hour += 12
		}
	} else if period == "上午" || period == "早上" {
		if hour == 12 {
			hour = 0
		}
	} else {
		// 无前缀时，如果 hour <= 6 默认为下午（口语习惯："四点半"通常指16:30）
		if hour >= 1 && hour <= 6 {
			hour += 12
		}
	}

	return &chineseTimeResult{hour: hour, min: min}
}

func (p *smartParser) parseTime(text string) (time.Time, error) {
	now := time.Now()

	// 检查并处理时区标注
	tzOffset := p.detectTimezone(text)

	// 尝试匹配标准格式: 2026-04-02 10:00
	stdRegex := regexp.MustCompile(`(\d{4})[-/](\d{1,2})[-/](\d{1,2})\s+(\d{1,2}):(\d{2})`)
	if m := stdRegex.FindStringSubmatch(text); m != nil {
		year, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])
		hour, _ := strconv.Atoi(m[4])
		min, _ := strconv.Atoi(m[5])
		t := time.Date(year, time.Month(month), day, hour, min, 0, 0, time.Local)
		return p.applyTimezoneOffset(t, tzOffset), nil
	}

	// 尝试匹配中文年月日格式: 2026年4月10日 14:00
	cnDateRegex := regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})[日号]\s*(\d{1,2}):(\d{2})`)
	if m := cnDateRegex.FindStringSubmatch(text); m != nil {
		year, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])
		hour, _ := strconv.Atoi(m[4])
		min, _ := strconv.Atoi(m[5])
		t := time.Date(year, time.Month(month), day, hour, min, 0, 0, time.Local)
		return p.applyTimezoneOffset(t, tzOffset), nil
	}

	// 尝试匹配简短中文日期: 4月10日 14:00 或 4月10号 下午2点
	shortCnRegex := regexp.MustCompile(`(\d{1,2})月(\d{1,2})[日号]`)
	if m := shortCnRegex.FindStringSubmatch(text); m != nil {
		month, _ := strconv.Atoi(m[1])
		day, _ := strconv.Atoi(m[2])
		year := now.Year()
		// 如果月份已过，默认为明年
		if time.Month(month) < now.Month() || (time.Month(month) == now.Month() && day < now.Day()) {
			year++
		}
		targetDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		if t, ok := p.extractTimePoint(text, targetDate); ok {
			return p.applyTimezoneOffset(t, tzOffset), nil
		}
		return time.Time{}, errors.New("found date but no time point")
	}

	// 尝试匹配相对日期
	var targetDate time.Time
	foundDate := false

	if strings.Contains(text, "大后天") {
		targetDate = now.AddDate(0, 0, 3)
		foundDate = true
	} else if strings.Contains(text, "后天") {
		targetDate = now.AddDate(0, 0, 2)
		foundDate = true
	} else if strings.Contains(text, "明天") {
		targetDate = now.AddDate(0, 0, 1)
		foundDate = true
	} else if strings.Contains(text, "今天") {
		targetDate = now
		foundDate = true
	}

	// 匹配 "下周几" / "本周几"
	nextWeekRegex := regexp.MustCompile(`下(?:周|星期)([一二三四五六日天])`)
	thisWeekRegex := regexp.MustCompile(`本周([一二三四五六日天])`)

	if !foundDate {
		if m := nextWeekRegex.FindStringSubmatch(text); m != nil {
			w := p.weekdayFromChinese(m[1])
			diff := int(w - now.Weekday())
			if diff <= 0 {
				diff += 7
			}
			diff += 7 // 下周
			targetDate = now.AddDate(0, 0, diff)
			foundDate = true
		} else if m := thisWeekRegex.FindStringSubmatch(text); m != nil {
			w := p.weekdayFromChinese(m[1])
			diff := int(w - now.Weekday())
			if diff <= 0 {
				diff += 7
			}
			targetDate = now.AddDate(0, 0, diff)
			foundDate = true
		}
	}

	// 匹配 "周几" / "星期几"（无前缀时默认为最近的下一个）
	if !foundDate {
		weekdays := map[string]time.Weekday{
			"周一": time.Monday, "周二": time.Tuesday, "周三": time.Wednesday,
			"周四": time.Thursday, "周五": time.Friday, "周六": time.Saturday, "周日": time.Sunday,
			"周天": time.Sunday,
			"星期一": time.Monday, "星期二": time.Tuesday, "星期三": time.Wednesday,
			"星期四": time.Thursday, "星期五": time.Friday, "星期六": time.Saturday, "星期日": time.Sunday,
			"星期天": time.Sunday,
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
	}

	// 如果没找到日期但有中文数字时间（如"四点半"），默认为今天
	if !foundDate {
		if ct := p.parseChineseNumberTime(text); ct != nil {
			t := time.Date(now.Year(), now.Month(), now.Day(), ct.hour, ct.min, 0, 0, time.Local)
			// 如果解析出的时间已过，默认为明天
			if t.Before(now) {
				t = t.AddDate(0, 0, 1)
			}
			return p.applyTimezoneOffset(t, tzOffset), nil
		}
		return time.Time{}, errors.New("could not find date in text")
	}

	// 先尝试阿拉伯数字时间点
	if t, ok := p.extractTimePoint(text, targetDate); ok {
		return p.applyTimezoneOffset(t, tzOffset), nil
	}

	// 再尝试中文数字时间点（如"明天四点半"）
	if ct := p.parseChineseNumberTime(text); ct != nil {
		t := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), ct.hour, ct.min, 0, 0, time.Local)
		return p.applyTimezoneOffset(t, tzOffset), nil
	}

	return time.Time{}, errors.New("could not find time point in text")
}

// extractTimePoint 从文本中提取时间点，结合目标日期构建完整时间
func (p *smartParser) extractTimePoint(text string, targetDate time.Time) (time.Time, bool) {
	// 匹配时间点: 上午10:00 / 下午2点30 / 10:00 / 14点
	timeRegex := regexp.MustCompile(`(?:上午|下午|早上|晚上)?\s*(\d{1,2})[:点](\d{2})?`)
	if m := timeRegex.FindStringSubmatch(text); m != nil {
		hour, _ := strconv.Atoi(m[1])
		min := 0
		if m[2] != "" {
			min, _ = strconv.Atoi(strings.Trim(m[2], "点"))
		}

		// 处理上午/下午/早上/晚上
		prefix := m[0]
		if strings.Contains(prefix, "下午") || strings.Contains(prefix, "晚上") {
			if hour < 12 {
				hour += 12
			}
		} else if strings.Contains(prefix, "上午") || strings.Contains(prefix, "早上") {
			if hour == 12 {
				hour = 0
			}
		}

		return time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), hour, min, 0, 0, time.Local), true
	}
	return time.Time{}, false
}

// parseEndTime 尝试从文本中提取结束时间（支持 "-11:30" 或 "至 15:00" 格式）
func (p *smartParser) parseEndTime(text string, startTime time.Time) (time.Time, bool) {
	// 匹配 "10:00-11:30" 或 "10:00 - 11:30" 或 "10:00～11:30"
	rangeRegex := regexp.MustCompile(`\d{1,2}:\d{2}\s*[-~～至]\s*(\d{1,2}):(\d{2})`)
	if m := rangeRegex.FindStringSubmatch(text); m != nil {
		hour, _ := strconv.Atoi(m[1])
		min, _ := strconv.Atoi(m[2])
		endTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), hour, min, 0, 0, time.Local)
		if endTime.After(startTime) {
			return endTime, true
		}
	}
	return time.Time{}, false
}

// detectTimezone 检测文本中的时区标注，返回与本地时区的偏移量
// 返回 nil 表示未检测到时区或为本地时区
func (p *smartParser) detectTimezone(text string) *time.Location {
	// 常见时区映射
	tzMap := map[string]*time.Location{
		"UTC":  time.UTC,
		"GMT":  time.UTC,
	}

	// 固定偏移时区
	fixedTzMap := map[string]int{
		"PST": -8, "PDT": -7,
		"EST": -5, "EDT": -4,
		"CST": -6, "CDT": -5, // 美国中部（注意：CST 在中文场景通常指中国标准时间 UTC+8）
		"MST": -7, "MDT": -6,
		"JST": 9,
		"KST": 9,
		"IST": 5, // 印度 UTC+5:30 简化为5
	}

	// 检查 UTC+N / GMT+N 格式
	utcOffsetRegex := regexp.MustCompile(`(?:UTC|GMT)\s*([+-])(\d{1,2})(?::(\d{2}))?`)
	if m := utcOffsetRegex.FindStringSubmatch(text); m != nil {
		hours, _ := strconv.Atoi(m[2])
		minutes := 0
		if m[3] != "" {
			minutes, _ = strconv.Atoi(m[3])
		}
		offset := hours*3600 + minutes*60
		if m[1] == "-" {
			offset = -offset
		}
		return time.FixedZone("Custom", offset)
	}

	// 检查命名时区
	for name, loc := range tzMap {
		if strings.Contains(text, name) {
			return loc
		}
	}
	for name, offset := range fixedTzMap {
		// 避免 "CST" 误判（在中文上下文中通常指 UTC+8 即本地时间）
		if name == "CST" && (strings.Contains(text, "中国") || strings.Contains(text, "北京")) {
			continue
		}
		tzRegex := regexp.MustCompile(`\b` + name + `\b`)
		if tzRegex.MatchString(text) {
			return time.FixedZone(name, offset*3600)
		}
	}

	return nil // 未检测到，使用本地时区
}

// applyTimezoneOffset 将源时区的时间转换为本地时间
func (p *smartParser) applyTimezoneOffset(t time.Time, sourceTz *time.Location) time.Time {
	if sourceTz == nil {
		return t // 无时区信息，保持不变
	}
	// 将时间解释为源时区，然后转为本地时区
	sourceTime := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, sourceTz)
	return sourceTime.In(time.Local)
}

// weekdayFromChinese 将中文星期映射为 time.Weekday
func (p *smartParser) weekdayFromChinese(s string) time.Weekday {
	m := map[string]time.Weekday{
		"一": time.Monday, "二": time.Tuesday, "三": time.Wednesday,
		"四": time.Thursday, "五": time.Friday, "六": time.Saturday,
		"日": time.Sunday, "天": time.Sunday,
	}
	if w, ok := m[s]; ok {
		return w
	}
	return time.Monday
}

func (p *smartParser) extractSubject(text string) string {
	lines := strings.Split(text, "\n")
	keywords := []string{"分享", "评审", "会议", "PPT", "讨论", "需求", "优化", "同步", "周会"}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
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
