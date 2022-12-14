import type { VirtualizedTableConfig } from "../virtualized-table/types";

export const config: VirtualizedTableConfig = {
  defaultColumnWidth: 200,
  maxColumnWidth: 320,
  minColumnWidth: 120,
  minHeaderWidthWhenColumsAreSmall: 160,
  rowHeight: 36,
  columnHeaderHeight: 32,
  indexWidth: 60,
  columnHeaderFontWeightClass: "ui-copy-strong",
  defaultFontWeightClass: "ui-copy",
  table: "PreviewTable",
};
