import type {
  EntityState,
  EntityStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  EntityStateService,
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { DataProfileEntity } from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import type { Source } from "$lib/types";

export interface DerivedModelEntity extends DataProfileEntity {
  type: EntityType.Model;
  /** sanitizedQuery is always a 1:1 function of the query itself */
  sanitizedQuery: string;
  error?: string;
  sources?: Source[];
}
export type DerivedModelState = EntityState<DerivedModelEntity>;
export type DerivedModelStateActionArg = EntityStateActionArg<
  DerivedModelEntity,
  DerivedModelState,
  DerivedModelEntityService
>;

export class DerivedModelEntityService extends EntityStateService<
  DerivedModelEntity,
  DerivedModelState
> {
  public readonly entityType = EntityType.Model;
  public readonly stateType = StateType.Derived;
}
