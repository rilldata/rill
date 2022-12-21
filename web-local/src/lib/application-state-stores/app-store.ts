import type { EntityType } from "@rilldata/web-common/lib/entity";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { Readable, writable } from "svelte/store";

export interface ActiveEntity {
  type: EntityType;
  id?: string;
  name: string;
}

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
      if (state.previousActiveEntity) {
        httpRequestQueue.inactiveByName(name);
      }
      return state;
    });
  },
};

export const appStore: Readable<AppStore> & typeof appStoreReducers = {
  subscribe,
  ...appStoreReducers,
};
