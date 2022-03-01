import { DataModelerActions } from "$common/data-modeler-service/DataModelerActions";
import type {
    ApplicationStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class CommonActions extends  DataModelerActions {
    @DataModelerActions.ApplicationAction()
    public async setActiveAsset(args: ApplicationStateActionArg,
                                entityType: EntityType, entityId: string) {
        this.dataModelerStateService.dispatch("setActiveAsset", [entityType, entityId]);
    }
}
