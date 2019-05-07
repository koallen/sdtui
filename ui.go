package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"strings"
)

func modal(p tview.Primitive) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, 0, 7, false).
			AddItem(nil, 0, 1, false), 0, 7, false).
		AddItem(nil, 0, 1, false)
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
			table.SetCell(index+1, 0,
				tview.NewTableCell(unit.File.Type).
					SetMaxWidth(1).
					SetExpansion(1))
			table.SetCell(index+1, 1,
				tview.NewTableCell(unit.Status.ActiveState).
					SetMaxWidth(1).
					SetExpansion(1))
			table.SetCell(index+1, 2,
				tview.NewTableCell(unit.File.Path).
					SetMaxWidth(10).
					SetExpansion(10))
			table.SetCell(index+1, 3,
				tview.NewTableCell(unit.Status.Description).
					SetMaxWidth(6).
					SetExpansion(6))
			index++
		}
	}
	table.ScrollToBeginning()
}
