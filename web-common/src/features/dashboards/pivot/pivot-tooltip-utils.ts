import type { Cell } from "@tanstack/svelte-table";
import type { MeasureColumnProps } from "./pivot-column-definition";
import type { PivotDataRow } from "./types";

export function getMeasureColumnForCell(
  cell: Cell<PivotDataRow, unknown>,
  measures: MeasureColumnProps,
) {
  const columnName =
    (cell.column.columnDef as { name?: string }).name ?? cell.column.id;
  return measures.find((m) => m.name === columnName);
}

export function getCellTooltipValue(
  cell: Cell<PivotDataRow, unknown>,
  measures: MeasureColumnProps,
) {
  const measureColumn = getMeasureColumnForCell(cell, measures);
  if (!measureColumn) return undefined;

  const value = cell.getValue() as string | number | null | undefined;
  const formattedValue = measureColumn.tooltipFormatter(value);
  if (formattedValue === null || formattedValue === undefined) return undefined;
  return formattedValue;
}
