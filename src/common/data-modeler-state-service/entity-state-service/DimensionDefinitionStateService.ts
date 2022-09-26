import {
  EntityRecord,
  EntityState,
  EntityStateActionArg,
  EntityStateService,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { ProfileColumnSummary } from "$lib/types";

export interface DimensionDefinitionEntity extends EntityRecord {
  metricsDefId: string;
  creationTime: number;
  // mandatory user defined data
  dimensionColumn: string;
  // optional user defined metadata
  sqlName?: string;
  labelSingle?: string;
  labelPlural?: string;
  description?: string;
  dimensionIsValid?: ValidationState;
  sqlNameIsValid?: ValidationState;
  summary?: ProfileColumnSummary;
}

export type DimensionDefinitionState = EntityState<DimensionDefinitionEntity>;

export type DimensionDefinitionStateActionArg = EntityStateActionArg<
  DimensionDefinitionEntity,
  DimensionDefinitionState,
  DimensionDefinitionStateService
>;

export class DimensionDefinitionStateService extends EntityStateService<
  DimensionDefinitionEntity,
  DimensionDefinitionState
> {
  public readonly entityType = EntityType.DimensionDefinition;
  public readonly stateType = StateType.Persistent;
}

// we need a fallback for dimension name. this is needed when sqlName is not entered.
export function getFallbackDimensionName(index: number, sqlName?: string) {
  return sqlName?.length ? sqlName : `dimension_${index}`;
}
