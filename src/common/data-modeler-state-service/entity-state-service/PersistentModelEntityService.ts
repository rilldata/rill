import {
    EntityRecord, EntityState, EntityStateActionArg,
    EntityStateService, EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export interface PersistentModelEntity extends EntityRecord {
    type: EntityType.Model;
    query: string;
    /** name is used for the filename and exported file */
    name: string;
    tableName?: string;
}
export type PersistentModelState = EntityState<PersistentModelEntity>;
export type PersistentModelStateActionArg = EntityStateActionArg<PersistentModelEntity>;

export class PersistentModelEntityService extends EntityStateService<PersistentModelEntity> {
    public readonly entityType = EntityType.Model;
    public readonly stateType = StateType.Persistent;
}
