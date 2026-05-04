import {
  createRuntimeServiceListResources,
  type V1ListResourcesResponse,
  type V1OlapTableInfo,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { connectorServiceOLAPListTables } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createInfiniteQuery } from "@tanstack/svelte-query";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { derived, type Readable } from "svelte/store";

/**
 * Paginated tables list using cursor pagination.
 * Accumulates pages into a flat array via `select`.
 * Supports server-side search via ILIKE `searchPattern`.
 *
 * Accepts a reactive store so that `createInfiniteQuery` is called once
 * during component initialization; TanStack Query updates the observer
 * in-place when the derived options change.
 */
export function useInfiniteTablesList(
  params: Readable<{
    client: RuntimeClient;
    connector: string;
    searchPattern?: string;
  }>,
) {
  const optionsStore = derived(params, ($p) => ({
    queryKey: [
      "/v1/olap/tables-infinite",
      {
        instanceId: $p.client.instanceId,
        connector: $p.connector,
        searchPattern: $p.searchPattern,
      },
    ],
    enabled: !!$p.client.instanceId && !!$p.connector,
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage: { nextPageToken?: string }) =>
      lastPage?.nextPageToken || undefined,
    queryFn: ({ pageParam }: { pageParam?: string }) =>
      connectorServiceOLAPListTables($p.client, {
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
 */
function buildModelResourcesMap(
  resources: V1Resource[] | undefined,
): Map<string, V1Resource> {
  const map = new Map<string, V1Resource>();
  resources?.forEach((resource) => {
    const tableName = resource.model?.state?.resultTable;
    if (tableName) {
      map.set(tableName.toLowerCase(), resource);
    }
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
export function useModelResources(client: RuntimeClient) {
  return createRuntimeServiceListResources(
    client,
    { kind: ResourceKind.Model },
    {
      query: {
        select: (data: V1ListResourcesResponse) =>
          buildModelResourcesMap(data.resources),
        enabled: !!client.instanceId,
      },
    },
  );
}
