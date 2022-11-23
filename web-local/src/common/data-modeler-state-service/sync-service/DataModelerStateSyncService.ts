import type { RootConfig } from "../../config/RootConfig";
import type {
  EntityRecord,
  EntityStateService,
} from "../entity-state-service/EntityStateService";
import {
  EntityType,
  StateType,
} from "../entity-state-service/EntityStateService";
import type { DataModelerService } from "../../data-modeler-service/DataModelerService";
import type { DataModelerStateService } from "../DataModelerStateService";
import { EntityStateSyncService } from "./EntityStateSyncService";
import { EntityRepository } from "./EntityRepository";
import { DerivedEntityUpdateHandler } from "./DerivedEntityUpdateHandler";
import { EntityStateUpdatesHandler } from "./EntityStateUpdatesHandler";
import { PersistentModelRepository } from "./PersistentModelRepository";
import { PersistentModelUpdateHandler } from "./PersistentModelUpdateHandler";

/**
 * A single interface to start and stop all entity state sync services
 */
export class DataModelerStateSyncService {
  private readonly entityStateSyncServices: Array<
    EntityStateSyncService<EntityRecord, EntityStateService<EntityRecord>>
  >;

  public constructor(
    config: RootConfig,
    entityStateServices: Array<EntityStateService<EntityRecord>>,
    dataModelerService: DataModelerService,
    dataModelerStateService: DataModelerStateService
  ) {
    this.entityStateSyncServices = entityStateServices.map(
      (entityStateService) => {
        return new EntityStateSyncService(
          config,
          DataModelerStateSyncService.getEntityRepository(
            config,
            dataModelerService,
            entityStateService.entityType,
            entityStateService.stateType
          ),
          DataModelerStateSyncService.getEntityStateUpdatesHandler(
            config,
            dataModelerService,
            entityStateService.entityType,
            entityStateService.stateType
          ),
          dataModelerStateService,
          entityStateService
        );
      }
    );
  }

  private static getEntityRepository(
    config: RootConfig,
    dataModelerService: DataModelerService,
    entityType: EntityType,
    stateType: StateType
  ): EntityRepository<EntityRecord> {
    return new EntityRepository(
      config.state,
      dataModelerService,
      entityType,
      stateType
    );
  }

  private static getEntityStateUpdatesHandler(
    config: RootConfig,
    dataModelerService: DataModelerService,
    entityType: EntityType,
    stateType: StateType
  ) {
    if (
      stateType === StateType.Derived &&
      (entityType === EntityType.Model || entityType === EntityType.Table)
    ) {
      return new DerivedEntityUpdateHandler(
        config,
        dataModelerService,
        entityType === EntityType.Model
          ? "collectModelInfo"
          : "collectTableInfo"
      );
    } else if (
      stateType === StateType.Persistent &&
      entityType === EntityType.Model
    ) {
      return new PersistentModelUpdateHandler(config, dataModelerService);
    }
    return new EntityStateUpdatesHandler(config, dataModelerService);
  }

  public async init(): Promise<void> {
    await Promise.all(
      this.entityStateSyncServices.map((entityStateSyncService) =>
        entityStateSyncService.init()
      )
    );
  }

  public async destroy(): Promise<void> {
    await Promise.all(
      this.entityStateSyncServices.map((entityStateSyncService) =>
        entityStateSyncService.destroy()
      )
    );
  }
}
