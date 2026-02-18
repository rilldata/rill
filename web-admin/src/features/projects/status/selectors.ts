import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListResources,
  createRuntimeServicePing,
  createConnectorServiceOLAPListTables,
  createConnectorServiceOLAPGetTable,
  type V1ListResourcesResponse,
  type V1OlapTableInfo,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { readable, type Readable } from "svelte/store";
import { smartRefetchIntervalFunc } from "@rilldata/web-admin/lib/refetch-interval-store";

/** Type for the table metadata store result */
export type TableMetadataResult = {
  data: {
    isView: Map<string, boolean>;
  };
  isLoading: boolean;
  isError: boolean;
};

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

export function useTablesList(instanceId: string, connector: string = "") {
  return createConnectorServiceOLAPListTables(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId,
      },
    },
  );
}

/**
 * Fetches metadata (view status) for each table.
 *
 * Note: This creates a separate query per table (N+1 pattern). This is acceptable here because:
 * 1. The OLAPGetTable API doesn't support batch requests
 * 2. Tables are typically few in number on the status page
 * 3. Queries are cached and run in parallel via svelte-query
 *
 * If performance becomes an issue with many tables, consider adding a batch API endpoint.
 */
export function useTableMetadata(
  instanceId: string,
  connector: string = "",
  tables: V1OlapTableInfo[] | undefined,
): Readable<TableMetadataResult> {
  // If no tables, return empty store immediately
  if (!tables || tables.length === 0) {
    return readable(
      {
        data: {
          isView: new Map<string, boolean>(),
        },
        isLoading: false,
        isError: false,
      },
      () => {},
    );
  }

  return readable<TableMetadataResult>(
    {
      data: {
        isView: new Map<string, boolean>(),
      },
      isLoading: true,
      isError: false,
    },
    (set) => {
      const isView = new Map<string, boolean>();
      const tableNames = (tables ?? [])
        .map((t) => t.name)
        .filter((n) => !!n) as string[];
      const subscriptions: Array<() => void> = [];

      const completedTables = new Set<string>();
      const erroredTables = new Set<string>();
      const totalOperations = tableNames.length;

      // Helper to update and notify
      const updateAndNotify = () => {
        const isLoading = completedTables.size < totalOperations;
        set({
          data: { isView },
          isLoading,
          isError: erroredTables.size > 0,
        });
      };

      // Fetch view status for each table in parallel
      for (const tableName of tableNames) {
        const tableQuery = createConnectorServiceOLAPGetTable(
          {
            instanceId,
            connector,
            table: tableName,
          },
          {
            query: {
              enabled: !!instanceId && !!tableName,
            },
          },
        );

        const unsubscribe = tableQuery.subscribe((result) => {
          // Capture the view field from the response
          if (result.data?.view !== undefined) {
            isView.set(tableName, result.data.view);
          }
          // Track errors
          if (result.isError) {
            erroredTables.add(tableName);
          }
          // Only mark complete when the query has finished loading
          if (!result.isLoading) {
            completedTables.add(tableName);
          }
          updateAndNotify();
        });

        subscriptions.push(unsubscribe);
      }

      // Return cleanup function
      return () => {
        subscriptions.forEach((unsub) => unsub());
      };
    },
  );
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
