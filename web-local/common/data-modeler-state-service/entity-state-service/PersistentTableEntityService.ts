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
import type { TableSourceType } from "../../../lib/types";

export interface PersistentTableEntity extends EntityRecord {
  type: EntityType.Table;
  path: string;
  name?: string;
  // we have a separate field to maintain different names in the future.
  // currently, name = tableName
  tableName?: string;

  sourceType?: TableSourceType;
  csvDelimiter?: string;
}
export type PersistentTableState = EntityState<PersistentTableEntity>;
export type PersistentTableStateActionArg = EntityStateActionArg<
  PersistentTableEntity,
  PersistentTableState,
  PersistentTableEntityService
>;

export class PersistentTableEntityService extends EntityStateService<
  PersistentTableEntity,
  PersistentTableState
> {
  public readonly entityType = EntityType.Table;
  public readonly stateType = StateType.Persistent;
}
