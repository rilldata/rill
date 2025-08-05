import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import {
  type V1TableInfo,
  type V1OlapTableInfo,
  createConnectorServiceListDatabaseSchemas,
  createConnectorServiceListTables,
  createConnectorServiceGetTable,
  createConnectorServiceOLAPListTables,
  createRuntimeServiceGetInstance,
  type V1GetResourceResponse,
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
  type RpcStatus,
  type V1ConnectorSpec,
} from "../../runtime-client";
import { ResourceKind } from "../entity-management/resource-selectors";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";

// Helper function to determine if a connector prefers SQL-based modeling
// Connectors that prefer SQL models include:
// - DuckDB (embedded database with full SQL support)
// - Provisioned (managed) connectors
// - Read-write mode connectors
function prefersSqlBasedModeling(spec: V1ConnectorSpec) {
  return (
    spec.driver === "duckdb" ||
    spec.provision === true ||
    spec.properties?.mode === "readwrite"
  );
}

/**
 * Creates query options for checking modeling support of a connector
 */
function createModelingSupportQueryOptions(
  instanceId: string,
  connectorName: string,
  modelingType: "sql" | "yaml",
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

      if (modelingType === "sql") {
        // SQL-based modeling is preferred for connectors that prefer SQL models
        return prefersSqlBasedModeling(spec);
      } else {
        // YAML-based modeling is preferred for connectors that prefer YAML models
        // These are typically external databases (Postgres, MySQL, etc.)
        return !prefersSqlBasedModeling(spec);
      }
    },
  };
}

/**
 * LEGACY OLAP SELECTORS
 * These use the legacy OLAP-specific APIs and should be migrated to the new generic APIs above
 */

/**
 * Check if SQL-based modeling is preferred for a specific connector
 * Note: When modeling is supported, users can use either SQL or YAML
 * This function determines which approach is preferred for the connector
 */
export function usePrefersSqlBasedModelingForConnector(
  instanceId: string,
  connectorName: string,
): CreateQueryResult<boolean, ErrorType<RpcStatus>> {
  return createQuery(
    createModelingSupportQueryOptions(instanceId, connectorName, "sql"),
  );
}

/**
 * Check if YAML-based modeling is preferred for a specific connector
 * Note: When modeling is supported, users can use either SQL or YAML
 * This function determines which approach is preferred for the connector
 */
export function usePrefersYamlBasedModelingForConnector(
  instanceId: string,
  connectorName: string,
): CreateQueryResult<boolean, ErrorType<RpcStatus>> {
  return createQuery(
    createModelingSupportQueryOptions(instanceId, connectorName, "yaml"),
  );
}

export function usePrefersSqlBasedModelingForDefaultOlapDriver(
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
      "sql",
    );
  });

  return createQuery(queryOptions);
}

// Legacy alias for backward compatibility
export const useIsModelingSupportedForDefaultOlapDriverOLAP =
  usePrefersSqlBasedModelingForDefaultOlapDriver;

export function useDatabasesOLAP(instanceId: string, connector: string) {
  return createConnectorServiceOLAPListTables(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          // Get the unique databases
          return (
            data.tables
              ?.map((tableInfo) => tableInfo.database ?? "")
              .filter((value, index, self) => self.indexOf(value) === index) ??
            []
          );
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
  return createConnectorServiceOLAPListTables(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          return (
            data.tables
              ?.filter((table) => table.database === database)
              .map((table) => table.databaseSchema)
              .filter((value, index, self) => self.indexOf(value) === index) ??
            []
          );
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
): CreateQueryResult<V1OlapTableInfo[]> {
  return createConnectorServiceOLAPListTables(
    {
      instanceId,
      connector,
    },
    {
      query: {
        enabled: !!instanceId && !!connector,
        select: (data) => {
          return (
            data.tables?.filter(
              (table) =>
                table.database === database &&
                table.databaseSchema === databaseSchema,
            ) ?? []
          );
        },
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
 * Fetches schemas for a specific database
 */
export function useSchemasForDatabase(
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
