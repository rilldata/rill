import type { ColumnConfig, CellConfig } from "./ColumnConfig";

export function columnIsPinned(
  name,
  selectedColumns: Array<ColumnConfig<CellConfig>>
) {
  return selectedColumns.map((column) => column.name).includes(name);
}

export function togglePin(columnConfig: ColumnConfig<any>, selectedColumns) {
  // if column is already pinned, remove.
  if (columnIsPinned(columnConfig.name, selectedColumns)) {
    return [
      ...selectedColumns.filter((column) => column.name !== columnConfig.name),
    ];
  } else {
    return [...selectedColumns, columnConfig];
  }
}
