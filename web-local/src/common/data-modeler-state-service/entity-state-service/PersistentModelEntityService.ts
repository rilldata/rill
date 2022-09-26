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

export interface PersistentModelEntity extends EntityRecord {
  type: EntityType.Model;
  query: string;
  /** name is used for the filename and exported file */
  name: string;
  tableName?: string;
}
export interface PersistentModelState
  extends EntityState<PersistentModelEntity> {
  modelNumber: number;
}
export type PersistentModelStateActionArg = EntityStateActionArg<
  PersistentModelEntity,
  PersistentModelState,
  PersistentModelEntityService
>;

export class PersistentModelEntityService extends EntityStateService<
  PersistentModelEntity,
  PersistentModelState
> {
  public readonly entityType = EntityType.Model;
  public readonly stateType = StateType.Persistent;

  public init(initialState: PersistentModelState): void {
    if (!("modelNumber" in initialState)) {
      initialState.modelNumber = 0;
    }
    initialState.entities.forEach((entity) => {
      const match = entity.name.match(/model_(\d*).sql/);
      const num = Number(match?.[1]);
      if (!Number.isNaN(num)) {
        initialState.modelNumber = Math.max(
          initialState.modelNumber,
          Number(num)
        );
      }
    });
    super.init(initialState);
  }
}
