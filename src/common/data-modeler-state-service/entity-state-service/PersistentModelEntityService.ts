import type {
    EntityRecord, EntityState, EntityStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
    EntityStateService, EntityType, StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export interface PersistentModelEntity extends EntityRecord {
    type: EntityType.Model;
    query: string;
    /** name is used for the filename and exported file */
    name: string;
    tableName?: string;
}
export interface PersistentModelState extends EntityState<PersistentModelEntity> {
    modelNumber: number;
}
export type PersistentModelStateActionArg = EntityStateActionArg<
    PersistentModelEntity, PersistentModelState, PersistentModelEntityService>;

export class PersistentModelEntityService extends EntityStateService<PersistentModelEntity, PersistentModelState> {
    public readonly entityType = EntityType.Model;
    public readonly stateType = StateType.Persistent;

    public init(initialState: PersistentModelState): void {
        if (!("modelNumber" in initialState)) {
            initialState.modelNumber = 0;
        }
        initialState.entities.forEach(entity => {
            const match = entity.name.match(/query_(\d*).sql/);
            const num = Number(match?.[1]);
            if (!Number.isNaN(num)) {
                initialState.modelNumber =
                    Math.max(initialState.modelNumber, Number(num));
            }
        });
        super.init(initialState);
    }
}
