export interface ColumnName {
  name: string;
  type: string;
  cellType?: string;
  cellComponent?: unknown;
}

export function columnIsPinned(name, selectedColumns: Array<ColumnName>) {
  return selectedColumns.map((column) => column.name).includes(name);
}

export function togglePin(name, type, selectedColumns) {
  // if column is already pinned, remove.
  if (columnIsPinned(name, selectedColumns)) {
    return [...selectedColumns.filter((column) => column.name !== name)];
  } else {
    return [...selectedColumns, { name, type }];
  }
}
