package ui

import (
	_ "embed"
	"os"
	"path/filepath"
	"time"
)

//go:embed assets/icon_32.png
var DefaultIcon []byte

//go:embed assets/icon_success.png
var IconSuccess []byte

//go:embed assets/icon_duplicate.png
var IconDuplicate []byte

//go:embed assets/icon_conflict.png
var IconConflict []byte

//go:embed assets/icon_error.png
var IconError []byte

// GetIconPath 将嵌入的图标写入临时文件并返回绝对路径
func GetIconPath(name string, data []byte) string {
	home, _ := os.UserHomeDir()
	caDir := filepath.Join(home, ".calendar-assistant")
	_ = os.MkdirAll(caDir, 0755)
	
	path := filepath.Join(caDir, "ca_icon_"+name+".png")
	_ = os.Remove(path)
	_ = os.WriteFile(path, data, 0644)
	
	abs, _ := filepath.Abs(path)
	return abs
}

func Log(msg string) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".calendar-assistant", "debug.log")
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(time.Now().Format("15:04:05") + " " + msg + "\n")
}
