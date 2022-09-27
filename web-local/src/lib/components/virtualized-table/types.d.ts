export type HeaderPosition = "top" | "left" | "top-left";

export type PinnedColumnSide = "left" | "right";

export interface VirtualizedTableConfig {
  defaultColumnWidth: number;
  maxColumnWidth: number;
  minColumnWidth: number;
  minHeaderWidthWhenColumsAreSmall: number;
  rowHeight: number;
  indexWidth: number;
  columnHeaderHeight: number;
  columnHeaderFontWeightClass: string;
  defaultFontWeightClass: string;
  table: string;
}
