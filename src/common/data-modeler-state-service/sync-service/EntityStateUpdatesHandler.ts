import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RootConfig } from "$common/config/RootConfig";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";

/**
 * Abstract class to handle entity state updates.
 * Any new entity, updated entity or deleted entity will trigger respective method by {@link EntityStateSyncService}
 */
export class EntityStateUpdatesHandler<Entity extends EntityRecord> {
  public constructor(
    protected readonly config: RootConfig,
    protected readonly dataModelerService: DataModelerService
  ) {}

  public async handleEntityInit(_entity: Entity): Promise<void> {
    // FIXME add descriptive comment describing why this empy method is needed
  }
  public async handleNewEntity(_entity: Entity): Promise<void> {
    // FIXME add descriptive comment describing why this empy method is needed
  }
  public async handleUpdatedEntity(_entity: Entity): Promise<void> {
    // FIXME add descriptive comment describing why this empy method is needed
  }
}
