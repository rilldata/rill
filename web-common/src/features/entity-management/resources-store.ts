import { newFileArtifactStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store-new";
import {
  ResourceKind,
  useProjectParser,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type {
  V1ParseError,
  V1Resource,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import {
  V1ReconcileStatus,
  runtimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult, QueryClient } from "@tanstack/svelte-query";
import { Readable, derived, get, writable } from "svelte/store";
import { removeLeadingSlash } from "./entity-mappers";

/**
 * Global resources store that maps file name to a resource.
 */
// TODO: Merge with FileArtifactsStore.
//       Have an entry with filePath to object with resource name and reconciling and other stuff from FileArtifactsStore
export type ResourcesState = {
  // this is just a mapping of file path to resource name
  // storing the entire resource is not necessary since tanstack query will do that for the get resource api
  resources: Record<string, V1ResourceName>;
  // array of paths currently reconciling
  // we use path since parse error will only give us paths from ProjectParser
  currentlyReconciling: Record<string, V1ResourceName>;
  // last time the state of the resource `kind/name` was updated
  // used to make sure we do not have unnecessary refreshes
  lastStateUpdatedOn: Record<string, string>;
};

const { update, subscribe } = writable<ResourcesState>({
  resources: {},
  currentlyReconciling: {},
  lastStateUpdatedOn: {},
});

const resourcesStoreReducers = {
  async init(instanceId: string) {
    const resourcesResp = await runtimeServiceListResources(instanceId);
    for (const resource of resourcesResp.resources) {
      switch (resource.meta.name.kind) {
        case ResourceKind.Source:
        case ResourceKind.Model:
        case ResourceKind.MetricsView:
          this.setResource(resource);
          if (
            resource.meta.reconcileStatus ===
            V1ReconcileStatus.RECONCILE_STATUS_RUNNING
          ) {
            this.reconciling(resource);
          }
          this.setVersion(resource);
          break;
      }
    }
  },

  setResource(resource: V1Resource) {
    update((state) => {
      for (const filePath of resource.meta.filePaths) {
        state.resources[removeLeadingSlash(filePath)] = resource.meta.name;
      }
      return state;
    });
  },

  deleteFile(filePath: string) {
    update((state) => {
      delete state.resources[filePath];
      delete state.currentlyReconciling[filePath];
      return state;
    });
  },

  reconciling(resource: V1Resource) {
    update((state) => {
      for (const path of resource.meta.filePaths) {
        state.currentlyReconciling[path] = resource.meta.name;
      }
      return state;
    });
  },

  doneReconciling(resource: V1Resource) {
    update((state) => {
      if (resource.meta.name.kind === ResourceKind.ProjectParser) {
        for (const parseError of resource.projectParser.state.parseErrors) {
          delete state.currentlyReconciling[parseError.filePath];
        }
      } else {
        for (const filePath of resource.meta.filePaths) {
          delete state.currentlyReconciling[filePath];
        }
      }
      return state;
    });
  },

  setVersion(resource: V1Resource) {
    update((state) => {
      state.lastStateUpdatedOn[getKeyForResource(resource)] =
        resource.meta.stateUpdatedOn;
      return state;
    });
  },

  deleteResource(resource: V1Resource) {
    update((state) => {
      delete state.lastStateUpdatedOn[getKeyForResource(resource)];
      return state;
    });
  },
};

export type ResourcesStore = Readable<ResourcesState> &
  typeof resourcesStoreReducers;
export const resourcesStore: ResourcesStore = {
  subscribe,
  ...resourcesStoreReducers,
};

function getKeyForResource(resource: V1Resource) {
  return `${resource.meta.name.kind}/${resource.meta.name.name}`;
}
