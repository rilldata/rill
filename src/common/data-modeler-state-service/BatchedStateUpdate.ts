import type {
  EntityRecord,
  EntityState,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { EntityStateService } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { Debounce } from "$common/utils/Debounce";
import type { Patch } from "immer";

const StateUpdateThrottle = 200;

export class BatchedStateUpdate {
  private callbacksByType = new Map<
    string,
    Array<(draft: EntityState<EntityRecord>) => void>
  >();
  private serviceMap = new Map<string, EntityStateService<any>>();
  private debounce = new Debounce();
  private promisesByType = new Map<string, Array<() => void>>();

  constructor(
    private readonly patchesCallback: (
      patches: Array<Patch>,
      entityType: EntityType,
      stateType: StateType
    ) => void
  ) {}

  public updateState(
    service: EntityStateService<any>,
    callback: (draft: EntityState<EntityRecord>) => void
  ) {
    const key = BatchedStateUpdate.getKey(
      service.entityType,
      service.stateType
    );

    if (!this.callbacksByType.has(key)) {
      this.callbacksByType.set(key, [callback]);
    } else {
      this.callbacksByType.get(key).push(callback);
    }

    if (!this.serviceMap.has(key)) {
      this.serviceMap.set(key, service);
    }

    this.debounce.debounce(
      key,
      () => this.batchUpdateState(key),
      StateUpdateThrottle
    );
  }

  public callThrough(service: EntityStateService<any>) {
    this.batchUpdateState(
      BatchedStateUpdate.getKey(service.entityType, service.stateType)
    );
  }

  public async waitForNextUpdate(
    entityType: EntityType,
    stateType: StateType
  ): Promise<void> {
    const key = BatchedStateUpdate.getKey(entityType, stateType);

    if (!this.callbacksByType.get(key)?.length) return;

    return new Promise<void>((resolve) => {
      if (!this.promisesByType.has(key)) {
        this.promisesByType.set(key, [resolve]);
      } else {
        this.promisesByType.get(key).push(resolve);
      }
    });
  }

  private batchUpdateState(key: string) {
    if (!this.callbacksByType.get(key)?.length) return;
    const callbacks = this.callbacksByType.get(key);
    this.callbacksByType.set(key, []);
    this.debounce.clear(key);

    const service = this.serviceMap.get(key);

    service.updateState(
      (draft) => {
        callbacks.forEach((callback) => callback(draft));
        draft.lastUpdated = Date.now();
      },
      (patches) => {
        this.patchesCallback(patches, service.entityType, service.stateType);
      }
    );

    if (this.promisesByType.has(key)) {
      const resolves = this.promisesByType.get(key);
      this.promisesByType.set(key, []);
      // wait for a small amount of time to make sure state propagates
      setTimeout(() => {
        resolves.forEach((resolve) => resolve());
      }, 5);
    }
  }

  private static getKey(entityType: EntityType, stateType: StateType): string {
    return `${entityType}-${stateType}`;
  }
}
