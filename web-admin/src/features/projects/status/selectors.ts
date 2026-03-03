import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListResources,
  createRuntimeServicePing,
  type V1ListResourcesResponse,
  type V1OlapTableInfo,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { connectorServiceOLAPListTables } from "@rilldata/web-common/runtime-client/gen/connector-service/connector-service";
import { createInfiniteQuery } from "@tanstack/svelte-query";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { derived, type Readable } from "svelte/store";
import { smartRefetchIntervalFunc } from "@rilldata/web-admin/lib/refetch-interval-store";

export function useProjectDeployment(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1Deployment | undefined>(
    orgName,
    projName,
    undefined,
    {
      query: {
        select: (data: { deployment?: V1Deployment }) => {
          // There may not be a deployment if the project is hibernating
          return data?.deployment;
        },
      },
    },
  );
}

/**
 * Filters resources for display, removing hidden and internal resource kinds.
 */
export function filterResourcesForDisplay(
  resources: V1Resource[] | undefined,
): V1Resource[] {
  return (
    resources?.filter(
      (resource) =>
        !resource?.meta?.hidden &&
        resource?.meta?.name?.kind !== ResourceKind.ProjectParser &&
        resource?.meta?.name?.kind !== ResourceKind.RefreshTrigger &&
        resource?.meta?.name?.kind !== ResourceKind.Component &&
        resource?.meta?.name?.kind !== ResourceKind.Migration,
    ) ?? []
  );
}

export function useResources(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    {},
    {
      query: {
        select: (data: V1ListResourcesResponse) => ({
          ...data,
          resources: filterResourcesForDisplay(data?.resources),
        }),
        refetchInterval: smartRefetchIntervalFunc,
      },
    },
  );
}

/**
 * Paginated tables list using cursor pagination.
 * Accumulates pages into a flat array via `select`.
 * Supports server-side search via ILIKE `searchPattern`.
 *
 * Accepts a reactive store so that `createInfiniteQuery` is called once
 * during component initialization; TanStack Query updates the observer
 * in-place when the derived options change.
 *
 * NOTE: `createInfiniteQuery` cannot be re-created inside a Svelte `$:` block
 * (unlike `createQuery`). Re-creation causes a white-page crash on first
 * client-side navigation because the InfiniteQueryObserver teardown/setup
 * cycle corrupts Svelte's flush. The store-based approach avoids this.
 */
export function useInfiniteTablesList(
  params: Readable<{
    instanceId: string;
    connector: string;
    searchPattern?: string;
  }>,
) {
  const optionsStore = derived(params, ($p) => ({
    queryKey: [
      "/v1/olap/tables-infinite",
      {
        instanceId: $p.instanceId,
        connector: $p.connector,
        searchPattern: $p.searchPattern,
      },
    ],
    enabled: !!$p.instanceId && !!$p.connector,
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage: { nextPageToken?: string }) =>
      lastPage?.nextPageToken || undefined,
    queryFn: ({ pageParam }: { pageParam?: string }) =>
      connectorServiceOLAPListTables({
        instanceId: $p.instanceId,
        connector: $p.connector,
        searchPattern: $p.searchPattern,
        pageToken: pageParam,
      }),
    select: (data: any) => ({
      tables: data.pages.flatMap(
        (p: { tables?: V1OlapTableInfo[] }) => p.tables ?? [],
      ),
    }),
  }));

  return createInfiniteQuery(optionsStore);
}

/**
 * Builds a Map of model resources indexed by result table name and model name (case-insensitive).
 * This allows looking up model resource data by the OLAP table name.
 */
export function buildModelResourcesMap(
  resources: V1Resource[] | undefined,
): Map<string, V1Resource> {
  const map = new Map<string, V1Resource>();
  resources?.forEach((resource) => {
    // Index by resultTable (the actual output table name)
    const tableName = resource.model?.state?.resultTable;
    if (tableName) {
      map.set(tableName.toLowerCase(), resource);
    }
    // Also index by model name as fallback
    const modelName = resource.meta?.name?.name;
    if (modelName) {
      map.set(modelName.toLowerCase(), resource);
    }
  });
  return map;
}

/**
 * Fetches model resources and maps them by their result table name.
 */
export function useModelResources(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    { kind: ResourceKind.Model },
    {
      query: {
        select: (data: V1ListResourcesResponse) =>
          buildModelResourcesMap(data.resources),
        enabled: !!instanceId,
      },
    },
  );
}

export function useRuntimeVersion() {
  return createRuntimeServicePing({
    query: {
      staleTime: 60000, // Cache for 1 minute
    },
  });
}
