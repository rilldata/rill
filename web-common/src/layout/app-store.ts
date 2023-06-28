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
interface AppStore {
  activeEntity: ActiveEntity;
  previousActiveEntity: ActiveEntity;
}

export const appScreen = derived(page, ($page) => {
  let activeEntity;
  switch ($page.route.id) {
    case "/(application)":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Home,
      };
      break;
    case "/(application)/source/[name]":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Source,
      };
      break;
    case "/(application)/model/[name]":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Model,
      };
      break;
    case "/(application)/dashboard/[name]":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Dashboard,
      };
      break;
    case "/(application)/dashboard/[name]/edit":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.MetricsDefinition,
      };
      break;
    case "/(application)/welcome":
      activeEntity = {
        name: $page?.params?.name,
        type: MetricsEventScreenName.Splash,
      };
      break;
    default:
      // Return home as default
      activeEntity = { name: "", type: MetricsEventScreenName.Home };
  }

  appStore.setActiveEntity(activeEntity.name, activeEntity.type);
  return activeEntity;
});

// App store is being utilized for making previous entity inactive in the HTTP request queue
const { update } = writable({
  activeEntity: undefined,
  previousActiveEntity: undefined,
} as AppStore);

const appStoreReducers = {
  setActiveEntity(name: string, type: EntityType) {
    update((state) => {
      state.previousActiveEntity = state.activeEntity;

      if (state.previousActiveEntity) {
        httpRequestQueue.inactiveByName(state.previousActiveEntity.name);
      }
      state.activeEntity = {
        name,
        type,
      };
      return state;
    });
  },
};

const appStore: Readable<AppStore> & typeof appStoreReducers = {
  ...appStoreReducers,
};
