import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import {
  type V1TableInfo,
  createConnectorServiceListDatabaseSchemas,
  createConnectorServiceListTables,
  createConnectorServiceGetTable,
  createRuntimeServiceAnalyzeConnectors,
} from "../../runtime-client";
export * from "./olap/selectors";

/**
 * NEW SELECTORS - Using the new granular APIs
 * These work with all connector types, not just OLAP
 */

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
  const databaseSchemasQuery = useDatabaseSchemas(instanceId, connector);

  return derived([databaseSchemasQuery], ([$query]) => {
    if ($query.isLoading)
      return { isLoading: true, data: undefined, error: undefined };
    if ($query.error)
      return { isLoading: false, data: undefined, error: $query.error };

    // Check if all databases are empty (flat schema structure like MySQL)
    const hasEmptyDatabases = $query.data?.every((schema) => !schema.database);

    const databases = hasEmptyDatabases
      ? // For flat structures, use databaseSchema as the primary level
        Array.from(
          new Set(
            $query.data?.map((schema) => schema.databaseSchema ?? "") ?? [],
          ),
        )
      : // For hierarchical structures, use database as the primary level
        Array.from(
          new Set($query.data?.map((schema) => schema.database ?? "") ?? []),
        ).filter(Boolean);

    return { isLoading: false, data: databases, error: undefined };
  });
}

/**
 * Fetches schemas for a specific database
 */
export function useSchemasForDatabase(
  instanceId: string,
  connector: string,
  database: string,
) {
  const databaseSchemasQuery = useDatabaseSchemas(instanceId, connector);

  return derived([databaseSchemasQuery], ([$query]) => {
    if ($query.isLoading)
      return { isLoading: true, data: undefined, error: undefined };
    if ($query.error)
      return { isLoading: false, data: undefined, error: $query.error };

    // Check if this is a flat schema structure (like MySQL)
    const hasEmptyDatabases = $query.data?.every((schema) => !schema.database);

    const schemas = hasEmptyDatabases
      ? // For flat structures, the "database" parameter is actually a schema name
        [database]
      : // For hierarchical structures, filter by actual database
        ($query.data
          ?.filter((schema) => schema.database === database)
          ?.map((schema) => schema.databaseSchema ?? "") ?? []);

    return { isLoading: false, data: schemas, error: undefined };
  });
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

/**
 * COMPATIBILITY LAYER
 * These adapt the new APIs to match the old API shapes for easier migration
 */

/**
 * Compatibility wrapper that mimics the old useDatabases behavior
 */
export function useDatabases(instanceId: string, connector: string) {
  return useDatabasesFromSchemas(instanceId, connector);
}

/**
 * Compatibility wrapper that mimics the old useDatabaseSchemas behavior
 */
export function useDatabaseSchemasForDatabase(
  instanceId: string,
  connector: string,
  database: string,
) {
  return useSchemasForDatabase(instanceId, connector, database);
}

/**
 * Compatibility wrapper that mimics the old useTables behavior
 * Note: This now returns V1TableInfo instead of V1OlapTableInfo
 */
export function useTables(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
) {
  return useTablesForSchema(instanceId, connector, database, databaseSchema);
}
