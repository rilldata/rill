import {
  ResourceKind,
  useProjectParser,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { runtimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type {
  V1ParseError,
  V1Resource,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, Readable, writable } from "svelte/store";

/**
 * Global resources store that maps file name to a resource.
 */
export type ResourcesState = {
  // this is just a mapping of file path to resource name
  // storing the entire resource is not necessary since tanstack query will do that for the get resource api
  resources: Record<string, V1ResourceName>;
};

const { update, subscribe } = writable({
  resources: {},
} as ResourcesState);

const resourcesStoreReducers = {
  async init(instanceId: string) {
    const resourcesResp = await runtimeServiceListResources(instanceId);
    for (const resource of resourcesResp.resources) {
      switch (resource.meta.name.kind) {
        case ResourceKind.Source:
        case ResourceKind.Model:
        case ResourceKind.MetricsView:
          this.setResource(resource);
          break;
      }
    }
  },

  setResource(resource: V1Resource) {
    update((state) => {
      for (const filePath of resource.meta.filePaths) {
        state.resources[filePath] = resource.meta.name;
      }
      return state;
    });
  },

  deleteFile(filePath: string) {
    update((state) => {
      if (state.resources[filePath]) delete state.resources[filePath];
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

export function getResourceNameForFile(filePath: string) {
  return derived([resourcesStore], ([state]) => state.resources[filePath]);
}

export function useResourceForFile(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string
): CreateQueryResult<V1Resource> {
  return derived([getResourceNameForFile(filePath)], ([resourceName], set) => {
    return useResource(
      instanceId,
      resourceName?.name,
      resourceName?.kind as ResourceKind,
      undefined,
      queryClient
    ).subscribe(set);
  });
}

// TODO: memoize?
export function getAllErrorsForFile(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string
): Readable<Array<V1ParseError>> {
  return derived(
    [
      useProjectParser(queryClient, instanceId),
      useResourceForFile(queryClient, instanceId, filePath),
    ],
    ([projectParser, resource]) => {
      if (
        projectParser.isFetching ||
        projectParser.isError ||
        resource.isFetching
      ) {
        // TODO: what should the error be for failed get resource API
        return [];
      }
      console.log(
        filePath,
        projectParser.data?.projectParser?.state?.parseErrors
      );
      return [
        ...(projectParser.data?.projectParser?.state?.parseErrors ?? []).filter(
          (e) => e.filePath === filePath
        ),
        ...(resource.data?.meta?.reconcileError
          ? [
              {
                filePath,
                message: resource.data.meta.reconcileError,
              },
            ]
          : []),
      ];
    },
    []
  );
}

export function getFileHasErrors(
  queryClient: QueryClient,
  instanceId: string,
  filePath: string
): Readable<boolean> {
  return derived(
    [getAllErrorsForFile(queryClient, instanceId, filePath)],
    ([errors]) => errors.length > 0
  );
}
