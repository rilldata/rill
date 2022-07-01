import type { EntityState } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { readFile, writeFile } from "fs/promises";
import type { StateConfig } from "$common/config/StateConfig";
import type {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { existsSync } from "fs";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";

/**
 * Entity repository that writes to file.
 * This class will deal with loading and saving all entity data.
 * Currently, this directly saves and reads from files.
 * This can later be swapped with something that persists into a DB.
 */
export class EntityRepository<Entity extends EntityRecord> {
  private readonly fileName: string;

  constructor(
    protected readonly stateConfig: StateConfig,
    protected readonly dataModelerService: DataModelerService,
    entityType: EntityType,
    stateType: StateType
  ) {
    this.fileName =
      `${stateConfig.stateFolder}/` +
      `${stateType.toLowerCase()}_${entityType.toLowerCase()}_state.json`;
  }

  public async sourceExists(): Promise<boolean> {
    return existsSync(this.fileName);
  }

  public async saveAll(state: EntityState<Entity>): Promise<void> {
    if (!this.stateConfig.autoSync) return;
    await writeFile(this.fileName, JSON.stringify(state));
    await Promise.all(state.entities.map((entity) => this.save(entity)));
  }

  /**
   * Save a specific entity
   */
  public async save(_entity: Entity): Promise<void> {
    return Promise.resolve();
  }

  public async getAll(): Promise<EntityState<Entity>> {
    const state: EntityState<Entity> = JSON.parse(
      (await readFile(this.fileName)).toString()
    );
    const updates = await Promise.all(
      state.entities.map((entity) => this.update(entity))
    );
    // if any entity updated save it back
    if (updates.some((update) => update)) {
      await this.saveAll(state);
    }
    // if lastUpdated of any entity has updated then update state's lastUpdated as well
    state.lastUpdated = Math.max(
      state.lastUpdated,
      ...state.entities.map((entity) => entity.lastUpdated)
    );
    return state;
  }

  /**
   * Update specific fields in entity based on id or any other field
   */
  public async update(_entity: Entity): Promise<boolean> {
    return false;
  }
}
