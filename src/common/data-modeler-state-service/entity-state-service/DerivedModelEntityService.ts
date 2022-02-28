import {
    EntityState, EntityStateActionArg,
    EntityStateService, EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { DataProfileEntity } from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import type { SourceTable } from "$lib/types";

export interface DerivedModelEntity extends DataProfileEntity {
    type: EntityType.Model;
    /** sanitizedQuery is always a 1:1 function of the query itself */
    sanitizedQuery: string;
    error?: string;
    sources?: SourceTable[];
}
export type DerivedModelState = EntityState<DerivedModelEntity>;
export type DerivedModelStateActionArg = EntityStateActionArg<DerivedModelEntity>;

export class DerivedModelEntityService extends EntityStateService<DerivedModelEntity> {
    public readonly entityType = EntityType.Model;
    public readonly stateType = StateType.Derived;
}
