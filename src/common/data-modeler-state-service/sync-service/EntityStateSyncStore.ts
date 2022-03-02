import type { EntityState } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { readFile, writeFile } from "fs/promises";
import type { RootConfig } from "$common/config/RootConfig";
import type { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { existsSync } from "fs";

/**
 * Wrapper to read and write to the store.
 * Currently, maintains a file. Can be modified to read from a database.
 */
export class EntityStateSyncStore<Entity extends EntityRecord> {
    private readonly fileName: string;

    constructor(config: RootConfig, entityType: EntityType, stateType: StateType) {
        this.fileName = `${config.projectFolder}/` +
            `${stateType.toLowerCase()}_${entityType.toLowerCase()}_state.json`;
    }

    public async sourceExists(): Promise<boolean> {
        return existsSync(this.fileName);
    }

    public async writeToSource(state: EntityState<Entity>): Promise<void> {
        await writeFile(this.fileName, JSON.stringify(state));
    }

    public async readFromSource(): Promise<EntityState<Entity>> {
        return JSON.parse((await readFile(this.fileName)).toString());
    }
}
