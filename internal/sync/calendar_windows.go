//go:build windows
package sync

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/yideng/calendar-assistant/internal/parser"
)

type winCalendarProvider struct{}

func NewCalendarProvider() CalendarProvider {
	return &winCalendarProvider{}
}

func (p *winCalendarProvider) CheckAuthorization() error {
	// Windows 通过 Outlook COM 访问，无需额外授权
	// 但需要检查 Outlook 是否可用
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return errors.New("Outlook 未安装或无法访问，请确认已安装 Microsoft Outlook")
	}
	outlook, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer outlook.Release()
	return nil
}

func (p *winCalendarProvider) SyncEvent(event *parser.MeetingEvent, options SyncOptions) error {
	// 先检查重复
	isDup, err := p.HasEvent(event)
	if err != nil {
		return fmt.Errorf("检查重复事件失败: %v", err)
	}
	if isDup {
		return errors.New("DUPLICATE")
	}

	// 检查冲突
	conflicts, err := p.GetConflicts(event)
	if err != nil {
		return fmt.Errorf("检查冲突事件失败: %v", err)
	}

	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return errors.New("Outlook 未安装或无法访问")
	}
	outlook, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer outlook.Release()

	appointment, err := oleutil.CallMethod(outlook, "CreateItem", 1)
	if err != nil {
		return fmt.Errorf("创建日程项失败: %v", err)
	}
	appt := appointment.ToIDispatch()
	defer appt.Release()

	oleutil.PutProperty(appt, "Subject", event.Subject)
	oleutil.PutProperty(appt, "Start", event.StartTime.Format("2006-01-02 15:04:05"))
	oleutil.PutProperty(appt, "End", event.EndTime.Format("2006-01-02 15:04:05"))
	oleutil.PutProperty(appt, "Location", event.Location)
	oleutil.PutProperty(appt, "Body", event.Description)

	if len(options.Reminders) > 0 {
		oleutil.PutProperty(appt, "ReminderSet", true)
		oleutil.PutProperty(appt, "ReminderMinutesBeforeStart", int(options.Reminders[0].Abs().Minutes()))
	}

	_, err = oleutil.CallMethod(appt, "Save")
	if err != nil {
		return fmt.Errorf("保存日程失败: %v", err)
	}

	// 如果有冲突，返回冲突信息（事件仍然已创建）
	if len(conflicts) > 0 {
		return fmt.Errorf("CONFLICT|%s", conflicts[0])
	}

	return nil
}

// HasEvent 检查日历中是否已存在相同的事件（基于主题和开始时间）
func (p *winCalendarProvider) HasEvent(event *parser.MeetingEvent) (bool, error) {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return false, errors.New("Outlook 未安装或无法访问")
	}
	outlook, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer outlook.Release()

	// 获取 MAPI 命名空间
	ns, err := oleutil.CallMethod(outlook, "GetNamespace", "MAPI")
	if err != nil {
		return false, fmt.Errorf("无法访问 MAPI 命名空间: %v", err)
	}
	namespace := ns.ToIDispatch()
	defer namespace.Release()

	// 获取日历文件夹（olFolderCalendar = 9）
	folder, err := oleutil.CallMethod(namespace, "GetDefaultFolder", 9)
	if err != nil {
		return false, fmt.Errorf("无法访问日历文件夹: %v", err)
	}
	calFolder := folder.ToIDispatch()
	defer calFolder.Release()

	// 获取事件集合
	items, err := oleutil.GetProperty(calFolder, "Items")
	if err != nil {
		return false, fmt.Errorf("无法获取日历事件: %v", err)
	}
	calItems := items.ToIDispatch()
	defer calItems.Release()

	// 设置过滤条件：按开始时间和主题筛选
	startStr := event.StartTime.Format("01/02/2006 3:04 PM")
	endStr := event.StartTime.Add(time.Minute).Format("01/02/2006 3:04 PM")
	filter := fmt.Sprintf("[Start] >= '%s' AND [Start] < '%s' AND [Subject] = '%s'",
		startStr, endStr, strings.ReplaceAll(event.Subject, "'", "''"))

	restricted, err := oleutil.CallMethod(calItems, "Restrict", filter)
	if err != nil {
		return false, nil // 过滤失败不阻塞，返回无重复
	}
	restrictedItems := restricted.ToIDispatch()
	defer restrictedItems.Release()

	count, err := oleutil.GetProperty(restrictedItems, "Count")
	if err != nil {
		return false, nil
	}

	return count.Val > 0, nil
}

// GetConflicts 检查与给定事件时间段冲突的已有事件
func (p *winCalendarProvider) GetConflicts(event *parser.MeetingEvent) ([]string, error) {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return nil, errors.New("Outlook 未安装或无法访问")
	}
	outlook, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer outlook.Release()

	ns, err := oleutil.CallMethod(outlook, "GetNamespace", "MAPI")
	if err != nil {
		return nil, fmt.Errorf("无法访问 MAPI 命名空间: %v", err)
	}
	namespace := ns.ToIDispatch()
	defer namespace.Release()

	folder, err := oleutil.CallMethod(namespace, "GetDefaultFolder", 9)
	if err != nil {
		return nil, fmt.Errorf("无法访问日历文件夹: %v", err)
	}
	calFolder := folder.ToIDispatch()
	defer calFolder.Release()

	items, err := oleutil.GetProperty(calFolder, "Items")
	if err != nil {
		return nil, fmt.Errorf("无法获取日历事件: %v", err)
	}
	calItems := items.ToIDispatch()
	defer calItems.Release()

	// 查找时间段重叠的事件
	startStr := event.StartTime.Format("01/02/2006 3:04 PM")
	endStr := event.EndTime.Format("01/02/2006 3:04 PM")
	filter := fmt.Sprintf("[Start] < '%s' AND [End] > '%s'", endStr, startStr)

	restricted, err := oleutil.CallMethod(calItems, "Restrict", filter)
	if err != nil {
		return nil, nil
	}
	restrictedItems := restricted.ToIDispatch()
	defer restrictedItems.Release()

	count, err := oleutil.GetProperty(restrictedItems, "Count")
	if err != nil {
		return nil, nil
	}

	var conflicts []string
	for i := 1; i <= int(count.Val); i++ {
		item, err := oleutil.CallMethod(restrictedItems, "Item", i)
		if err != nil {
			continue
		}
		calItem := item.ToIDispatch()
		subject, err := oleutil.GetProperty(calItem, "Subject")
		if err == nil {
			conflicts = append(conflicts, subject.ToString())
		}
		calItem.Release()
	}

	return conflicts, nil
}

// SendNotification 使用 PowerShell 发送 Windows Toast 通知
func SendNotification(title, message, iconPath string) {
	// 使用 PowerShell 调用 Windows Toast 通知 API
	// 通过 stdin 传递脚本，避免命令行编码问题
	psScript := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom, ContentType = WindowsRuntime] | Out-Null

$template = @"
<toast>
    <visual>
        <binding template="ToastGeneric">
            <text>%s</text>
            <text>%s</text>
        </binding>
    </visual>
</toast>
"@

$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml($template)

$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("CalendarAssistant")
$notifier.Show($toast)
`,
		escapeXml(title),
		escapeXml(strings.ReplaceAll(message, "\n", "&#xA;")),
	)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", "-")
	cmd.Stdin = strings.NewReader(psScript)
	// 静默执行，不阻塞主线程
	go cmd.Run()
}

// escapeXml 转义 XML 特殊字符
func escapeXml(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
