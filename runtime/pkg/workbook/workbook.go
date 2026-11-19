// Package workbook renders a canvas dashboard, described by a CanvasWorkbookExportRequest, as an Excel workbook.
package workbook

import (
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/xuri/excelize/v2"
)

// Write validates req and writes it to w as an .xlsx file.
func Write(w io.Writer, req *runtimev1.CanvasWorkbookExportRequest) error {
	if err := validate(req); err != nil {
		return err
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	styles, err := newStyleSet(f)
	if err != nil {
		return err
	}
	wr := &writer{
		f:              f,
		req:            req,
		styles:         styles,
		names:          newNamer(),
		dataSheets:     make(map[string]*runtimev1.WorkbookDataSheet, len(req.DataSheets)),
		dataSheetNames: make(map[string]string, len(req.DataSheets)),
	}
	if err := wr.createSheets(); err != nil {
		return err
	}
	// Data sheets go first so the charts on the dashboard sheets reference cells that exist.
	for _, ds := range req.DataSheets {
		if err := wr.writeDataSheet(wr.dataSheetNames[ds.Id], ds); err != nil {
			return err
		}
	}
	for i, sheet := range req.DashboardSheets {
		if err := wr.writeDashboardSheet(wr.dashboardSheetNames[i], sheet); err != nil {
			return err
		}
	}
	f.SetActiveSheet(0)
	if err := f.SetDocProps(&excelize.DocProperties{Title: req.Meta.GetTitle(), Creator: "Rill"}); err != nil {
		return err
	}
	return f.Write(w)
}

type writer struct {
	f      *excelize.File
	req    *runtimev1.CanvasWorkbookExportRequest
	styles *styleSet
	names  *namer
	// Sheet names in the order of req.DashboardSheets.
	dashboardSheetNames []string
	// Data sheets and their sheet names by data sheet id.
	dataSheets     map[string]*runtimev1.WorkbookDataSheet
	dataSheetNames map[string]string
}

// createSheets allocates every sheet up front, dashboard sheets first, so the workbook keeps the canvas order.
func (w *writer) createSheets() error {
	for i, sheet := range w.req.DashboardSheets {
		title := sheet.Title
		if title == "" {
			title = "Dashboard"
		}
		name := w.names.sheetName(title)
		if i == 0 {
			// A new workbook starts with "Sheet1"; it becomes the main dashboard sheet.
			if err := w.f.SetSheetName("Sheet1", name); err != nil {
				return err
			}
		} else if _, err := w.f.NewSheet(name); err != nil {
			return err
		}
		w.dashboardSheetNames = append(w.dashboardSheetNames, name)
	}
	for _, ds := range w.req.DataSheets {
		title := "Data"
		if ds.Title != "" {
			title = "Data - " + ds.Title
		}
		name := w.names.sheetName(title)
		if _, err := w.f.NewSheet(name); err != nil {
			return err
		}
		w.dataSheets[ds.Id] = ds
		w.dataSheetNames[ds.Id] = name
	}
	return nil
}
