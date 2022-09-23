import type {
  EntityState,
  EntityStateActionArg,
} from "./EntityStateService";
import {
  EntityStateService,
  EntityType,
  StateType,
} from "./EntityStateService";
import type { DataProfileEntity } from "./DataProfileEntity";

export interface DerivedTableEntity extends DataProfileEntity {
  type: EntityType.Table;
}
export type DerivedTableState = EntityState<DerivedTableEntity>;
export type DerivedTableStateActionArg = EntityStateActionArg<
  DerivedTableEntity,
  DerivedTableState,
  DerivedTableEntityService
>;

export class DerivedTableEntityService extends EntityStateService<
  DerivedTableEntity,
  DerivedTableState
> {
  public readonly entityType = EntityType.Table;
  public readonly stateType = StateType.Derived;
}
