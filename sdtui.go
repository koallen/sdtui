package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	// display services
	sdUnitList := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	sdUnitList.SetBorder(true).
		SetBorderColor(tcell.ColorGray).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(" sdtui ")
	for row := 0; row < 10; row++ {
		sdUnitList.AddItem("test service", "", 0, nil)
	}

	// define key handler
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'r':
				sdUnitList.SetItemText(sdUnitList.GetCurrentItem(), "this service is restarted", "")
				return nil
			}
		}
		return event
	})

	frame := tview.NewFrame(sdUnitList).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText("(q) Exit (r) Restart service (e) Edit", false, tview.AlignCenter, tcell.ColorWhite)
	frame.SetBackgroundColor(tcell.ColorLime)
	if err := app.SetRoot(frame, true).SetFocus(sdUnitList).Run(); err != nil {
		panic(err)
	}
}
