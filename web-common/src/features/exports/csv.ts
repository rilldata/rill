import { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { markdownTables } from "./markdown-content";
import { mapChart } from "./xlsx/chart-mapping";
import { customChartTable, jsonCell } from "./xlsx/collect";
import type { ComponentExportPayload, ExportTable } from "./xlsx/payload";

// The data a canvas component shows, as CSV: the rows behind a chart,
// the result of each query behind a custom chart, or the tables in a markdown text.
// Undefined when the component has nothing to download.
// The rows are the ones the Excel export writes to the component's data sheets,
// and several tables are separated by an empty line.
export function componentCsv(
  payload: ComponentExportPayload | undefined,
  timeGrain: V1TimeGrain | undefined,
  timeZone: string,
): string | undefined {
  let tables: string[][][];
  switch (payload?.kind) {
    case "chart": {
      const { table } = mapChart({
        chartType: payload.chartType,
        spec: payload.spec,
        data: payload.data,
        timeGrain,
        timeZone,
      });
      tables = [exportTableRows(table)];
      break;
    }
    case "customChart":
      tables = payload.results.flatMap((result) => {
        const table = customChartTable(result, timeZone);
        return table ? [exportTableRows(table)] : [];
      });
      break;
    case "text":
      // Tables written in markdown and in HTML both count, and each cell holds the text it shows.
      tables = markdownTables(payload.unformattedText).map(
        (table) => table.rows,
      );
      break;
    default:
      return undefined;
  }
  if (tables.length === 0) return undefined;

  // RFC 4180 quoting; rows end in "\n" like the runtime's CSV exports.
  const csvField = (value: string) =>
    /[",\r\n]/.test(value) ? `"${value.replaceAll('"', '""')}"` : value;
  return tables
    .map((rows) =>
      rows.map((cells) => cells.map(csvField).join(",") + "\n").join(""),
    )
    .join("\n");
}

// The header and rows of a table as text. Like a data sheet, it drops a column
// named like an earlier one, as when a chart puts one measure on two channels.
function exportTableRows(table: ExportTable): string[][] {
  const seen = new Set<string>();
  const kept = table.columns.flatMap((column, i) => {
    if (seen.has(column.name)) return [];
    seen.add(column.name);
    return [i];
  });
  const cell = (value: unknown, i: number) => {
    const json = jsonCell(value);
    if (json === null) return "";
    // A date grain's times are midnights; the date alone says the same.
    if (
      table.columns[i].type === WorkbookColumnType.DATE &&
      typeof json === "string"
    ) {
      return json.slice(0, 10);
    }
    return typeof json === "object" ? JSON.stringify(json) : String(json);
  };
  return [
    kept.map((i) => table.columns[i].displayName),
    ...table.rows.map((row) => kept.map((i) => cell(row[i], i))),
  ];
}
