import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  EntityStateService,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  EntityState,
  EntityStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";
import type { ProfileColumnSummary } from "$lib/types";

// or whatever we're usng for ids
export type UUID = string;

export enum ValidationState {
  OK = "OK",
  WARNING = "WARNING",
  ERROR = "ERROR",
}

export type MeasureDefinition = {
  // mandatory user defined metadata
  expression: string;
  // optional user defined metadata
  sqlName?: string;
  label?: string;
  description?: string;
  // internal state for rendering etc
  id: UUID;
  expressionIsValid: ValidationState;
  sqlNameIsValid: ValidationState;
  sparkLineId: UUID;
};

export type DimensionDefinition = {
  // mandatory user defined data
  dimensionColumn: string;
  // optional user defined metadata
  sqlName?: string;
  nameSingle?: string;
  namePlural?: string;
  description?: string;
  // internal state for rendering etc
  id: UUID;
  dimensionIsValid: ValidationState;
  sqlNameIsValid: ValidationState;
  summary?: ProfileColumnSummary;
};

export interface MetricsDefinitionEntity extends EntityRecord {
  type: EntityType.MetricsDefinition;
  metricDefLabel: string;
  sourceModelId: UUID | undefined;
  timeDimension: string | undefined;
  rollupInterval?: RollupInterval;
  measures: MeasureDefinition[];
  dimensions: DimensionDefinition[];
}

export interface MetricsDefinitionState
  extends EntityState<MetricsDefinitionEntity> {
  counter: number;
}

export type MetricsDefinitionStateActionArg = EntityStateActionArg<
  MetricsDefinitionEntity,
  MetricsDefinitionState,
  MetricsDefinitionStateService
>;

export class MetricsDefinitionStateService extends EntityStateService<
  MetricsDefinitionEntity,
  MetricsDefinitionState
> {
  public readonly entityType = EntityType.MetricsDefinition;
  public readonly stateType = StateType.Persistent;

  public getEmptyState(): MetricsDefinitionState {
    return {
      lastUpdated: 0,
      entities: [],
      counter: 0,
    } as MetricsDefinitionState;
  }
}
