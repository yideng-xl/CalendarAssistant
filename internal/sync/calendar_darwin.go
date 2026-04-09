//go:build darwin
package sync

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yideng/calendar-assistant/internal/parser"
	"github.com/yideng/calendar-assistant/internal/ui"
)

type macOSCalendarProvider struct {
	authorized    bool // 是否已通过授权检查
	authCheckDone bool // 是否已完成授权检查
}

func NewCalendarProvider() CalendarProvider {
	return &macOSCalendarProvider{}
}

// CheckAuthorization 检查日历访问权限，结果缓存
func (p *macOSCalendarProvider) CheckAuthorization() error {
	if p.authCheckDone && p.authorized {
		return nil
	}

	// 用一个轻量 AppleScript 测试是否有权限访问 Calendar
	script := `tell application "Calendar" to get name of first calendar`
	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		p.authCheckDone = true
		p.authorized = false
		ui.Log("Calendar authorization check failed: " + strings.TrimSpace(string(out)))
		return fmt.Errorf("日历访问未授权: %s", strings.TrimSpace(string(out)))
	}

	p.authCheckDone = true
	p.authorized = true
	ui.Log("Calendar authorization confirmed")
	return nil
}

func (p *macOSCalendarProvider) SyncEvent(event *parser.MeetingEvent, options SyncOptions) error {
	// 先检查授权状态，未授权则跳过，不重复弹窗
	if err := p.CheckAuthorization(); err != nil {
		return err
	}

	ui.Log("Syncing via AppleScript: " + event.Subject)

	// 准备 AppleScript 脚本
	// 1. 转换时间格式为 AppleScript 兼容的 format
	sDay, sMonth, sYear := event.StartTime.Day(), event.StartTime.Format("January"), event.StartTime.Year()
	sHour, sMin := event.StartTime.Hour(), event.StartTime.Minute()
	eDay, eMonth, eYear := event.EndTime.Day(), event.EndTime.Format("January"), event.EndTime.Year()
	eHour, eMin := event.EndTime.Hour(), event.EndTime.Minute()

	// 转义字段以防止注入
	escapedSubject := escapeAppleScriptString(event.Subject)
	escapedLocation := escapeAppleScriptString(event.Location)
	escapedDescription := escapeAppleScriptString(event.Description)

	// 2. 构建核心同步脚本
	// 调整日期设置顺序为：Day=1 -> Year -> Month -> Day -> Hours -> Minutes -> Seconds，彻底解决日期滚动 Bug
	script := fmt.Sprintf(`
    set startDate to (current date)
    set day of startDate to 1 -- 防止 31 号滚动
    set year of startDate to %d
    set month of startDate to %s
    set day of startDate to %d
    set hours of startDate to %d
    set minutes of startDate to %d
    set seconds of startDate to 0

    set endDate to (current date)
    set day of endDate to 1 -- 防止 31 号滚动
    set year of endDate to %d
    set month of endDate to %s
    set day of endDate to %d
    set hours of endDate to %d
    set minutes of endDate to %d
    set seconds of endDate to 0
`, sYear, sMonth, sDay, sHour, sMin,
		eYear, eMonth, eDay, eHour, eMin)

	script += fmt.Sprintf(`
    tell application "Calendar"
        try
            set theCal to first calendar whose name is "工作"
        on error
            set theCal to first calendar
        end try
        
        -- 重复检测
        set duplicateEvents to (every event of theCal whose summary is "%s" and start date is startDate)
        if (count of duplicateEvents) > 0 then
            return "DUPLICATE"
        end if
        
        -- 冲突检测
        set conflictEvents to (every event of theCal whose start date is less than endDate and end date is greater than startDate)
        
        -- 创建日程
        set currentEvent to make new event at theCal with properties {summary:"%s", start date:startDate, end date:endDate, location:"%s", description:"%s"}
        
        -- 设置提醒 (根据 options.Reminders)
        -- 注意：AppleScript 提醒是以秒为单位的负数，例如 -900 是提前 15 分钟
        %s

        if (count of conflictEvents) > 0 then
            set conflictTitle to summary of first item of conflictEvents
            return "CONFLICT|" & conflictTitle
        end if

        return "SUCCESS"
    end tell
    `, escapedSubject, escapedSubject, escapedLocation, escapedDescription,
		p.buildAlarmScript(options.Reminders))

	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	resStr := strings.TrimSpace(string(out))
	ui.Log("AppleScript result: " + resStr)

	if err != nil {
		return errors.New(resStr)
	}

	if resStr == "DUPLICATE" {
		return errors.New("DUPLICATE")
	}
	if strings.HasPrefix(resStr, "CONFLICT|") {
		return errors.New(resStr)
	}
	if resStr == "SUCCESS" {
		return nil
	}
	return errors.New(resStr)
}

func (p *macOSCalendarProvider) buildAlarmScript(reminders []time.Duration) string {
	if len(reminders) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("        tell currentEvent\n")
	for _, r := range reminders {
		// AppleScript trigger interval: minutes, negative = before event
		minutes := int(r.Minutes())
		if minutes > 0 {
			minutes = -minutes
		}
		line := fmt.Sprintf("            make new sound alarm at end with properties {trigger interval:%d}\n", minutes)
		sb.WriteString(line)
	}
	sb.WriteString("        end tell")
	return sb.String()
}

func (p *macOSCalendarProvider) HasEvent(event *parser.MeetingEvent) (bool, error) {
	// 复用 SyncEvent 的逻辑
	return false, nil
}

func (p *macOSCalendarProvider) GetConflicts(event *parser.MeetingEvent) ([]string, error) {
	return nil, nil
}

func escapeAppleScriptString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return strings.ReplaceAll(s, "\"", "\\\"")
}

func SendNotification(title, message, iconPath string) {
	// 使用 JXA（JavaScript for Automation）发送通知
	// 通过 stdin 管道传递脚本，避免命令行参数中文/emoji 编码乱码问题
	escapedTitle := strings.ReplaceAll(title, `\`, `\\`)
	escapedTitle = strings.ReplaceAll(escapedTitle, `"`, `\"`)
	escapedMessage := strings.ReplaceAll(message, `\`, `\\`)
	escapedMessage = strings.ReplaceAll(escapedMessage, `"`, `\"`)
	escapedMessage = strings.ReplaceAll(escapedMessage, "\n", `\n`)

	script := fmt.Sprintf(
		`var app = Application.currentApplication(); app.includeStandardAdditions = true; app.displayNotification("%s", {withTitle: "%s"});`,
		escapedMessage, escapedTitle,
	)
	cmd := exec.Command("osascript", "-l", "JavaScript")
	cmd.Stdin = strings.NewReader(script)
	cmd.Run()
}
