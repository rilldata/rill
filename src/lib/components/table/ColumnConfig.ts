import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

export interface ColumnConfig {
  name: string;
  label?: string;
  type: string;

  renderer?: unknown;

  validation?: (row: EntityRecord, value: unknown) => ValidationState;

  copyable?: boolean;
}
