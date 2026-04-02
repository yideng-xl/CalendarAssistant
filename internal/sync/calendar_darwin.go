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

type macOSCalendarProvider struct{}

func NewCalendarProvider() CalendarProvider {
	return &macOSCalendarProvider{}
}

func (p *macOSCalendarProvider) SyncEvent(event *parser.MeetingEvent, options SyncOptions) error {
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
		// time.Duration 是纳秒，转为秒。AppleScript 提醒通常是负数表示“提前”
		seconds := int(r.Seconds())
		if seconds > 0 {
			seconds = -seconds
		}
		sb.WriteString(fmt.Sprintf("            make new sound alarm at end with properties {trigger interval:%d}\n", seconds))
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
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, escapeAppleScriptString(message), escapeAppleScriptString(title))
	exec.Command("osascript", "-e", script).Run()
}
