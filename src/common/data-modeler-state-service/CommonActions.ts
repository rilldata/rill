import { StateActions } from "$common/data-modeler-state-service/StateActions";
import type {
    EntityRecord, EntityStateActionArg,
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class CommonActions extends StateActions {
    @StateActions.GenericAction()
    public addEntity({stateService, draftState}: EntityStateActionArg<any>,
                     entityType: EntityType, stateType: StateType,
                     entity: EntityRecord): void {
        stateService.addEntity(draftState, entity);
    }

    @StateActions.GenericAction()
    public updateEntity({stateService, draftState}: EntityStateActionArg<any>,
                        entityType: EntityType, stateType: StateType,
                        entity: EntityRecord): void {
        stateService.updateEntity(draftState, entity.id, entity);
    }
}
