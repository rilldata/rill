import type { VirtualizedTableConfig } from "@rilldata/web-common/components/virtualized-table/types";
import {
  COMPARISON_COLUMN_WIDTH,
  DEFAULT_COLUMN_WIDTH,
} from "../leaderboard/leaderboard-widths";

export type DimensionTableConfig = VirtualizedTableConfig & {
  comparisonColumnWidth: number;
};

export const DIMENSION_TABLE_CONFIG: DimensionTableConfig = {
  defaultColumnWidth: DEFAULT_COLUMN_WIDTH,
  maxColumnWidth: 320,
  minColumnWidth: 104,
  minHeaderWidthWhenColumsAreSmall: 160,
  comparisonColumnWidth: COMPARISON_COLUMN_WIDTH,
  rowHeight: 24,
  columnHeaderHeight: 28,
  indexWidth: 24,
  columnHeaderFontWeightClass: "font-normal",
  defaultFontWeightClass: "font-normal",
  table: "DimensionTable",
  headerBgColorClass: "bg-surface-background",
  headerBgColorHighlightClass: "bg-surface-hover",
};
