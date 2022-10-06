import type { VirtualizedTableConfig } from "../virtualized-table/types";

export const DimensionTableConfig: VirtualizedTableConfig = {
  defaultColumnWidth: 120,
  maxColumnWidth: 320,
  minColumnWidth: 104,
  minHeaderWidthWhenColumsAreSmall: 160,
  rowHeight: 24,
  columnHeaderHeight: 28,
  indexWidth: 24,
  columnHeaderFontWeightClass: "font-normal",
  defaultFontWeightClass: "font-normal",
  table: "DimensionTable",
};
