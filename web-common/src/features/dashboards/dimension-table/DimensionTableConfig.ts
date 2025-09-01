import type { VirtualizedTableConfig } from "@rilldata/web-common/components/virtualized-table/types";

export type DimensionTableConfig = VirtualizedTableConfig & {
  comparisonColumnWidth: number;
};

export const DIMENSION_TABLE_CONFIG: DimensionTableConfig = {
  defaultColumnWidth: 110,
  maxColumnWidth: 320,
  minColumnWidth: 104,
  minHeaderWidthWhenColumsAreSmall: 160,
  comparisonColumnWidth: 64,
  rowHeight: 24,
  columnHeaderHeight: 28,
  indexWidth: 24,
  columnHeaderFontWeightClass: "font-normal",
  defaultFontWeightClass: "font-normal",
  table: "DimensionTable",
  headerBgColorClass: "bg-surface",
  headerBgColorHighlightClass: "bg-gray-50",
};
