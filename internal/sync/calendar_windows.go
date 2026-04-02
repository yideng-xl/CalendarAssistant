//go:build windows
package sync

import (
	"errors"
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/yideng/calendar-assistant/internal/parser"
)

type winCalendarProvider struct{}

func NewCalendarProvider() CalendarProvider {
	return &winCalendarProvider{}
}

func (p *winCalendarProvider) SyncEvent(event *parser.MeetingEvent, options SyncOptions) error {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return errors.New("Outlook is not installed or accessible")
	}
	outlook, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer outlook.Release()

	appointment, err := oleutil.CallMethod(outlook, "CreateItem", 1)
	if err != nil {
		return fmt.Errorf("failed to create appointment: %v", err)
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
		return fmt.Errorf("failed to save appointment: %v", err)
	}

	return nil
}

func (p *winCalendarProvider) HasEvent(event *parser.MeetingEvent) (bool, error) {
	return false, nil
}

func (p *winCalendarProvider) GetConflicts(event *parser.MeetingEvent) ([]string, error) {
	return nil, nil
}

func SendNotification(title, message, iconPath string) {
}
