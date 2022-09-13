import type { VirtualizedTableConfig } from "$lib/components/virtualized-table/types";

export const DimensionTableConfig: VirtualizedTableConfig = {
  defaultColumnWidth: 200,
  maxColumnWidth: 320,
  minColumnWidth: 104,
  minHeaderWidthWhenColumsAreSmall: 160,
  rowHeight: 24,
  columnHeaderHeight: 28,
  indexWidth: 60,
  columnHeaderFontWeightClass: "font-normal",
  defaultFontWeightClass: "font-normal",
  table: "DimensionTable",
};
