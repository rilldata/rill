import type { ActiveEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { Readable, writable } from "svelte/store";

/**
 * App wide store to store metadata
 * Currently caches active entity from URL
 */
export interface AppStore {
  activeEntity: ActiveEntity;
  previousActiveEntity: ActiveEntity;
}

const { update, subscribe } = writable({
  activeEntity: undefined,
  previousActiveEntity: undefined,
} as AppStore);

const appStoreReducers = {
  setActiveEntity(name: string, type: EntityType) {
    update((state) => {
      state.activeEntity = {
        name,
        type,
      };
      state.previousActiveEntity = state.activeEntity;
      return state;
    });
  },
};

export const appStore: Readable<AppStore> & typeof appStoreReducers = {
  subscribe,
  ...appStoreReducers,
};
