import { EntityStateUpdatesHandler } from "$common/data-modeler-state-service/sync-service/EntityStateUpdatesHandler";
import type { RootConfig } from "$common/config/RootConfig";
import type { DataModelerActionsDefinition, DataModelerService } from "$common/data-modeler-service/DataModelerService";
import type { DataProfileEntity } from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";

/**
 * Update handler that triggers action to profile if not already profiled.
 */
export class DerivedEntityUpdateHandler extends EntityStateUpdatesHandler<DataProfileEntity> {
    public constructor(protected readonly config: RootConfig,
                       protected readonly dataModelerService: DataModelerService,
                       protected readonly collectEntityInfoAction: keyof DataModelerActionsDefinition) {
        super(config, dataModelerService);
    }

    public async handleEntityInit(entity: DataProfileEntity): Promise<void> {
        return this.handleModelProfiling(entity);
    }

    public async handleNewEntity(entity: DataProfileEntity): Promise<void> {
        return this.handleModelProfiling(entity);
    }

    public async handleUpdatedEntity(entity: DataProfileEntity): Promise<void> {
        return this.handleModelProfiling(entity);
    }

    private async handleModelProfiling(entity: DataProfileEntity): Promise<void> {
        if (!entity.profiled) {
            // make sure to run it after a little delay
            // we need entry in both derived and persistent states
            // TODO: Find a better way to sync this
            setTimeout(() => {
                this.dataModelerService.dispatch(this.collectEntityInfoAction, [entity.id]);
            }, this.config.state.syncInterval * 2);
        }
    }
}
