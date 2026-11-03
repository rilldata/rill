# workbook

Renders a canvas dashboard as an Excel workbook.

The input is a `CanvasWorkbookExportRequest` (see `proto/rill/runtime/v1/workbook.proto`), which the browser builds from the data it already displays for a canvas: the dashboard sheets with their blocks (native charts, PNG snapshots, KPI values, table previews, text), and the data sheets the charts and tables are bound to. The package validates the request against size caps, lays the blocks out on a uniform cell grid that mirrors the canvas geometry, writes each data sheet as a typed Excel table, and binds every native chart to the cells of its data sheet with excelize. Data sheets are written with excelize's stream writer, so their rows go to a temporary file instead of the in-memory sheet model; the total number of cells a request may hold is capped (`MaxCells`).

`Write` is the only entry point. It returns a `ValidationError` for requests that can't be rendered (unknown references, caps exceeded, bad images), which callers map to a client error; see `runtime/server/workbook.go`.
