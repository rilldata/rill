import type { SvelteComponent } from "svelte";
import DimensionCell from "./DimensionCell.svelte";
import FormattedNumberCell from "./FormattedNumberCell.svelte";

export type CellRendering = {
  cellComponent: typeof SvelteComponent;
  cellComponentDefaultProps: Record<string, any>;
};

export const cellRenderings: CellRendering[] = [
  // 0: Dimension Cell
  {
    cellComponent: DimensionCell,
    cellComponentDefaultProps: {},
  },
  // 1: Measure total cell
  {
    cellComponent: FormattedNumberCell,
    cellComponentDefaultProps: {
      negClass: "text-red-500 font-semibold",
      posClass: "font-semibold",
    },
  },
  // 2: Percent of total cell
  {
    cellComponent: FormattedNumberCell,
    cellComponentDefaultProps: { negClass: "text-red-500", format: "0.1%" },
  },
  // 3: Absolute delta cell
  {
    cellComponent: FormattedNumberCell,
    cellComponentDefaultProps: { negClass: "text-red-500" },
  },
  // 4: Percent delta cell
  {
    cellComponent: FormattedNumberCell,
    cellComponentDefaultProps: { negClass: "text-red-500", format: "0.1%" },
  },
];

const pivotCell: CellRendering = {
  cellComponent: FormattedNumberCell,
  cellComponentDefaultProps: {},
};

export function getCellComponent(colIdx: number, measureFormat: string) {
  const cellRendering = cellRenderings[colIdx] ?? pivotCell;
  // If no format is specified, use the measure format
  if (!cellRendering.cellComponentDefaultProps.format) {
    cellRendering.cellComponentDefaultProps.format = measureFormat;
  }
  return cellRendering;
}
