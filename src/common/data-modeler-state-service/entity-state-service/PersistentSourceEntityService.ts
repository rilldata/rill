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
import type { SourceType } from "$lib/types";

export interface PersistentSourceEntity extends EntityRecord {
  type: EntityType.Source;
  path: string;
  name?: string;
  // we have a separate field to maintain different names in the future.
  // currently, name = sourceName
  sourceName?: string;

  sourceType?: SourceType;
  csvDelimiter?: string;
}
export type PersistentSourceState = EntityState<PersistentSourceEntity>;
export type PersistentSourceStateActionArg = EntityStateActionArg<
  PersistentSourceEntity,
  PersistentSourceState,
  PersistentSourceEntityService
>;

export class PersistentSourceEntityService extends EntityStateService<
  PersistentSourceEntity,
  PersistentSourceState
> {
  public readonly entityType = EntityType.Source;
  public readonly stateType = StateType.Persistent;
}
