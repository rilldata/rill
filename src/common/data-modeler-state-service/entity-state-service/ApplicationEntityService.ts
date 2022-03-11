import type {
    DerivedEntityRecord, EntityState
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    EntityStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
    EntityStateService, EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export interface ActiveEntity {
    type: EntityType,
    id: string;
}

export interface ApplicationEntity extends DerivedEntityRecord {}
export interface ApplicationState extends EntityState<ApplicationEntity> {
    activeEntity: ActiveEntity;
    databasePaused: boolean;
}
export type ApplicationStateActionArg = EntityStateActionArg<ApplicationEntity, ApplicationState>;

export class ApplicationStateService extends EntityStateService<ApplicationEntity, ApplicationState> {
    public readonly entityType = EntityType.Application;
    public readonly stateType = StateType.Derived;
}
