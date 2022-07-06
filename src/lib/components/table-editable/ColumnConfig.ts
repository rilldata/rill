import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

/**
 * config info for table columns
 *
 * name: the property name used in an EntityRecord
 * label?: label used for display in table header (`name` is used if not provided)
 * tooltip: tooltip when hovering over column header
 */
export interface ColumnConfig {
  name: string;
  label?: string;
  type?: string;

  renderer?: unknown;
  tooltip?: string;

  validation?: (row: EntityRecord, value: unknown) => ValidationState;

  copyable?: boolean;
}
