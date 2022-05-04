import {EntityStateUpdatesHandler} from "$common/data-modeler-state-service/sync-service/EntityStateUpdatesHandler";
import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";

export class PersistentModelUpdateHandler extends EntityStateUpdatesHandler<PersistentModelEntity> {
    public async handleUpdatedEntity(modelEntity: PersistentModelEntity): Promise<void> {
        await this.dataModelerService.dispatch(
            "updateModelQuery", [modelEntity.id, modelEntity.query]);
    }
}
