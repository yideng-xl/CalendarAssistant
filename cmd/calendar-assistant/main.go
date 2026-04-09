package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/yideng/calendar-assistant/internal/config"
	"github.com/yideng/calendar-assistant/internal/monitor"
	"github.com/yideng/calendar-assistant/internal/parser"
	"github.com/yideng/calendar-assistant/internal/sync"
	"github.com/yideng/calendar-assistant/internal/ui"
)

// Version 版本号规则：
// - bug fix: 小版本 +1（如 0.5.1 → 0.5.2）
// - 新功能: 中版本 +1（如 0.5.1 → 0.6.1）
const Version = "0.5.1"

func main() {
	// 单实例检测
	lockFile := filepath.Join(os.TempDir(), "calendar-assistant.lock")
	if _, err := os.Stat(lockFile); err == nil {
		// 检查进程是否真的还在（简单处理：尝试删除，如果删不掉说明被占用）
		err = os.Remove(lockFile)
		if err != nil {
			// 已经有一个实例在运行了
			return 
		}
	}
	os.WriteFile(lockFile, []byte("locked"), 0644)
	defer os.Remove(lockFile)

	systray.Run(onReady, onExit)
}

func onReady() {
	ui.Log("App Starting...")
	systray.SetIcon(ui.DefaultIcon)
	systray.SetTitle("🟢 监听中")
	systray.SetTooltip("正在监听剪贴板日程")

	// 1. 初始化同步层与监听器
	provider := sync.NewCalendarProvider()
	
	// 启动时主动检查一次权限（仅弹一次授权对话框）
	go func() {
		ui.Log("Requesting Calendar Access...")
		if err := provider.CheckAuthorization(); err != nil {
			ui.Log("Calendar authorization failed: " + err.Error())
			sync.SendNotification("⚠️ 日历权限未授权", "请在系统设置 > 隐私与安全 > 自动化中允许访问日历", "")
		}
	}()
	opts := sync.SyncOptions{
		Reminders: []time.Duration{
			-15 * time.Minute,
			-5 * time.Minute,
			0,
		},
		Silent: true,
	}
	m := monitor.NewClipboardMonitor(provider, opts)

	// 2. 创建菜单项
	mStatus := systray.AddMenuItem("🟢 监听中", "点击暂停或恢复监听")
	systray.AddSeparator()
	
	mHistory := systray.AddMenuItem("最近同步历史", "")
	const maxHistory = 5
	historySlots := make([]*systray.MenuItem, maxHistory)
	historyTitles := make([]string, maxHistory)
	for i := 0; i < maxHistory; i++ {
		historySlots[i] = mHistory.AddSubMenuItem("", "")
		historySlots[i].Hide()
	}
	historyCount := 0

	mSettings := systray.AddMenuItem("提醒偏好设置", "")
	mRem15 := mSettings.AddSubMenuItemCheckbox("会议开始前 15 分钟", "", true)
	mRem5 := mSettings.AddSubMenuItemCheckbox("会议开始前 5 分钟", "", true)
	mRem0 := mSettings.AddSubMenuItemCheckbox("会议开始时", "", true)

	mAutoStart := systray.AddMenuItemCheckbox("开机自启动", "", config.IsAutoStartEnabled())

	systray.AddSeparator()
	systray.AddMenuItem("v"+Version, "当前版本")
	mQuit := systray.AddMenuItem("退出", "")

	// 3. 核心监听器启动
	ctx, cancel := context.WithCancel(context.Background())
	go m.Start(ctx)

	// 4. 处理同步回调
	m.SetOnSync(func(event *parser.MeetingEvent, success bool, message string) {
		title := "✅ 日程同步成功"
		finalMessage := "已添加到日历并设置提醒"
		iconData := ui.IconSuccess
		iconName := "success"
		
		if !success {
			if message == "DUPLICATE" {
				title = "🚫 日程已存在"
				finalMessage = "该内容已经同步过，无需重复添加"
				iconData = ui.IconDuplicate
				iconName = "duplicate"
			} else if strings.HasPrefix(message, "CONFLICT|") {
				title = "⚠️ 时间重叠提醒"
				finalMessage = fmt.Sprintf("已同步，但与已有会议 '%s' 冲突", message[9:])
				iconData = ui.IconConflict
				iconName = "conflict"
			} else {
				title = "❌ 同步失败"
				finalMessage = message
				iconData = ui.IconError
				iconName = "error"
			}
		}
		
		// 动态获取图标路径并发送原生通知
		iconPath := ui.GetIconPath(iconName, iconData)
		sync.SendNotification(title, fmt.Sprintf("%s\n%s", event.Subject, finalMessage), iconPath)

		// 更新历史记录菜单（使用预创建的槽位，避免动态添加不刷新）
		if success || strings.HasPrefix(message, "CONFLICT|") {
			// 已满时，向上滚动：把 1~4 的文本移到 0~3，腾出最后一个槽位
			if historyCount >= maxHistory {
				for i := 0; i < maxHistory-1; i++ {
					historyTitles[i] = historyTitles[i+1]
					historySlots[i].SetTitle(historyTitles[i])
				}
				historyCount = maxHistory - 1
			}
			label := fmt.Sprintf("[%s] %s", event.StartTime.Format("15:04"), event.Subject)
			historyTitles[historyCount] = label
			historySlots[historyCount].SetTitle(label)
			historySlots[historyCount].SetTooltip(event.Subject)
			historySlots[historyCount].Show()
			historyCount++
		}
	})

	// 5. 事件循环
	go func() {
		active := true
		for {
			select {
			case <-mStatus.ClickedCh:
				active = !active
				m.SetActive(active)
				if active {
					systray.SetTitle("🟢 监听中")
					mStatus.SetTitle("🟢 监听中")
					systray.SetTooltip("正在监听剪贴板日程")
				} else {
					systray.SetTitle("🔴 已暂停")
					mStatus.SetTitle("🔴 已暂停")
					systray.SetTooltip("已暂停监听剪贴板")
				}
			case <-mRem15.ClickedCh:
				if mRem15.Checked() { mRem15.Uncheck() } else { mRem15.Check() }
			case <-mRem5.ClickedCh:
				if mRem5.Checked() { mRem5.Uncheck() } else { mRem5.Check() }
			case <-mRem0.ClickedCh:
				if mRem0.Checked() { mRem0.Uncheck() } else { mRem0.Check() }
			case <-mAutoStart.ClickedCh:
				if mAutoStart.Checked() {
					if err := config.SetAutoStart(false); err == nil {
						mAutoStart.Uncheck()
					}
				} else {
					if err := config.SetAutoStart(true); err == nil {
						mAutoStart.Check()
					}
				}
			case <-mQuit.ClickedCh:
				cancel()
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	// 清理工作
}
