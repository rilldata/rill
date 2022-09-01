import { shallowCopy } from "$common/utils/shallowCopy";
import type { Patch } from "immer";
import { applyPatches, produce } from "immer";
import type { Writable } from "svelte/store";
import { get, writable } from "svelte/store";

export enum EntityType {
  Table = "Table",
  Model = "Model",
  Application = "Application",
  MetricsDefinition = "MetricsDefinition",
  MeasureDefinition = "MeasureDefinition",
  DimensionDefinition = "DimensionDefinition",
  MetricsExplorer = "MetricsExplorer",
}

export enum StateType {
  Persistent = "Persistent",
  Derived = "Derived",
}

export interface EntityRecord {
  id: string;
  type: EntityType;
  lastUpdated: number;
}

export enum EntityStatus {
  Idle,
  Running,
  Error,

  Importing,
  Validating,
  Profiling,
  Exporting,
}
export interface DerivedEntityRecord extends EntityRecord {
  status: EntityStatus;
}

export interface EntityState<Entity extends EntityRecord> {
  entities: Array<Entity>;
  lastUpdated: number;
}

export type EntityStateActionArg<
  Entity extends EntityRecord,
  State extends EntityState<Entity> = EntityState<Entity>,
  Service extends EntityStateService<Entity, State> = EntityStateService<
    Entity,
    State
  >
> = {
  stateService: Service;
  draftState: State;
};

/**
 * Each entity will have Persistent or Derived states. (Could be more, depends on {@link StateType} enum)
 * This is an abstraction around such states where there is an array of entities.
 * Each entity must have and id and type ({@link EntityRecord}).
 *
 * Has CRUD methods. Can be overridden later on to fetch from a DB or an API.
 */
export abstract class EntityStateService<
  Entity extends EntityRecord,
  State extends EntityState<Entity> = EntityState<Entity>
> {
  public store: Writable<State>;

  public readonly entityType: EntityType;
  public readonly stateType: StateType;

  public constructor() {
    // need an empty state for UI where state init is async but component init is not
    this.init({ entities: [], lastUpdated: 0 } as State);
  }

  public init(initialState: State): void {
    this.store = writable(initialState);
  }

  public getCurrentState(): State {
    return get(this.store);
  }

  public getEmptyState(): State {
    return { lastUpdated: 0, entities: [] } as State;
  }

  public updateState(
    draftModCallback: (draft: State) => void,
    pathCallback: (patches: Array<Patch>) => void
  ): void {
    this.store.set(
      produce(
        this.getCurrentState(),
        (draft) => {
          draftModCallback(draft as State);
        },
        pathCallback
      )
    );
  }

  public applyPatches(patches: Array<Patch>): void {
    this.store.set(applyPatches(this.getCurrentState(), patches));
  }

  public getById(id: string, state: State = this.getCurrentState()): Entity {
    return state.entities.find((entity) => entity.id === id);
  }

  public getByField<Field extends keyof Entity>(
    field: Field,
    value: Entity[Field],
    state: State = this.getCurrentState()
  ): Entity {
    return state.entities.find((entity) => entity[field] === value);
  }
  public getManyByField<Field extends keyof Entity>(
    field: Field,
    value: Entity[Field],
    state: State = this.getCurrentState()
  ): Array<Entity> {
    return state.entities.filter((entity) => entity[field] === value);
  }

  public addEntity(
    draftState: State,
    newEntity: Entity,
    atIndex?: number
  ): void {
    // TODO: validate id conflicts
    if (atIndex) {
      draftState.entities.splice(atIndex, 0, newEntity);
    } else {
      draftState.entities.push(newEntity);
    }
  }

  public updateEntity(draftState: State, id: string, newEntity: Entity): void {
    const entity = this.getById(id, draftState);
    if (!entity) {
      console.error(
        `Record not found. entityType=${this.entityType} stateType=${this.stateType} id=${id}`
      );
    }
    shallowCopy(newEntity, entity);
    entity.lastUpdated = Date.now();
  }

  public updateEntityField<Field extends keyof Entity>(
    draftState: State,
    id: string,
    field: Field,
    value: Entity[Field]
  ): void {
    const entity = this.getById(id, draftState);
    if (!entity) {
      console.error(
        `Record not found. entityType=${this.entityType} stateType=${this.stateType} id=${id}`
      );
    }
    entity[field] = value;
    entity.lastUpdated = Date.now();
  }

  public deleteEntity(draftState: State, id: string): void {
    const index = draftState.entities.findIndex((entity) => entity.id === id);
    if (index === -1) {
      console.error(
        `Record not found. entityType=${this.entityType} stateType=${this.stateType} id=${id}`
      );
    }
    draftState.entities.splice(index, 1);
  }

  public moveEntityDown(draftState: State, id: string): void {
    const index = draftState.entities.findIndex((entity) => entity.id === id);
    if (index === -1 || index === draftState.entities.length - 1) return;

    draftState.entities[index].lastUpdated = Date.now();
    draftState.entities[index + 1].lastUpdated = Date.now();
    [draftState.entities[index], draftState.entities[index + 1]] = [
      draftState.entities[index + 1],
      draftState.entities[index],
    ];
  }

  public moveEntityUp(draftState: State, id: string): void {
    const index = draftState.entities.findIndex((entity) => entity.id === id);
    if (index === -1 || index === 0) return;

    draftState.entities[index].lastUpdated = Date.now();
    draftState.entities[index - 1].lastUpdated = Date.now();
    [draftState.entities[index], draftState.entities[index - 1]] = [
      draftState.entities[index - 1],
      draftState.entities[index],
    ];
  }
}
