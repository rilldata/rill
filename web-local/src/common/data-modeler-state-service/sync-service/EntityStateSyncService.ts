import type {
  EntityRecord,
  EntityState,
} from "../entity-state-service/EntityStateService";
import type { RootConfig } from "../../config/RootConfig";
import type { EntityStateService } from "../entity-state-service/EntityStateService";
import type { DataModelerStateService } from "../DataModelerStateService";
import type { EntityRepository } from "./EntityRepository";
import type { EntityStateUpdatesHandler } from "./EntityStateUpdatesHandler";
import { mkdirSync } from "fs";

/**
 * This class periodically checks source and compares it with in-memory state.
 * Any changes are written to in-memory state and vice-versa.
 */
export class EntityStateSyncService<
  Entity extends EntityRecord,
  StateService extends EntityStateService<Entity>
> {
  private syncTimer: NodeJS.Timer;

  public constructor(
    private readonly config: RootConfig,
    private readonly entityRepository: EntityRepository<Entity>,
    private readonly entityStateUpdatesHandler: EntityStateUpdatesHandler<Entity>,
    private readonly dataModelerStateService: DataModelerStateService,
    private readonly entityStateService: StateService
  ) {}

  public async init(): Promise<void> {
    mkdirSync(this.config.state.stateFolder, { recursive: true });

    let initialState: EntityState<Entity>;

    if (
      this.config.state.autoSync &&
      (await this.entityRepository.sourceExists())
    ) {
      try {
        initialState = await this.entityRepository.getAll();
      } catch (err) {
        initialState = this.entityStateService.getEmptyState();
      }
    } else {
      initialState = this.entityStateService.getEmptyState();
    }
    this.entityStateService.init(initialState);
    if (!this.config.state.autoSync) return;

    initialState.entities.forEach((entity) =>
      this.entityStateUpdatesHandler.handleEntityInit(entity)
    );

    this.syncTimer = setInterval(() => {
      this.sync();
    }, this.config.state.syncInterval);
  }

  public async destroy(): Promise<void> {
    if (this.syncTimer) {
      clearInterval(this.syncTimer);
    }
    await this.sync(true);
  }

  private async sync(writeOnly = false): Promise<void> {
    if (!(await this.entityRepository.sourceExists())) {
      await this.syncCurrentWithSource();
    }

    let sourceState: EntityState<Entity>;
    try {
      sourceState = await this.entityRepository.getAll();
    } catch (err) {
      sourceState = this.entityStateService.getEmptyState();
    }

    const currentState = this.entityStateService.getCurrentState();
    if (sourceState.lastUpdated > currentState.lastUpdated && !writeOnly) {
      this.syncSourceWithCurrent(sourceState);
    } else if (sourceState.lastUpdated < currentState.lastUpdated) {
      await this.syncCurrentWithSource();
    }
  }

  private syncSourceWithCurrent(sourceState: EntityState<Entity>): void {
    const existingEntitiesMap = new Map<string, Entity>();
    this.entityStateService
      .getCurrentState()
      .entities.forEach((entity) => existingEntitiesMap.set(entity.id, entity));

    const updatedEntities = new Array<Entity>();
    const addedEntities = new Array<[Entity, number]>();
    sourceState.entities.forEach((entity, index) => {
      if (existingEntitiesMap.has(entity.id)) {
        if (
          entity.lastUpdated <= existingEntitiesMap.get(entity.id).lastUpdated
        )
          return;
        existingEntitiesMap.delete(entity.id);
        updatedEntities.push(entity);
      } else {
        addedEntities.push([entity, index]);
      }
    });

    // only initiate state update if there are any changes
    if (
      updatedEntities.length > 0 ||
      addedEntities.length > 0 ||
      existingEntitiesMap.size > 0
    ) {
      this.dataModelerStateService.updateStateAndEmitPatches(
        this.entityStateService,
        (draftState) => {
          updatedEntities.forEach((updatedEntity) => {
            this.entityStateService.updateEntity(
              draftState,
              updatedEntity.id,
              updatedEntity
            );
            this.entityStateUpdatesHandler.handleUpdatedEntity(updatedEntity);
          });
          addedEntities.forEach(([addedEntity, index]) => {
            this.entityStateService.addEntity(draftState, addedEntity, index);
            this.entityStateUpdatesHandler.handleNewEntity(addedEntity);
          });
        }
      );
    }
  }

  private async syncCurrentWithSource(): Promise<void> {
    await this.entityRepository.saveAll(
      this.entityStateService.getCurrentState()
    );
  }
}
