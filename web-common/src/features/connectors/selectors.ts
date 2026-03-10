import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import {
  type V1TableInfo,
  type V1GetResourceResponse,
  createRuntimeServiceAnalyzeConnectors,
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  runtimeServiceAnalyzeConnectors,
} from "../../runtime-client";
import type { RuntimeClient } from "../../runtime-client/v2";
import { isNotFoundError } from "../../lib/errors";
import {
  createRuntimeServiceGetInstance,
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
  createConnectorServiceListDatabaseSchemas,
  createConnectorServiceGetTable,
  createConnectorServiceListTablesInfinite,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "../entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

/**
 * Creates query options for checking modeling support of a connector
 */
function createModelingSupportQueryOptions(
  client: RuntimeClient,
  connectorName: string,
) {
  const instanceId = client.instanceId;
  return {
    queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
      name: { kind: ResourceKind.Connector, name: connectorName },
    }),
    queryFn: async () => {
      try {
        return await runtimeServiceGetResource(client, {
          name: { kind: ResourceKind.Connector, name: connectorName },
        });
      } catch (error) {
        // Handle legacy DuckDB projects where no explicit connector resource exists
        if (connectorName === "duckdb" && isNotFoundError(error)) {
          // Return a synthetic DuckDB connector
          return {
            resource: {
              connector: {
                spec: {
                  driver: "duckdb",
                },
              },
            },
          };
        }
        throw error;
      }
    },
    enabled: !!instanceId && !!connectorName,
    select: (data: V1GetResourceResponse) => {
      const spec = data?.resource?.connector?.spec;
      if (!spec) return false;

      // Modeling is supported if:
      // - DuckDB (embedded database with full SQL support)
      // - Provisioned (managed) connectors
      // - Read-write mode connectors
      return (
        spec.driver === "duckdb" ||
        spec.provision === true ||
        spec.properties?.mode === "readwrite"
      );
    },
  };
}

/**
 * Check if modeling is supported for a specific connector based on its properties
 */
export function useIsModelingSupportedForConnectorOLAP(
  client: RuntimeClient,
  connectorName: string,
): CreateQueryResult<boolean, Error> {
  return createQuery(createModelingSupportQueryOptions(client, connectorName));
}

export function useIsModelingSupportedForDefaultOlapDriverOLAP(
  client: RuntimeClient,
): CreateQueryResult<boolean, Error> {
  const instanceQuery = createRuntimeServiceGetInstance(client, {
    sensitive: true,
  });

  // Create queryOptions store that includes the dynamic connector name
  const queryOptions = derived([instanceQuery], ([$instanceQuery]) => {
    const olapConnectorName = $instanceQuery.data?.instance?.olapConnector;
    return createModelingSupportQueryOptions(client, olapConnectorName || "");
  });

  return createQuery(queryOptions);
}

/**
 * List databases (when `database` is undefined) or schemas for a given database (when provided).
 * The backend returns all schemas across databases; filtering is applied client-side.
 */
export function useListDatabaseSchemas(
  client: RuntimeClient,
  connector: string,
  database?: string,
  enabled: boolean = true,
) {
  return createConnectorServiceListDatabaseSchemas(
    client,
    {
      connector,
    },
    {
      query: {
        enabled: !!client.instanceId && !!connector && enabled,
        select: (data) => {
          const allSchemas = data.databaseSchemas ?? [];

          if (database !== undefined) {
            const hasEmptyDatabases = allSchemas.every((s) => !s.database);
            return hasEmptyDatabases
              ? [database]
              : allSchemas
                  .filter((s) => s.database === database)
                  .map((s) => s.databaseSchema ?? "");
          }

          // Derive databases (top-level)
          const hasEmptyDatabases = allSchemas.every(
            (schema) => !schema.database,
          );
          return hasEmptyDatabases
            ? Array.from(
                new Set(
                  allSchemas.map((schema) => schema.databaseSchema ?? ""),
                ),
              )
            : Array.from(
                new Set(allSchemas.map((schema) => schema.database ?? "")),
              ).filter(Boolean);
        },
      },
    },
  );
}

/**
 * Infinite tables loader using pageToken cursor
 */
export function useInfiniteListTables(
  client: RuntimeClient,
  connector: string,
  database: string,
  databaseSchema: string,
  pageSize = 5,
  enabled: boolean = true,
) {
  return createConnectorServiceListTablesInfinite(
    client,
    { connector, database, databaseSchema, pageSize },
    {
      query: {
        enabled:
          enabled &&
          !!client.instanceId &&
          !!connector &&
          (!!database || database === "") &&
          databaseSchema !== undefined,
        select: (data) => ({
          tables: data.pages.flatMap(
            (p) => (p as { tables?: V1TableInfo[] }).tables ?? [],
          ),
          nextPageToken:
            data.pages.length > 0
              ? (
                  data.pages[data.pages.length - 1] as {
                    nextPageToken?: string;
                  }
                ).nextPageToken
              : undefined,
        }),
      },
    },
  );
}

/**
 * Get metadata about a table or view
 * Called when a table is selected/expanded
 */
export function useGetTable(
  client: RuntimeClient,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
) {
  return createConnectorServiceGetTable(
    client,
    {
      connector,
      database,
      databaseSchema,
      table,
    },
    {
      query: {
        enabled:
          !!client.instanceId &&
          !!connector &&
          !!table &&
          database !== undefined &&
          databaseSchema !== undefined,
      },
    },
  );
}

export function getAnalyzedConnectors(
  client: RuntimeClient,
  olapOnly: boolean,
) {
  return createRuntimeServiceAnalyzeConnectors(
    client,
    {},
    {
      query: {
        // Retry transient errors during runtime resets (e.g. project initialization)
        retry: (failureCount) => failureCount < 3,
        retryDelay: 1000,
        // sort alphabetically
        select: (data) => {
          if (!data?.connectors) return;

          const filtered = (
            olapOnly
              ? data.connectors.filter((c) => c?.driver?.implementsOlap)
              : data.connectors
          ).sort((a, b) =>
            (a?.name as string).localeCompare(b?.name as string),
          );
          return { connectors: filtered };
        },
      },
    },
  );
}

export async function fetchAnalyzeConnectors(client: RuntimeClient) {
  const queryKey = getRuntimeServiceAnalyzeConnectorsQueryKey(
    client.instanceId,
  );
  const resp = await queryClient.fetchQuery({
    queryKey,
    queryFn: () => runtimeServiceAnalyzeConnectors(client, {}),
  });
  return resp?.connectors ?? [];
}
