import {EntityRepository} from "$common/data-modeler-state-service/sync-service/EntityRepository";
import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {StateConfig} from "$common/config/StateConfig";
import type {EntityType, StateType} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {existsSync, mkdirSync, readFileSync, writeFileSync, statSync} from "fs";

export class PersistentModelRepository extends EntityRepository<PersistentModelEntity> {
    constructor(stateConfig: StateConfig, entityType: EntityType, stateType: StateType) {
        super(stateConfig, entityType, stateType);
        if (!existsSync(stateConfig.modelFolder)) {
            mkdirSync(stateConfig.modelFolder, {recursive: true});
        }
    }

    /**
     * Persist the entity query to file.
     */
    public async save(entity: PersistentModelEntity): Promise<void> {
        writeFileSync(`${this.stateConfig.modelFolder}/${entity.name}`, entity.query);
    }

    /**
     * Update specific fields in entity based on id or any other field
     */
    public async update(entity: PersistentModelEntity): Promise<boolean> {
        const modelFileName = `${this.stateConfig.modelFolder}/${entity.name}`;
        // if file was deleted for any reason, recreate instead of throwing error
        if (!existsSync(modelFileName)) {
            await this.save(entity);
            return false;
        }

        const newQuery = readFileSync(modelFileName).toString();
        const fileUpdated = statSync(modelFileName).mtimeMs;
        if (newQuery !== entity.query && fileUpdated > entity.lastUpdated) {
            entity.query = newQuery;
            entity.lastUpdated = Date.now();
            return true;
        }
        return false;
    }
}
