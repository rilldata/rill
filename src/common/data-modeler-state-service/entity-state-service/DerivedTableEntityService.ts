import type {
    EntityState, EntityStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
    EntityStateService, EntityType, StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { DataProfileEntity } from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";

export interface DerivedTableEntity extends DataProfileEntity {
    type: EntityType.Table;
}
export type DerivedTableState = EntityState<DerivedTableEntity>;
export type DerivedTableStateActionArg = EntityStateActionArg<
    DerivedTableEntity, DerivedTableState, DerivedTableEntityService>;

export class DerivedTableEntityService extends EntityStateService<DerivedTableEntity, DerivedTableState> {
    public readonly entityType = EntityType.Table;
    public readonly stateType = StateType.Derived;
}
