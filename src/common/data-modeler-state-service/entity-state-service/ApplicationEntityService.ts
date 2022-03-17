import type {
    DerivedEntityRecord,
    EntityState,
    EntityStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
    EntityStateService,
    EntityStatus,
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export interface ActiveEntity {
    type: EntityType,
    id: string;
}

export interface ApplicationEntity extends DerivedEntityRecord {}
export interface ApplicationState extends EntityState<ApplicationEntity> {
    activeEntity?: ActiveEntity;
    status: EntityStatus;
}
export type ApplicationStateActionArg = EntityStateActionArg<ApplicationEntity, ApplicationState>;

export class ApplicationStateService extends EntityStateService<ApplicationEntity, ApplicationState> {
    public readonly entityType = EntityType.Application;
    public readonly stateType = StateType.Derived;

    public getEmptyState(): ApplicationState {
        return {lastUpdated: 0, entities: [], status: EntityStatus.Idle};
    }
}
