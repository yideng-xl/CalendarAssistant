package monitor

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/yideng/calendar-assistant/internal/parser"
	"github.com/yideng/calendar-assistant/internal/sync"
)

type ClipboardMonitor struct {
	parsers  []parser.MeetingParser
	provider sync.CalendarProvider
	options  sync.SyncOptions
	lastHash string
	active   bool
	onSync   func(event *parser.MeetingEvent, success bool, message string)
}

func NewClipboardMonitor(provider sync.CalendarProvider, opts sync.SyncOptions) *ClipboardMonitor {
	return &ClipboardMonitor{
		parsers: []parser.MeetingParser{
			parser.NewTencentMeetingParser(),
			parser.NewDingTalkMeetingParser(),
			parser.NewSmartParser(),
		},
		provider: provider,
		options:  opts,
		active:   true,
	}
}

func (m *ClipboardMonitor) SetActive(active bool) {
	m.active = active
}

func (m *ClipboardMonitor) SetOnSync(fn func(event *parser.MeetingEvent, success bool, message string)) {
	m.onSync = fn
}

func (m *ClipboardMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !m.active {
				continue
			}
			m.checkClipboard()
		}
	}
}

func (m *ClipboardMonitor) checkClipboard() {
	text, err := clipboard.ReadAll()
	if err != nil || text == "" {
		return
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(text)))
	if hash == m.lastHash {
		return
	}
	m.lastHash = hash

	// 尝试所有解析器
	for _, p := range m.parsers {
		if p.CanParse(text) {
			event, err := p.Parse(text)
			if err == nil {
				m.syncEvent(event)
				return // 只要有一个解析成功就停止
			}
		}
	}
}

func (m *ClipboardMonitor) syncEvent(event *parser.MeetingEvent) {
	// 同步到日历
	err := m.provider.SyncEvent(event, m.options)
	
	if m.onSync != nil {
		if err != nil {
			m.onSync(event, false, err.Error())
		} else {
			m.onSync(event, true, "同步成功")
		}
	}
}
