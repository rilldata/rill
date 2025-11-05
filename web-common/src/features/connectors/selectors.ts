import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import {
  type V1TableInfo,
  createConnectorServiceListDatabaseSchemas,
  createConnectorServiceListTables,
  createConnectorServiceGetTable,
  createRuntimeServiceGetInstance,
  type V1GetResourceResponse,
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
  type RpcStatus,
} from "../../runtime-client";
import { ResourceKind } from "../entity-management/resource-selectors";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";

/**
 * Creates query options for checking modeling support of a connector
 */
function createModelingSupportQueryOptions(
  instanceId: string,
  connectorName: string,
) {
  return {
    queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.kind": ResourceKind.Connector,
      "name.name": connectorName,
    }),
    queryFn: async () => {
      try {
        return await runtimeServiceGetResource(instanceId, {
          "name.kind": ResourceKind.Connector,
          "name.name": connectorName,
        });
      } catch (error) {
        // Handle legacy DuckDB projects where no explicit connector resource exists
        if (connectorName === "duckdb" && error?.response?.status === 404) {
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
 * LEGACY OLAP SELECTORS
 * These use the legacy OLAP-specific APIs and should be migrated to the new generic APIs above
 */

/**
 * Check if modeling is supported for a specific connector based on its properties
 */
export function useIsModelingSupportedForConnectorOLAP(
  instanceId: string,
  connectorName: string,
): CreateQueryResult<boolean, ErrorType<RpcStatus>> {
  return createQuery(
    createModelingSupportQueryOptions(instanceId, connectorName),
  );
}

export function useIsModelingSupportedForDefaultOlapDriverOLAP(
  instanceId: string,
): CreateQueryResult<boolean, ErrorType<RpcStatus>> {
  const instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });

  // Create queryOptions store that includes the dynamic connector name
  const queryOptions = derived([instanceQuery], ([$instanceQuery]) => {
    const olapConnectorName = $instanceQuery.data?.instance?.olapConnector;
    return createModelingSupportQueryOptions(
      instanceId,
      olapConnectorName || "",
    );
  });

  return createQuery(queryOptions);
}

export function useDatabasesOLAP(instanceId: string, connector: string) {
  return createConnectorServiceListDatabaseSchemas(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          const allSchemas = data.databaseSchemas ?? [];
          const hasEmptyDatabases = allSchemas.every((s) => !s.database);
          return hasEmptyDatabases
            ? Array.from(new Set(allSchemas.map((s) => s.databaseSchema ?? "")))
            : Array.from(
                new Set(allSchemas.map((s) => s.database ?? "")),
              ).filter(Boolean);
        },
      },
    },
  );
}

export function useDatabaseSchemasOLAP(
  instanceId: string,
  connector: string,
  database: string,
) {
  return createConnectorServiceListDatabaseSchemas(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          const allSchemas = data.databaseSchemas ?? [];
          const hasEmptyDatabases = allSchemas.every((s) => !s.database);
          return hasEmptyDatabases
            ? [database]
            : allSchemas
                .filter((s) => s.database === database)
                .map((s) => s.databaseSchema ?? "");
        },
      },
    },
  );
}

export function useTablesOLAP(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  enabled: boolean = true,
): CreateQueryResult<V1TableInfo[]> {
  return createConnectorServiceListTables(
    {
      instanceId,
      connector,
      database,
      databaseSchema,
    },
    {
      query: {
        enabled:
          enabled &&
          !!instanceId &&
          !!connector &&
          !!database &&
          databaseSchema !== undefined,
        select: (data) => data.tables ?? [],
      },
    },
  );
}

/**
 * Fetches database schemas for any connector type
 * Replaces the need to filter OLAPListTables client-side
 */
export function useDatabaseSchemas(instanceId: string, connector: string) {
  return createConnectorServiceListDatabaseSchemas(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => data.databaseSchemas ?? [],
      },
    },
  );
}

/**
 * Extracts unique databases from database schemas
 * More efficient than the old approach
 */
export function useDatabasesFromSchemas(instanceId: string, connector: string) {
  return createConnectorServiceListDatabaseSchemas(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          const allSchemas = data.databaseSchemas ?? [];

          // Check if all databases are empty (flat schema structure like MySQL)
          const hasEmptyDatabases = allSchemas.every(
            (schema) => !schema.database,
          );

          const databases = hasEmptyDatabases
            ? // For flat structures, use databaseSchema as the primary level
              Array.from(
                new Set(
                  allSchemas.map((schema) => schema.databaseSchema ?? ""),
                ),
              )
            : // For hierarchical structures, use database as the primary level
              Array.from(
                new Set(allSchemas.map((schema) => schema.database ?? "")),
              ).filter(Boolean);

          return databases;
        },
      },
    },
  );
}

/**
 * List all schemas across databases
 */
export function useListDatabaseSchemas(
  instanceId: string,
  connector: string,
  database: string,
) {
  return createConnectorServiceListDatabaseSchemas(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          const allSchemas = data.databaseSchemas ?? [];

          // Check if this is a flat schema structure (like MySQL)
          const hasEmptyDatabases = allSchemas.every(
            (schema) => !schema.database,
          );

          const schemas = hasEmptyDatabases
            ? // For flat structures, the "database" parameter is actually a schema name
              [database]
            : // For hierarchical structures, filter by actual database
              allSchemas
                .filter((schema) => schema.database === database)
                .map((schema) => schema.databaseSchema ?? "");

          return schemas;
        },
      },
    },
  );
}

/**
 * Fetches tables for a specific database and schema
 * This is called on-demand when a schema is expanded
 */
export function useTablesForSchema(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  enabled: boolean = true,
): CreateQueryResult<V1TableInfo[]> {
  return createConnectorServiceListTables(
    {
      instanceId,
      connector,
      database,
      databaseSchema,
    },
    {
      query: {
        enabled:
          enabled &&
          !!instanceId &&
          !!connector &&
          !!database &&
          databaseSchema !== undefined,
        select: (data) => data.tables ?? [],
      },
    },
  );
}

/**
 * Fetches detailed metadata for a specific table
 * Called when a table is selected/expanded
 */
export function useTableMetadata(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
) {
  return createConnectorServiceGetTable(
    {
      instanceId,
      connector,
      database,
      databaseSchema,
      table,
    },
    {
      query: {
        enabled:
          !!instanceId &&
          !!connector &&
          !!table &&
          database !== undefined &&
          databaseSchema !== undefined,
      },
    },
  );
}
