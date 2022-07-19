import type {
  EntityRecord,
  EntityState,
  EntityStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  EntityStateService,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";

export enum ValidationState {
  OK = "OK",
  WARNING = "WARNING",
  ERROR = "ERROR",
}

export interface MetricsDefinitionEntity extends EntityRecord {
  type: EntityType.MetricsDefinition;
  metricDefLabel: string;
  sourceModelId: string | undefined;
  timeDimension: string | undefined;
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
