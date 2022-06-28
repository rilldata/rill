import type { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export interface ColumnConfig {
  name: string;
  label?: string;
  type: string;

  renderer?: unknown;

  validation?: (row: EntityRecord, value: unknown) => ValidationState;

  copyable?: boolean;
}

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
