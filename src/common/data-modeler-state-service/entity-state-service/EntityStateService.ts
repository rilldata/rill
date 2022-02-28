import { writable, get } from "svelte/store";
import type {Writable} from "svelte/store";
import { shallowCopy } from "$common/utils/shallowCopy";
import produce, { applyPatches, Patch } from "immer";

export enum EntityType {
    Table = "Table",
    Model = "Model",
}

export enum StateType {
    Persistent = "Persistent",
    Derived = "Derived",
}
export const AllStateTypes = [StateType.Persistent, StateType.Derived];

export interface EntityRecord {
    id: string;
    type: EntityType;
    lastUpdated: number;
}

export enum EntityStatus {
    Idle,
    Profiling,
    Syncing
}
export interface DerivedEntityRecord extends EntityRecord {
    status: EntityStatus;
}

export interface EntityState<Entity extends EntityRecord> {
    entities: Array<Entity>;
    lastUpdated: number;
}

export type EntityStateActionArg<Entity extends EntityRecord, Service = EntityStateService<Entity>> = {
    stateService: Service;
    draftState: EntityState<Entity>;
}

/**
 * Each entity will have Persistent or Derived states. (Could be more, depends on {@link StateType} enum)
 * This is an abstraction around such states where there is an array of entities.
 * Each entity must have and id and type ({@link EntityRecord}).
 *
 * Has CRUD methods. Can be overridden later on to fetch from a DB or an API.
 */
export abstract class EntityStateService<Entity extends EntityRecord> {
    public store: Writable<EntityState<Entity>>;

    public readonly entityType: EntityType;
    public readonly stateType: StateType;

    public init(initialState: EntityState<Entity>): void {
        this.store = writable(initialState);
    }

    public getCurrentState(): EntityState<Entity> {
        return get(this.store);
    }

    public updateState(draftModCallback: (draft: EntityState<Entity>) => void,
                       pathCallback: (patches: Array<Patch>) => void): void {
        this.store.set(produce(this.getCurrentState(), (draft) => {
            draftModCallback(draft as any);
        }, pathCallback));
    }

    public applyPatches(patches: Array<Patch>): void {
        this.store.set(applyPatches(this.getCurrentState(), patches));
    }

    public getById(id: string, state = this.getCurrentState()): Entity {
        return state.entities.find(entity => entity.id === id);
    }

    public getByField<Field extends keyof Entity>(field: Field, value: Entity[Field],
                                                  state = this.getCurrentState()): Entity {
        return state.entities.find(entity => entity[field] === value);
    }

    public addEntity(draftState: EntityState<Entity>, newEntity: Entity, atIndex?: number): void {
        // TODO: validate id conflicts
        if (atIndex) {
            draftState.entities.splice(atIndex, 0, newEntity);
        }
        else {
            draftState.entities.push(newEntity);
        }
    }

    public updateEntity(draftState: EntityState<Entity>, id: string, newEntity: Entity): void {
        const entity = this.getById(id, draftState);
        if (!entity) {
            console.error(`Record not found. entityType=${this.entityType} stateType=${this.stateType} id=${id}`);
        }
        shallowCopy(newEntity, entity);
        entity.lastUpdated = Date.now();
    }

    public updateEntityField<Field extends keyof Entity>(draftState: EntityState<Entity>,
                                                         id: string, field: Field, value: Entity[Field]): void {
        const entity = this.getById(id, draftState);
        if (!entity) {
            console.error(`Record not found. entityType=${this.entityType} stateType=${this.stateType} id=${id}`);
        }
        entity[field] = value;
        entity.lastUpdated = Date.now();
    }

    public deleteEntity(draftState: EntityState<Entity>, id: string): void {
        const index = draftState.entities.findIndex(entity => entity.id === id);
        if (index === -1) {
            console.error(`Record not found. entityType=${this.entityType} stateType=${this.stateType} id=${id}`);
        }
        draftState.entities.splice(index, 1);
    }

    public moveEntityDown(draftState: EntityState<Entity>, id: string): void {
        const index = draftState.entities.findIndex(entity => entity.id === id);
        if (index === -1 || index === draftState.entities.length - 1) return;

        draftState.entities[index].lastUpdated = Date.now();
        draftState.entities[index + 1].lastUpdated = Date.now();
        [draftState.entities[index], draftState.entities[index + 1]] =
            [draftState.entities[index + 1], draftState.entities[index]];
    }

    public moveEntityUp(draftState: EntityState<Entity>, id: string): void {
        const index = draftState.entities.findIndex(entity => entity.id === id);
        if (index === -1 || index === 0) return;

        draftState.entities[index].lastUpdated = Date.now();
        draftState.entities[index - 1].lastUpdated = Date.now();
        [draftState.entities[index], draftState.entities[index - 1]] =
            [draftState.entities[index - 1], draftState.entities[index]];
    }
}
