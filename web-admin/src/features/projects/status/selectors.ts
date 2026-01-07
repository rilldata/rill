import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListResources,
  createConnectorServiceOLAPListTables,
  type V1ListResourcesResponse,
  type V1Resource,
  type V1OLAPListTablesResponse,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { createSmartRefetchInterval } from "@rilldata/web-admin/lib/refetch-interval-store";
import { derived, readable } from "svelte/store";

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

// Cache for connector queries to avoid recreating them
const connectorQueryCache = new Map<
  string,
  ReturnType<typeof createConnectorServiceOLAPListTables>
>();

function getConnectorQuery(instanceId: string, connector: string) {
  const cacheKey = `${instanceId}:${connector}`;
  if (!connectorQueryCache.has(cacheKey)) {
    const query = createConnectorServiceOLAPListTables(
      {
        instanceId,
        connector,
      },
      {
        query: {
          enabled: !!instanceId && !!connector,
        },
      },
    );
    connectorQueryCache.set(cacheKey, query);
  }
  return connectorQueryCache.get(cacheKey)!;
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

  // If no connectors, return an empty readable store
  if (connectorArray.length === 0) {
    return readable(
      {
        data: new Map<string, string | number>(),
        isLoading: false,
        isError: false,
      },
      () => {},
    );
  }

  // Use a readable store with custom subscription logic to handle pagination
  return readable(
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
            connectorTables.set(connector, [...existing, ...result.data.tables]);
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
}
