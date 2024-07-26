package mywidge

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func buildMonthSelect(dateToShow time.Time, owningDialog *dialog.Dialog, targetElement *widget.Entry) *fyne.Container {
	// Calculate the days shown
	startOfMonth, _ := time.Parse("2006-January-03", fmt.Sprintf("%s-%s", dateToShow.Format("2006-January"), "01"))
	startOfMonthDisplay := startOfMonth
	startOffset := int(startOfMonth.Weekday())
	if startOffset != 6 {
		startOfMonthDisplay = startOfMonthDisplay.AddDate(0, 0, -1*int(startOfMonth.Weekday()))
	} else {
		startOffset = 0
	}
	totalDays := startOffset + startOfMonth.AddDate(0, 1, -1).Day()
	remainder := totalDays % 7
	if remainder > 0 {
		totalDays += 7 - totalDays%7
	}

	days := []fyne.CanvasObject{
		widget.NewLabelWithStyle("S", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("M", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("T", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("W", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("T", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("F", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("S", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	}
	thisDay := startOfMonthDisplay
	todayString := strings.ReplaceAll(targetElement.Text, "/", "-")
	for i := 0; i < totalDays; i++ {
		mike := thisDay
		bg := canvas.NewRectangle(color.NRGBA{R: 220, G: 220, B: 220, A: 128})
		if thisDay.Format("2006-01-02") == todayString {
			bg = canvas.NewRectangle(color.NRGBA{R: 100, G: 200, B: 150, A: 128})
		}
		days = append(
			days,
			container.NewStack(
				widget.NewButton(fmt.Sprintf("%d", thisDay.Day()), func() {
					targetElement.SetText(mike.Format("2006-01-02"))
					(*owningDialog).Hide()
				}),
				bg,
			))
		thisDay = thisDay.AddDate(0, 0, 1)
	}
	return container.NewGridWithColumns(7,
		days...)
}

func CreateDatePicker(dateToShow time.Time, owningDialog *dialog.Dialog, targetElement *widget.Entry) fyne.CanvasObject {
	var calendarWidget *fyne.Container
	var monthSelect *widget.Label
	var monthDisplay *fyne.Container
	var backMonth *widget.Button
	var forwardMonth *widget.Button

	monthSelect = widget.NewLabel(dateToShow.Format("January 2006"))

	monthDisplay = buildMonthSelect(dateToShow, owningDialog, targetElement)

	backMonth = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		dateToShow = dateToShow.AddDate(0, -1, 0)
		monthSelect = widget.NewLabel(dateToShow.Format("January 2006"))
		monthDisplay = buildMonthSelect(dateToShow, owningDialog, targetElement)
		calendarWidget.RemoveAll()
		calendarWidget.Add(container.NewBorder(
			container.NewHBox(
				backMonth,
				layout.NewSpacer(),
				monthSelect,
				layout.NewSpacer(),
				forwardMonth,
			),
			nil,
			nil,
			nil,
			monthDisplay))
		calendarWidget.Refresh()
	})
	forwardMonth = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		dateToShow = dateToShow.AddDate(0, 1, 0)
		monthSelect = widget.NewLabel(dateToShow.Format("January 2006"))
		monthDisplay = buildMonthSelect(dateToShow, owningDialog, targetElement)
		calendarWidget.RemoveAll()
		calendarWidget.Add(container.NewBorder(
			container.NewHBox(
				backMonth,
				layout.NewSpacer(),
				monthSelect,
				layout.NewSpacer(),
				forwardMonth,
			),
			nil,
			nil,
			nil,
			monthDisplay))
		calendarWidget.Refresh()
	})
	// Build the UI
	// Note: RemoveAll/Add required so the above back/Forward months look the same
	calendarWidget = container.NewHBox(widget.NewLabel("Loading"))
	calendarWidget.RemoveAll()
	calendarWidget.Add(container.NewBorder(
		container.NewHBox(
			backMonth,
			layout.NewSpacer(),
			monthSelect,
			layout.NewSpacer(),
			forwardMonth,
		),
		nil,
		nil,
		nil,
		monthDisplay))
	return calendarWidget
}

var CalendarResource = fyne.NewStaticResource("calendar.svg", []byte(`<?xml version="1.0" encoding="utf-8"?>
	<!-- Svg Vector Icons : http://www.onlinewebfonts.com/icon -->
	<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
	<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px" viewBox="0 0 256 256" enable-background="new 0 0 256 256" xml:space="preserve">
	<metadata> Svg Vector Icons : http://www.onlinewebfonts.com/icon </metadata>
	<g><g><path fill="#000000" d="M125.1,111.1"/><path fill="#000000" d="M94.5,150.9H60.7v-33.7h33.7V150.9z M145.4,117.1h-33.7v33.7h33.7V117.1z M196.4,117.1h-33.7v33.7h33.7V117.1z M94.5,166.4H60.7v33.7h33.7V166.4z M145.4,166.4h-33.7v33.7h33.7V166.4z M196.4,166.4h-33.7v33.7h33.7V166.4z M209,49c0,5.8-4.7,10.4-10.4,10.4c-5.8,0-10.4-4.7-10.4-10.4V22.8c0-5.8,4.7-10.4,10.4-10.4c5.8,0,10.4,4.7,10.4,10.4V49z M68.4,22.8c0-5.8-4.7-10.4-10.4-10.4c-5.8,0-10.4,4.7-10.4,10.4V49c0,5.8,4.7,10.4,10.4,10.4c5.8,0,10.4-4.7,10.4-10.4V22.8z M221.6,38v10.7c0,12.7-10.3,23-23,23c-12.7,0-23-10.3-23-23V38H80.9v10.6c0,12.7-10.3,23-23,23c-12.7,0-23-10.3-23-23V38H10v205.6h236V38H221.6z M223.1,220.7H32.9V93.3h190.2L223.1,220.7L223.1,220.7z"/></g></g>
	</svg>`))
