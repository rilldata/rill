import type {RootConfig} from "$common/config/RootConfig";
import type {EntityStateService} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {EntityType, StateType} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import type {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import {EntityStateSyncService} from "$common/data-modeler-state-service/sync-service/EntityStateSyncService";
import {EntityRepository} from "$common/data-modeler-state-service/sync-service/EntityRepository";
import {DerivedEntityUpdateHandler} from "$common/data-modeler-state-service/sync-service/DerivedEntityUpdateHandler";
import {EntityStateUpdatesHandler} from "$common/data-modeler-state-service/sync-service/EntityStateUpdatesHandler";
import {PersistentModelRepository} from "$common/data-modeler-state-service/sync-service/PersistentModelRepository";
import {
    PersistentModelUpdateHandler
} from "$common/data-modeler-state-service/sync-service/PersistentModelUpdateHandler";

/**
 * A single interface to start and stop all entity state sync services
 */
export class DataModelerStateSyncService {
    private readonly entityStateSyncServices: Array<EntityStateSyncService<any, any>>;

    public constructor(config: RootConfig, entityStateServices: Array<EntityStateService<any>>,
                       dataModelerService: DataModelerService,
                       dataModelerStateService: DataModelerStateService) {
        this.entityStateSyncServices = entityStateServices.map((entityStateService) => {
            return new EntityStateSyncService(
                config,
                DataModelerStateSyncService.getEntityRepository(config, dataModelerService,
                    entityStateService.entityType, entityStateService.stateType),
                DataModelerStateSyncService.getEntityStateUpdatesHandler(config, dataModelerService,
                    entityStateService.entityType, entityStateService.stateType),
                dataModelerStateService, entityStateService);
        });
    }

    public async init(): Promise<void> {
        await Promise.all(this.entityStateSyncServices.map(
            entityStateSyncService => entityStateSyncService.init()));
    }

    public async destroy(): Promise<void> {
        await Promise.all(this.entityStateSyncServices.map(
            entityStateSyncService => entityStateSyncService.destroy()));
    }

    private static getEntityRepository(
        config: RootConfig, dataModelerService: DataModelerService,
        entityType: EntityType, stateType: StateType,
    ): EntityRepository<any> {
        if (entityType === EntityType.Model && stateType === StateType.Persistent) {
            return new PersistentModelRepository(config.state, dataModelerService, entityType, stateType);
        }
        return new EntityRepository(config.state, dataModelerService, entityType, stateType);
    }

    private static getEntityStateUpdatesHandler(
        config: RootConfig, dataModelerService: DataModelerService,
        entityType: EntityType, stateType: StateType,
    ) {
        if (stateType === StateType.Derived &&
            (entityType === EntityType.Model || entityType === EntityType.Table)) {

            return new DerivedEntityUpdateHandler(config, dataModelerService,
                entityType === EntityType.Model ? "collectModelInfo": "collectTableInfo")
        } else if (stateType === StateType.Persistent && entityType === EntityType.Model) {
            return new PersistentModelUpdateHandler(config, dataModelerService);
        }
        return new EntityStateUpdatesHandler(config, dataModelerService);
    }
}
