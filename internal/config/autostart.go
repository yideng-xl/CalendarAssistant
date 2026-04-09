package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"os/exec"
)

const AppName = "CalendarAssistant"

// SetAutoStart 设置自启动状态
func SetAutoStart(enabled bool) error {
	if runtime.GOOS == "darwin" {
		return setAutoStartDarwin(enabled)
	} else if runtime.GOOS == "windows" {
		return setAutoStartWindows(enabled)
	}
	return nil
}

// IsAutoStartEnabled 检查当前是否已设置自启动
func IsAutoStartEnabled() bool {
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		plistPath := filepath.Join(home, "Library/LaunchAgents", "com.yideng.calendar-assistant.plist")
		_, err := os.Stat(plistPath)
		return err == nil
	}
	// Windows 注册表检查
	if runtime.GOOS == "windows" {
		cmd := exec.Command("reg", "query", "HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Run", "/v", AppName)
		err := cmd.Run()
		return err == nil
	}
	return false
}

func setAutoStartDarwin(enabled bool) error {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library/LaunchAgents", "com.yideng.calendar-assistant.plist")

	if !enabled {
		_ = exec.Command("launchctl", "unload", plistPath).Run()
		return os.Remove(plistPath)
	}

	execPath := "/Applications/CalendarAssistant.app/Contents/MacOS/calendar-assistant"
	// 如果不在应用程序目录，尝试获取当前运行路径（仅作备选）
	if _, err := os.Stat(execPath); err != nil {
		p, _ := os.Executable()
		execPath = p
	}

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.yideng.calendar-assistant</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`, execPath)

	err := os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		return err
	}
	
	return nil // 仅写入文件，macOS 在下次登录时会自动扫描并启动
}

func setAutoStartWindows(enabled bool) error {
	// Windows 使用注册表实现
	// 命令: reg add "HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run" /v "CalendarAssistant" /t REG_SZ /d "路径" /f
	p, _ := os.Executable()
	execPath := filepath.Clean(p)
	
	if enabled {
		cmd := exec.Command("reg", "add", "HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Run", "/v", AppName, "/t", "REG_SZ", "/d", execPath, "/f")
		return cmd.Run()
	} else {
		cmd := exec.Command("reg", "delete", "HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Run", "/v", AppName, "/f")
		return cmd.Run()
	}
}
