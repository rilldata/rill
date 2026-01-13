import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListResources,
  createConnectorServiceOLAPListTables,
  createConnectorServiceOLAPGetTable,
  type V1ListResourcesResponse,
  type V1Resource,
  type V1OlapTableInfo,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { createSmartRefetchInterval } from "@rilldata/web-admin/lib/refetch-interval-store";
import { readable } from "svelte/store";
import { httpClient } from "@rilldata/web-common/runtime-client/http-client";

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
        refetchInterval: createSmartRefetchInterval,
      },
    },
  );
}

// Cache stores by instanceId and connector array to prevent recreating them
const modelSizesStoreCache = new Map<
  string,
  { store: any; unsubscribe: () => void }
>();

// Keep preloaded query subscriptions alive so they don't get cancelled
const preloadedQuerySubscriptions = new Map<string, Set<() => void>>();

// Preload queries to ensure they start immediately and keep them alive
function preloadConnectorQueries(instanceId: string, connectorArray: string[]) {
  const preloadKey = `${instanceId}:${connectorArray.join(",")}`;

  // Only preload once per connector set
  if (preloadedQuerySubscriptions.has(preloadKey)) {
    return;
  }

  const subscriptions = new Set<() => void>();

  for (const connector of connectorArray) {
    const query = createConnectorServiceOLAPListTables(
      {
        instanceId,
        connector,
      },
      {
        query: {
          enabled: true,
        },
      },
    );

    // Eagerly subscribe to keep the query alive
    const unsubscribe = query.subscribe(() => {});
    subscriptions.add(unsubscribe);
  }

  preloadedQuerySubscriptions.set(preloadKey, subscriptions);
}

function createCachedStore(
  cacheKey: string,
  instanceId: string,
  connectorArray: string[],
) {
  // Check if we already have a cached store
  if (modelSizesStoreCache.has(cacheKey)) {
    return modelSizesStoreCache.get(cacheKey)!.store;
  }

  // Preload queries immediately so they start running before store subscribers attach
  preloadConnectorQueries(instanceId, connectorArray);

  // If no connectors, return an empty readable store
  if (connectorArray.length === 0) {
    const emptyStore = readable(
      {
        data: new Map<string, string | number>(),
        isLoading: false,
        isError: false,
      },
      () => {},
    );
    modelSizesStoreCache.set(cacheKey, {
      store: emptyStore,
      unsubscribe: () => {},
    });
    return emptyStore;
  }

  // Create a new store with pagination support
  const store = readable(
    {
      data: new Map<string, string | number>(),
      isLoading: true,
      isError: false,
    },
    (set) => {
      const connectorTables = new Map<string, Array<any>>();
      const connectorLoading = new Map<string, boolean>();
      const connectorError = new Map<string, boolean>();
      const subscriptions = new Set<() => void>();

      const updateAndNotify = () => {
        const sizeMap = new Map<string, string | number>();
        let isLoading = false;
        let isError = false;

        for (const connector of connectorArray) {
          if (connectorLoading.get(connector)) isLoading = true;
          if (connectorError.get(connector)) isError = true;

          for (const table of connectorTables.get(connector) || []) {
            if (
              table.name &&
              table.physicalSizeBytes !== undefined &&
              table.physicalSizeBytes !== null
            ) {
              const key = `${connector}:${table.name}`;
              sizeMap.set(key, table.physicalSizeBytes as string | number);
            }
          }
        }

        set({ data: sizeMap, isLoading, isError });
      };

      const fetchPage = (connector: string, pageToken?: string) => {
        const query = createConnectorServiceOLAPListTables(
          {
            instanceId,
            connector,
            ...(pageToken && { pageToken }),
          } as any,
          {
            query: {
              enabled: true,
            },
          },
        );

        const unsubscribe = query.subscribe((result: any) => {
          connectorLoading.set(connector, result.isLoading);
          connectorError.set(connector, result.isError);

          if (result.data?.tables) {
            const existing = connectorTables.get(connector) || [];
            connectorTables.set(connector, [
              ...existing,
              ...result.data.tables,
            ]);
          }

          // If query completed and has more pages, fetch the next page
          if (!result.isLoading && result.data?.nextPageToken) {
            unsubscribe();
            subscriptions.delete(unsubscribe);
            fetchPage(connector, result.data.nextPageToken);
          }

          updateAndNotify();
        });

        subscriptions.add(unsubscribe);
      };

      // Start fetching for all connectors
      for (const connector of connectorArray) {
        connectorLoading.set(connector, true);
        connectorError.set(connector, false);
        connectorTables.set(connector, []);
        fetchPage(connector);
      }

      return () => {
        for (const unsub of subscriptions) {
          unsub();
        }
      };
    },
  );

  // Eagerly subscribe to keep queries alive across component re-renders
  const unsubscribe = store.subscribe(() => {});
  modelSizesStoreCache.set(cacheKey, { store, unsubscribe });

  return store;
}

