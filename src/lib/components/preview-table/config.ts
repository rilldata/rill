import type { VirtualizedTableConfig } from "$lib/components/virtualized-table/types";

export const config: VirtualizedTableConfig = {
  defaultColumnWidth: 200,
  maxColumnWidth: 320,
  minColumnWidth: 120,
  minHeaderWidthWhenColumsAreSmall: 160,
  rowHeight: 36,
  indexWidth: 60,
};
