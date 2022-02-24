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
            `${this.entityType.toLowerCase()}_${this.stateType.toLowerCase()}.json`;
    }

    public async init(): Promise<void> {
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
            await this.syncWithCurrent();
        }

        const sourceState: EntityState<Entity> =
            JSON.parse((await readFile(this.fileName)).toString());

        const currentState = this.entityStateService.getCurrentState();
        if (sourceState.lastUpdated > currentState.lastUpdated && !writeOnly) {
            this.syncWithSource(sourceState);
        } else if (sourceState.lastUpdated < currentState.lastUpdated) {
            await this.syncWithCurrent();
        }
    }

    private syncWithSource(sourceState: EntityState<Entity>): void {
        const existingEntitiesMap = new Map<string, Entity>();
        this.entityStateService.getCurrentState().entities.forEach(entity =>
            existingEntitiesMap.set(entity.id, entity));

        this.dataModelerStateService.updateStateAndEmitPatches(
            this.entityStateService,
            (draftState) => {
                this.syncWithSourceEntities(draftState, sourceState, existingEntitiesMap);
            }
        );
    }
    private syncWithSourceEntities(draftState: EntityState<Entity>, sourceState: EntityState<Entity>,
                                   existingEntitiesMap: Map<string, Entity>): void {

        sourceState.entities.forEach((entity, index) => {
            if (existingEntitiesMap.has(entity.id)) {
                existingEntitiesMap.delete(entity.id);
                if (entity.lastUpdated <= existingEntitiesMap.get(entity.id).lastUpdated) return;

                this.entityStateService.updateEntity(draftState, entity.id, entity);
            } else {
                this.entityStateService.addEntity(draftState, entity, index);
            }
        });
    }

    private async syncWithCurrent(): Promise<void> {
        await writeFile(this.fileName, JSON.stringify(this.entityStateService.getCurrentState()));
    }
}
