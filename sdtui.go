package main

import (
	"bytes"
	"os/exec"
	"strings"
	"log"
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

func getServiceName(unitPath string) string {
	strSplit := strings.Split(unitPath, "/")
	serviceName := strSplit[len(strSplit)-1]

	return serviceName
}

func getServiceStatus(unitPath string) string {
	cmd := exec.Command("systemctl", "status", getServiceName(unitPath))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run() // the err is not checked because systemctl returns non-zero code

	return out.String()
}

func getCurrentUnitPath(table *tview.Table) string {
	currentRow, _ := table.GetSelection()

	return table.GetCell(currentRow, 2).Text
}

func drawTable(table *tview.Table, unitList []ServiceUnit, filter string) {
	// draw headers
	table.Clear().
		SetFixed(1, 4).
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
				SetSelectable(false))

	// draw services
	index := 0
	for _, unit := range unitList {
		if filter == "" || strings.Contains(unit.File.Path, filter) {
			table.SetCell(index + 1, 0,
				tview.NewTableCell(unit.File.Type).
					SetMaxWidth(1).
					SetExpansion(1))
			table.SetCell(index + 1, 1,
				tview.NewTableCell(unit.Status.ActiveState).
					SetMaxWidth(1).
					SetExpansion(1))
			table.SetCell(index + 1, 2,
				tview.NewTableCell(unit.File.Path).
					SetMaxWidth(10).
					SetExpansion(10))
			table.SetCell(index + 1, 3,
				tview.NewTableCell(unit.Status.Description).
					SetMaxWidth(6).
					SetExpansion(6))
			index++
		}
	}
	table.ScrollToBeginning()
}

func main() {
	dbusConn, err := dbus.New()
	if err != nil {
		return
	}
	defer dbusConn.Close()

	allServiceUnits, err := getAllServiceUnits(dbusConn)
	if err != nil {
		return
	}

	filterText := ""
	statusShown := false
	app := tview.NewApplication()
	pages := tview.NewPages()
	grid := tview.NewGrid()
	filterInput := tview.NewInputField()
	sdUnitList := tview.NewTable()
	statusBox := tview.NewTextView()
	statusBox.SetBorder(true).
		SetTitle(" Service status ")
	helpText := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("(q) Exit (r) Reload/Restart (s) Start (S) Stop (e) Enable (d) Disable (/) Filter")

	filterInput.SetLabel("Filter by: ").
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter || key == tcell.KeyEscape {
				filterText = filterInput.GetText()
				grid.RemoveItem(filterInput).
					AddItem(helpText, 1, 0, 1, 1, 0, 0, false)
				drawTable(sdUnitList, allServiceUnits, filterText)
				app.SetFocus(sdUnitList)
			}
		})

	// display services
	sdUnitList.SetSelectable(true, false).
		SetBorder(true).
		SetBorderColor(tcell.ColorGray).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(" sdtui ")

	drawTable(sdUnitList, allServiceUnits, filterText)

	// define key handler
	sdUnitList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'r':
				fallthrough
			case 's':
				out := make(chan string)
				_, err = dbusConn.ReloadOrRestartUnit(getServiceName(getCurrentUnitPath(sdUnitList)), "replace", out)
				if err != nil {
					log.Fatal(err)
				}
				job := <-out
				if job != "done" {
					log.Fatal("Job is not done: ", job)
				}
				allServiceUnits, err = getAllServiceUnits(dbusConn)
				if err != nil {
					log.Fatal("Failed")
				}
				drawTable(sdUnitList, allServiceUnits, filterText)
				return nil
			case 'S':
				out := make(chan string)
				_, err = dbusConn.StopUnit(getServiceName(getCurrentUnitPath(sdUnitList)), "replace", out)
				if err != nil {
					log.Fatal(err)
				}
				job := <-out
				if job != "done" {
					log.Fatal("Job is not done: ", job)
				}
				allServiceUnits, err = getAllServiceUnits(dbusConn)
				if err != nil {
					log.Fatal("Failed")
				}
				drawTable(sdUnitList, allServiceUnits, filterText)
				return nil
			case 'e':
				_, _, err = dbusConn.EnableUnitFiles([]string{getServiceName(getCurrentUnitPath(sdUnitList))}, false, true)
				if err != nil {
					log.Fatal(err)
				}
				err = dbusConn.Reload()
				if err != nil {
					log.Fatal(err)
				}
				allServiceUnits, err = getAllServiceUnits(dbusConn)
				if err != nil {
					log.Fatal("Failed")
				}
				drawTable(sdUnitList, allServiceUnits, filterText)
				return nil
			case 'd':
				_, err = dbusConn.DisableUnitFiles([]string{getServiceName(getCurrentUnitPath(sdUnitList))}, false)
				if err != nil {
					log.Fatal(err)
				}
				err = dbusConn.Reload()
				if err != nil {
					log.Fatal(err)
				}
				allServiceUnits, err = getAllServiceUnits(dbusConn)
				if err != nil {
					log.Fatal("Failed")
				}
				drawTable(sdUnitList, allServiceUnits, filterText)
				return nil
			case '/':
				grid.RemoveItem(helpText).
					AddItem(filterInput, 1, 0, 1, 1, 0, 0, false)
				app.SetFocus(filterInput)
				return nil
			case ' ':
				//pages.HidePage("status")
				//statusShown = false
				//app.SetFocus(sdUnitList)
				statusBox.SetText(getServiceStatus(getCurrentUnitPath(sdUnitList)))
				pages.ShowPage("status")
				statusShown = true
				app.SetFocus(statusBox)
				return nil
			}
		}
		return event
	})

	grid.SetRows(0, 1).
		AddItem(sdUnitList, 0, 0, 1, 1, 0, 0, true).
		AddItem(helpText, 1, 0, 1, 1, 0, 0, false)

	modal := func(p tview.Primitive) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, 0, 7, false).
				AddItem(nil, 0, 1, false), 0, 7, false).
			AddItem(nil, 0, 1, false)
	}
	statusBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ':
				pages.HidePage("status")
				statusShown = false
				app.SetFocus(sdUnitList)
				return nil
			}
		}
		return event
	})

	pages.AddPage("main", grid, true, true).
		AddPage("status", modal(statusBox), true, false)

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}
