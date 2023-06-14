package queries

import (
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/structpb"
)

func writeXLSX(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, writer io.Writer) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()
	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}
	// styleID, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "777777"}})
	// if err != nil {
	// return err
	// }
	// showStripes := true
	// err = f.AddTable("Sheet1", &excelize.Table{
	// 	Range:             "A1:E100",
	// 	Name:              "table",
	// 	StyleName:         "TableStyleMedium2",
	// 	ShowFirstColumn:   true,
	// 	ShowLastColumn:    true,
	// 	ShowRowStripes:    &showStripes,
	// 	ShowColumnStripes: false,
	// })
	// if err != nil {
	// 	return err
	// }
	headers := make([]interface{}, 0, len(meta))
	for _, v := range meta {
		headers = append(headers, v.Name)
	}
	if err := sw.SetRow("A1",
		headers,
		excelize.RowOpts{Height: 45, Hidden: false}); err != nil {
		return err
	}

	row := make([]interface{}, len(meta))
	for i, s := range data {
		for _, f := range s.Fields {
			row = append(row, f.AsInterface())
		}
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			return err
		}

		if err := sw.SetRow(cell, row); err != nil {
			return err
		}
	}
	if err := sw.Flush(); err != nil {
		return err
	}
	err = f.Write(writer)
	// if err := f.SaveAs("Book1.xlsx"); err != nil {
	// return err
	// }

	// file, err := os.Open("Book1.xlsx")
	// if err != nil {
	// return err
	// }

	// defer file.Close()
	// _, err = io.Copy(writer, bufio.NewReader(file))

	return err
}