export function useModelTableSizes(
  instanceId: string,
  resources: V1Resource[] | undefined,
) {
  // Extract unique connectors from model resources
  const uniqueConnectors = new Set<string>();

  if (resources) {
    for (const resource of resources) {
      if (resource?.meta?.name?.kind === ResourceKind.Model) {
        const connector = resource.model?.state?.resultConnector;
        const table = resource.model?.state?.resultTable;

        if (connector && table) {
          uniqueConnectors.add(connector);
        }
      }
    }
  }

  const connectorArray = Array.from(uniqueConnectors).sort();
  const cacheKey = `${instanceId}:${connectorArray.join(",")}`;

  return createCachedStore(cacheKey, instanceId, connectorArray);
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
        refetchInterval: createSmartRefetchInterval,
      },
    },
  );
}

export async function fetchRowCount(
  instanceId: string,
  tableName: string,
): Promise<number | "error"> {
  try {
    console.log(`[RowCount] Fetching count for ${tableName}...`);

    const data = await httpClient<{ data: any[] }>({
      url: `/v1/instances/${instanceId}/query`,
      method: "POST",
      headers: { "Content-Type": "application/json" },
      data: {
        sql: `SELECT COUNT(*) as count FROM "${tableName}"`,
      },
    });

    console.log(`[RowCount] ${tableName} response:`, data);

    if (data?.data && Array.isArray(data.data) && data.data.length > 0) {
      const firstRow = data.data[0] as any;
      const count = parseInt(String(firstRow?.count ?? 0), 10);
      console.log(`[RowCount] ${tableName} success - count:`, count);
      return isNaN(count) ? "error" : count;
    }

    console.warn(`[RowCount] ${tableName} unexpected response structure:`, data);
    return "error";
  } catch (error: any) {
    console.error(`[RowCount] ${tableName} error:`, error);
    return "error";
  }
}

export function useTableMetadata(
  instanceId: string,
  connector: string = "",
  tables: V1OlapTableInfo[] | undefined,
) {
  // If no tables, return empty store immediately
  if (!tables || tables.length === 0) {
    return readable(
      {
        data: { columnCounts: new Map<string, number>(), rowCounts: new Map<string, number | "error">(), isView: new Map<string, boolean>() },
        isLoading: false,
        isError: false,
      },
      () => {},
    );
  }

  return readable(
    {
      data: { columnCounts: new Map<string, number>(), rowCounts: new Map<string, number | "error">(), isView: new Map<string, boolean>() },
      isLoading: true,
      isError: false,
    },
    (set) => {
      const columnCounts = new Map<string, number>();
      const rowCounts = new Map<string, number | "error">();
      const isView = new Map<string, boolean>();
      const tableNames = (tables ?? []).map((t) => t.name).filter((n) => !!n) as string[];
      const subscriptions: Array<() => void> = [];

      let completedCount = 0;
      const totalOperations = tableNames.length; // Only column counts; row counts fetched at component level

      // Helper to update and notify
      const updateAndNotify = () => {
        const isLoading = completedCount < totalOperations;
        set({
          data: { columnCounts, rowCounts, isView },
          isLoading,
          isError: false,
        });
      };

      // Fetch column counts and row counts for each table
      for (const tableName of tableNames) {
        // Fetch column count and view status
        const columnQuery = createConnectorServiceOLAPGetTable(
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

        const columnUnsubscribe = columnQuery.subscribe((result) => {
          if (result.data?.schema?.fields) {
            columnCounts.set(tableName, result.data.schema.fields.length);
          }
          // Capture the view field from the response
          if (result.data?.view !== undefined) {
            isView.set(tableName, result.data.view);
          }
          completedCount++;
          updateAndNotify();
        });

        subscriptions.push(columnUnsubscribe);

        // Initialize row count as not yet fetched
        // Row counts will be fetched separately at the component level where JWT is guaranteed ready
        completedCount++;
        updateAndNotify();
      }

      // Return cleanup function
      return () => {
        subscriptions.forEach((unsub) => unsub());
      };
    },
  );
}
