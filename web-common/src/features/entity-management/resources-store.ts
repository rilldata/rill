import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type {
  V1ParseError,
  V1Resource,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, Readable, writable } from "svelte/store";

/**
 * Global resources store that maps file name to a resource and parse errors.
 */
export type ResourcesState = {
  // this is just a mapping of file path to resource name
  // storing the entire resource is not necessary since tanstack query will do that for the get resource api
  resources: Record<string, V1ResourceName>;
  errors: Record<string, Array<V1ParseError>>;
};

const { update, subscribe } = writable({
  resources: {},
  errors: {},
} as ResourcesState);

const resourcesStoreReducers = {
  // TODO: set this on mount to get the initial errors
  setProjectParseErrors(parseErrors: Array<V1ParseError>) {
    const errors: Record<string, Array<V1ParseError>> = {};
    for (const parseError of parseErrors) {
      if (errors[parseError.filePath]) {
        errors[parseError.filePath].push(parseError);
      } else {
        errors[parseError.filePath] = [parseError];
      }
    }

    update((state) => {
      state.errors = errors;
      return state;
    });
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
      if (state.errors[filePath]) delete state.errors[filePath];
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

export function getParseErrorsForFile(filePath: string) {
  return derived([resourcesStore], ([state]) => state.errors[filePath] ?? []);
}

export function getResourceNameForFile(filePath: string) {
  return derived([resourcesStore], ([state]) => state.resources[filePath]);
}

export function useResourceForFile(
  instanceId: string,
  filePath: string
): CreateQueryResult<V1Resource> {
  return derived([getResourceNameForFile(filePath)], ([resourceName], set) =>
    useResource(
      instanceId,
      resourceName?.name,
      resourceName?.kind as ResourceKind
    ).subscribe(set)
  );
}

// TODO: memoize?
export function getAllErrorsForFile(
  instanceId: string,
  filePath: string
): Readable<Array<V1ParseError>> {
  return derived(
    [getParseErrorsForFile(filePath), useResourceForFile(instanceId, filePath)],
    ([parseErrors, resource]) => {
      if (resource.isFetching || resource.isError) {
        // TODO: what should the error be for failed get resource API
        return [...parseErrors];
      }
      return [
        ...parseErrors,
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
  instanceId: string,
  filePath: string
): Readable<boolean> {
  return derived(
    [getAllErrorsForFile(instanceId, filePath)],
    ([errors]) => errors.length > 0
  );
}
