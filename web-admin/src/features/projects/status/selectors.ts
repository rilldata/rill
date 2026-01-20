import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListResources,
  createConnectorServiceOLAPListTables,
  createConnectorServiceOLAPGetTable,
  type V1ListResourcesResponse,
  type V1OlapTableInfo,
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

export function useResources(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    {},
    {
      query: {
        select: (data: V1ListResourcesResponse) => {
          const filtered = data?.resources?.filter(
            (resource) =>
              resource?.meta?.name?.kind !== ResourceKind.ProjectParser &&
              resource?.meta?.name?.kind !== ResourceKind.RefreshTrigger,
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
        },
        isLoading: false,
        isError: false,
      },
      () => {},
    );
  }

  return readable(
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

      let completedCount = 0;
      const totalOperations = tableNames.length;

      // Helper to update and notify
      const updateAndNotify = () => {
        const isLoading = completedCount < totalOperations;
        set({
          data: { isView },
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
          completedCount++;
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
