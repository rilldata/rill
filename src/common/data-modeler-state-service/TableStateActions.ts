import {StateActions} from ".//StateActions";
import type {
    PersistentTableEntity,
    PersistentTableStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { DataProfileStateActionArg } from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import type { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class TableStateActions extends StateActions {
    @StateActions.PersistentTableAction()
    public addOrUpdateTableToState({stateService, draftState}: PersistentTableStateActionArg,
                                   table: PersistentTableEntity, isNew: boolean): void {
        if (isNew) {
            stateService.addEntity(draftState, table);
        } else {
            stateService.updateEntity(draftState, table.id, table);
        }
    }
}
