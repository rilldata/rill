import type {
    EntityRecord, EntityState,
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RootConfig } from "$common/config/RootConfig";
import type { EntityStateService } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {readFile, writeFile} from "fs/promises";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import { existsSync } from "fs";
import { execSync } from "node:child_process";

export class EntityStateSyncService<
    Entity extends EntityRecord,
    StateService extends EntityStateService<Entity>,
> {
    private readonly fileName: string;
    private syncTimer: NodeJS.Timer;

    public constructor(private readonly config: RootConfig,
                       private readonly entityType: EntityType,
                       private readonly stateType: StateType,
                       private readonly dataModelerStateService: DataModelerStateService,
                       private readonly entityStateService: StateService) {
        this.fileName = `${this.config.projectFolder}/` +
            `${this.stateType.toLowerCase()}_${this.entityType.toLowerCase()}_state.json`;
    }

    public async init(): Promise<void> {
        execSync(`mkdir -p ${this.config.projectFolder}`);

        let initialState: EntityState<Entity>;

        if (this.config.state.autoSync && existsSync(this.fileName)) {
            initialState = JSON.parse((await readFile(this.fileName)).toString());
        } else {
            initialState = {lastUpdated: Date.now(), entities: []};
        }
        this.entityStateService.init(initialState);
        if (!this.config.state.autoSync) return;

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
        if (!existsSync(this.fileName)) {
            await this.syncToCurrent();
        }

        let sourceState: EntityState<Entity>;
        try {
            sourceState = JSON.parse((await readFile(this.fileName)).toString());
        } catch (err) {
            sourceState = {lastUpdated: 0, entities: []};
        }

        const currentState = this.entityStateService.getCurrentState();
        if (sourceState.lastUpdated > currentState.lastUpdated && !writeOnly) {
            this.syncToSource(sourceState);
        } else if (sourceState.lastUpdated < currentState.lastUpdated) {
            await this.syncToCurrent();
        }
    }

    private syncToSource(sourceState: EntityState<Entity>): void {
        const existingEntitiesMap = new Map<string, Entity>();
        this.entityStateService.getCurrentState().entities.forEach(entity =>
            existingEntitiesMap.set(entity.id, entity));

        const updatedEntities = new Array<EntityRecord>();
        const addedEntities = new Array<[EntityRecord, number]>();
        sourceState.entities.forEach((entity, index) => {
            if (existingEntitiesMap.has(entity.id)) {
                if (entity.lastUpdated <= existingEntitiesMap.get(entity.id).lastUpdated) return;
                existingEntitiesMap.delete(entity.id);
                updatedEntities.push(entity);
            } else {
                addedEntities.push([entity, index]);
            }
        });

        // only initiate state update if there are any changes
        if (updatedEntities.length > 0 || addedEntities.length > 0) {
            this.dataModelerStateService.updateStateAndEmitPatches(
                this.entityStateService,
                (draftState) => {
                    updatedEntities.forEach(updatedEntity =>
                        this.entityStateService.updateEntity(draftState,
                            updatedEntity.id, updatedEntity as any));
                    addedEntities.forEach(([addedEntity, index]) =>
                        this.entityStateService.addEntity(draftState,
                            addedEntity as any, index));
                }
            );
        }
    }

    private async syncToCurrent(): Promise<void> {
        await writeFile(this.fileName, JSON.stringify(this.entityStateService.getCurrentState()));
    }
}
