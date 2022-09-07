export type HeaderPosition = "top" | "left" | "top-left";

export interface VirtualizedTableConfig {
  defaultColumnWidth: number;
  maxColumnWidth: number;
  minColumnWidth: number;
  minHeaderWidthWhenColumsAreSmall: number;
  rowHeight: number;
  indexWidth: number;
}
