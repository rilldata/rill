import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import type {
    ApplicationState,
    ApplicationStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";

export class CommonActions extends  DataModelerActions {
    @DataModelerActions.ApplicationAction()
    public async setActiveAsset({stateService}: ApplicationStateActionArg,
                                entityType: EntityType, entityId: string) {
        const currentActiveAsset =
            (stateService.getCurrentState() as ApplicationState).activeEntity;
        // mark older model as inactive.
        if (currentActiveAsset?.type === EntityType.Model) {
            this.databaseActionQueue.updatePriority(currentActiveAsset.id,
                DatabaseActionQueuePriority.InactiveModelProfile);
        }
        this.dataModelerStateService.dispatch("setActiveAsset", [entityType, entityId]);
    }
}
