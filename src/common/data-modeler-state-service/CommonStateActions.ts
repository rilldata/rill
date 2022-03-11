import { StateActions } from "$common/data-modeler-state-service/StateActions";
import type {
    EntityRecord, EntityStateActionArg,
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class CommonStateActions extends StateActions {
    @StateActions.GenericAction()
    public addEntity({stateService, draftState}: EntityStateActionArg<any>,
                     entityType: EntityType, stateType: StateType,
                     entity: EntityRecord, atIndex?: number): void {
        stateService.addEntity(draftState, entity, atIndex);
    }

    @StateActions.GenericAction()
    public updateEntity({stateService, draftState}: EntityStateActionArg<any>,
                        entityType: EntityType, stateType: StateType,
                        entity: EntityRecord): void {
        stateService.updateEntity(draftState, entity.id, entity);
    }

    @StateActions.GenericAction()
    public deleteEntity({stateService, draftState}: EntityStateActionArg<any>,
                        entityType: EntityType, stateType: StateType,
                        entityId: string): void {
        stateService.deleteEntity(draftState, entityId);
    }

    @StateActions.GenericAction()
    public moveEntityDown({stateService, draftState}: EntityStateActionArg<any>,
                          entityType: EntityType, stateType: StateType,
                          entityId: string): void {
        stateService.moveEntityDown(draftState, entityId);
    }

    @StateActions.GenericAction()
    public moveEntityUp({stateService, draftState}: EntityStateActionArg<any>,
                        entityType: EntityType, stateType: StateType,
                        entityId: string): void {
        stateService.moveEntityUp(draftState, entityId);
    }
}
