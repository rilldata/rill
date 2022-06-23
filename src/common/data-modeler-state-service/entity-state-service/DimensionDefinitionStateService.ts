import type { ProfileColumnSummary } from "$lib/types";
import type { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import {
  EntityRecord,
  EntityState,
  EntityStateActionArg,
  EntityStateService,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export interface DimensionDefinitionEntity extends EntityRecord {
  metricsDefId: string;
  // mandatory user defined data
  dimensionColumn: string;
  // optional user defined metadata
  sqlName?: string;
  nameSingle?: string;
  namePlural?: string;
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
