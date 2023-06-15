import type { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { Readable, derived, writable } from "svelte/store";
import { page } from "$app/stores";
import { MetricsEventScreenName } from "../metrics/service/MetricsTypes";

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

// We should rewrite ActiveEntity using appScreen dervied store
const { update, subscribe } = writable({
  activeEntity: undefined,
  previousActiveEntity: undefined,
} as AppStore);

const appStoreReducers = {
  setActiveEntity(name: string, type: EntityType) {
    update((state) => {
      if (state.previousActiveEntity) {
        httpRequestQueue.inactiveByName(state.previousActiveEntity.name);
      }
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

export const appScreen = derived(page, ($page) => {
  switch ($page.route.id) {
    case "/(application)":
      return MetricsEventScreenName.Home;
    case "/(application)/source/[name]":
      return MetricsEventScreenName.Source;
    case "/(application)/model/[name]":
      return MetricsEventScreenName.Model;
    case "/(application)/dashboard/[name]":
      return MetricsEventScreenName.Dashboard;
    case "/(application)/dashboard/[name]/edit":
      return MetricsEventScreenName.MetricsDefinition;
    case "/(application)/welcome":
      return MetricsEventScreenName.Splash;
    default:
      // Return home as default
      return MetricsEventScreenName.Home;
  }
});
