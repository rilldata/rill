import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListResources,
  createRuntimeServicePing,
  createConnectorServiceOLAPListTables,
  createConnectorServiceOLAPGetTable,
  createQueryServiceTableCardinality,
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
    columnCount: Map<string, number>;
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

export function useResources(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    {},
    {
      query: {
        select: (data: V1ListResourcesResponse) => {
          const filtered = data?.resources?.filter(
            (resource) =>
              !resource?.meta?.hidden &&
              resource?.meta?.name?.kind !== ResourceKind.ProjectParser &&
              resource?.meta?.name?.kind !== ResourceKind.RefreshTrigger &&
              resource?.meta?.name?.kind !== ResourceKind.Component &&
              resource?.meta?.name?.kind !== ResourceKind.Migration,
          );
          return {
            ...data,
            resources: filtered,
          };
        },
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
          columnCount: new Map<string, number>(),
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
        columnCount: new Map<string, number>(),
      },
      isLoading: true,
      isError: false,
    },
    (set) => {
      const isView = new Map<string, boolean>();
      const columnCount = new Map<string, number>();
      const tableNames = (tables ?? [])
        .map((t) => t.name)
        .filter((n) => !!n) as string[];
      const subscriptions: Array<() => void> = [];

      const completedTables = new Set<string>();
      const totalOperations = tableNames.length;

      // Helper to update and notify
      const updateAndNotify = () => {
        const isLoading = completedTables.size < totalOperations;
        set({
          data: { isView, columnCount },
          isLoading,
          isError: false,
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
          // Capture the column count from the schema
          if (result.data?.schema?.fields !== undefined) {
            columnCount.set(tableName, result.data.schema.fields.length);
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

/** Type for the table cardinality store result */
export type TableCardinalityResult = {
  data: {
    rowCount: Map<string, number>;
  };
  isLoading: boolean;
  isError: boolean;
};

/**
 * Fetches row count (cardinality) for each table.
 */
export function useTableCardinality(
  instanceId: string,
  tables: V1OlapTableInfo[] | undefined,
): Readable<TableCardinalityResult> {
  // If no tables, return empty store immediately
  if (!tables || tables.length === 0) {
    return readable(
      {
        data: {
          rowCount: new Map<string, number>(),
        },
        isLoading: false,
        isError: false,
      },
      () => {},
    );
  }

  return readable<TableCardinalityResult>(
    {
      data: {
        rowCount: new Map<string, number>(),
      },
      isLoading: true,
      isError: false,
    },
    (set) => {
      const rowCount = new Map<string, number>();
      const tableNames = (tables ?? [])
        .map((t) => t.name)
        .filter((n) => !!n) as string[];
      const subscriptions: Array<() => void> = [];

      const completedTables = new Set<string>();
      const totalOperations = tableNames.length;

      // Helper to update and notify
      const updateAndNotify = () => {
        const isLoading = completedTables.size < totalOperations;
        set({
          data: { rowCount },
          isLoading,
          isError: false,
        });
      };

      // Fetch cardinality for each table in parallel
      for (const tableName of tableNames) {
        const cardinalityQuery = createQueryServiceTableCardinality(
          instanceId,
          tableName,
          {},
          {
            query: {
              enabled: !!instanceId && !!tableName,
            },
          },
        );

        const unsubscribe = cardinalityQuery.subscribe((result) => {
          if (result.data?.cardinality !== undefined) {
            rowCount.set(tableName, parseInt(result.data.cardinality, 10) || 0);
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
 * Fetches model resources and maps them by their result table name.
 * This allows looking up model resource data by the OLAP table name.
 */
export function useModelResources(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    { kind: ResourceKind.Model },
    {
      query: {
        select: (data: V1ListResourcesResponse) => {
          const map = new Map<string, V1Resource>();
          data.resources?.forEach((resource) => {
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
        },
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
