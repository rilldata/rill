import type { VirtualizedTableConfig } from "@rilldata/web-common/components/virtualized-table/types";

export const DimensionTableConfig: VirtualizedTableConfig = {
  defaultColumnWidth: 110,
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
