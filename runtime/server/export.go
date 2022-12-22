package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/rilldata/rill/runtime/drivers"
)

// ExportTable exports a table or view as a flat file and triggers a HTTP download of it.
// It's mounted as a REST API only, and is not available over gRPC.
//
// TODO: This is a temporary hack that only supports DuckDB.
// We should add a generic workflow for data export that also supports possibly very large tables.
func (s *Server) ExportTable(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	var exportString string
	switch pathParams["format"] {
	case "csv":
		exportString = "FORMAT CSV, HEADER"
	case "parquet":
		exportString = "FORMAT PARQUET"
	default:
		http.Error(w, fmt.Sprintf("unknown format: %s", pathParams), http.StatusBadRequest)
	}

	if pathParams["instance_id"] == "" || pathParams["table_name"] == "" {
		http.Error(w, "missing params", http.StatusBadRequest)
		return
	}

	fileName := fmt.Sprintf("%s.%s", pathParams["table_name"], pathParams["format"])
	filePath := path.Join(os.TempDir(), fileName)
	defer os.Remove(filePath)

	// select * from the table and write to the temp file (DuckDB only)
	olap, err := s.runtime.OLAP(req.Context(), pathParams["instance_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = olap.Execute(req.Context(), &drivers.Statement{
		Query:    fmt.Sprintf("COPY (SELECT * FROM %s) TO '%s' (%s)", pathParams["table_name"], filePath, exportString),
		DryRun:   false,
		Priority: 0,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// set the header to trigger download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", req.Header.Get("Content-Type"))

	// read and stream the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
