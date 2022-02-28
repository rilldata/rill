import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RootConfig } from "$common/config/RootConfig";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";

/**
 * Abstract class to handle entity state updates.
 * Any new entity, updated entity or deleted entity will trigger respective method by {@link EntityStateSyncService}
 */
export class EntityStateUpdatesHandler<Entity extends EntityRecord> {
    public constructor(protected readonly config: RootConfig,
                       protected readonly dataModelerService: DataModelerService) {}

    public async handleEntityInit(entity: Entity): Promise<void> {}
    public async handleNewEntity(entity: Entity): Promise<void> {}
    public async handleUpdatedEntity(entity: Entity): Promise<void> {}
    public async handleDeletedEntity(entity: Entity): Promise<void> {}
}
