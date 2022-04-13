import { StateActions } from "$common/data-modeler-state-service/StateActions";
import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import type {
    ApplicationStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { ApplicationStatus } from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";

export class ApplicationStateActions extends StateActions {
    @DataModelerActions.ApplicationAction()
    public async setActiveAsset({draftState}: ApplicationStateActionArg,
                                entityType: EntityType, entityId: string) {
        draftState.activeEntity = {
            type: entityType, id: entityId
        };
    }

    @DataModelerActions.ApplicationAction()
    public async setApplicationStatus({draftState}: ApplicationStateActionArg,
                                      status: ApplicationStatus) {
        draftState.status = status;
    }

    @DataModelerActions.ApplicationAction()
    public async setDuckDbPath(
        {draftState}: ApplicationStateActionArg,
        duckDbPath: string,
    ) {
        draftState.duckDbPath = duckDbPath;
    }
}
