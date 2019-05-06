package main

import (
	"strings"
	"github.com/coreos/go-systemd/dbus"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type ServiceUnit struct {
	File dbus.UnitFile
	Status dbus.UnitStatus
}

// this function collects all service units, regardless of their status
func getAllServiceUnits(conn *dbus.Conn) ([]ServiceUnit, error) {
	sdUnitFiles, err := conn.ListUnitFiles()
	if err != nil {
		return nil, err
	}
	sdUnits, err := conn.ListUnits()
	if err != nil {
		return nil, err
	}

	numOfServiceUnits := 0
	for _, unitFile := range sdUnitFiles {
		if strings.HasSuffix(unitFile.Path, ".service") {
			numOfServiceUnits++
		}
	}
	serviceUnits := make([]ServiceUnit, numOfServiceUnits)
	index := 0
	for _, unitFile := range sdUnitFiles {
		if !strings.HasSuffix(unitFile.Path, ".service") {
			continue
		}
		serviceUnits[index].File = unitFile
		strSplit := strings.Split(unitFile.Path, "/")
		serviceName := strSplit[len(strSplit)-1]
		for _, unitStatus := range sdUnits {
			if unitStatus.Name == serviceName {
				serviceUnits[index].Status = unitStatus
				break
			}
		}
		index++
	}

	return serviceUnits, nil
}

func main() {
	filterText := ""
	app := tview.NewApplication()
	filterInput := tview.NewInputField()
	sdUnitList := tview.NewTable()

	filterInput.SetLabel("Filter by: ").
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				filterText = filterInput.GetText()
				app.SetFocus(sdUnitList)
			}
		})

	// display services
	sdUnitList.SetFixed(1, 4).
		SetCell(0, 0,
			tview.NewTableCell("Enabled").
				SetTextColor(tcell.ColorTeal).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false)).
		SetCell(0, 1,
			tview.NewTableCell("Active").
				SetTextColor(tcell.ColorTeal).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false)).
		SetCell(0, 2,
			tview.NewTableCell("Path").
				SetTextColor(tcell.ColorTeal).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false)).
		SetCell(0, 3,
			tview.NewTableCell("Description").
				SetTextColor(tcell.ColorTeal).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false)).
		SetSelectable(true, false)
	sdUnitList.SetBorder(true).
		SetBorderColor(tcell.ColorGray).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(" sdtui ")

	dbusConn, err := dbus.New()
	if err != nil {
		return
	}
	defer dbusConn.Close()

	serviceUnits, err := getAllServiceUnits(dbusConn)
	if err != nil {
		return
	}
	for row, unit := range serviceUnits {
		sdUnitList.SetCell(row + 1, 0,
			tview.NewTableCell(unit.File.Type).
				SetMaxWidth(1).
				SetExpansion(1))
		sdUnitList.SetCell(row + 1, 1,
			tview.NewTableCell(unit.Status.ActiveState).
				SetMaxWidth(1).
				SetExpansion(1))
		sdUnitList.SetCell(row + 1, 2,
			tview.NewTableCell(unit.File.Path).
				SetMaxWidth(10).
				SetExpansion(10))
		sdUnitList.SetCell(row + 1, 3,
			tview.NewTableCell(unit.Status.Description).
				SetMaxWidth(6).
				SetExpansion(6))
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
				//sdUnitList.SetItemText(sdUnitList.GetCurrentItem(), "this service is restarted", "")
				return nil
			case '/':
				app.SetFocus(filterInput)
				return nil
			}
		}
		return event
	})

	grid := tview.NewGrid().
		SetRows(0, 1).
		AddItem(sdUnitList, 0, 0, 1, 1, 0, 0, true).
		AddItem(filterInput, 1, 0, 1, 1, 0, 0, false)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
