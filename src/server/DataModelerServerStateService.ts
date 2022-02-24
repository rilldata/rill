import { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { EntityStateSyncService } from "$common/data-modeler-state-service/sync-service/EntityStateSyncService";

/**
 * State service class present on the server.
 * Reads the initial state from file and syncs it back.
 */
export class DataModelerServerStateService extends DataModelerStateService {
    private entityStateSyncServices = new Array<EntityStateSyncService<any, any>>();

    public async init(): Promise<void> {
        await Promise.all(this.entityStateServices.map(async (entityStateService) => {
            const entityStateSyncService = new EntityStateSyncService(
                this.config, entityStateService.entityType,
                entityStateService.stateType, this, entityStateService);
            this.entityStateSyncServices.push(entityStateSyncService);
            await entityStateSyncService.init();
        }));
    }

    public async destroy(): Promise<void> {
        await Promise.all(this.entityStateSyncServices.map(
            entityStateSyncService => entityStateSyncService.destroy()));
    }
}
