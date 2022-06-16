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

export interface DerivedSourceEntity extends DataProfileEntity {
  type: EntityType.Source;
}
export type DerivedSourceState = EntityState<DerivedSourceEntity>;
export type DerivedSourceStateActionArg = EntityStateActionArg<
  DerivedSourceEntity,
  DerivedSourceState,
  DerivedSourceEntityService
>;

export class DerivedSourceEntityService extends EntityStateService<
  DerivedSourceEntity,
  DerivedSourceState
> {
  public readonly entityType = EntityType.Source;
  public readonly stateType = StateType.Derived;
}
