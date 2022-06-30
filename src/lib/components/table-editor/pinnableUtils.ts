import type { ColumnConfig } from "$lib/components/table/ColumnConfig";

export function columnIsPinned(name, selectedColumns: Array<ColumnConfig>) {
  return selectedColumns.map((column) => column.name).includes(name);
}

export function togglePin(columnConfig: ColumnConfig, selectedColumns) {
  // if column is already pinned, remove.
  if (columnIsPinned(columnConfig.name, selectedColumns)) {
    return [
      ...selectedColumns.filter((column) => column.name !== columnConfig.name),
    ];
  } else {
    return [...selectedColumns, columnConfig];
  }
}
