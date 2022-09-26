import type {
  EntityRecord,
  EntityState,
  EntityStateActionArg,
} from "./EntityStateService";
import {
  EntityStateService,
  EntityType,
  StateType,
} from "./EntityStateService";
import type { RollupInterval } from "../../database-service/DatabaseColumnActions";

export enum ValidationState {
  OK = "OK",
  WARNING = "WARNING",
  ERROR = "ERROR",
}

export enum SourceModelValidationStatus {
  OK = "OK",
  // No source model selected.
  EMPTY = "EMPTY",
  // Source model query is invalid.
  INVALID = "INVALID",
  // Selected source model is no longer present.
  MISSING = "MISSING",
}

export interface MetricsDefinitionEntity extends EntityRecord {
  type: EntityType.MetricsDefinition;
  metricDefLabel: string;
  sourceModelId: string | undefined;
  sourceModelValidationStatus?: SourceModelValidationStatus;
  timeDimension: string | undefined;
  // We can reuse SourceModelStatus as everything there applies here as well.
  // EMPTY => no time dimension selected
  // INVALID => source model query is invalid. will apply once some time dimension was selected
  // MISSING => selected time dimension is no longer present.
  timeDimensionValidationStatus?: SourceModelValidationStatus;
  creationTime: number;
  rollupInterval?: RollupInterval;
  measureIds: Array<string>;
  dimensionIds: Array<string>;
  summaryExpandedInNav?: boolean;
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
