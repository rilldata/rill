import type { CreateQueryResult } from "@tanstack/svelte-query";
import {
  type V1TableInfo,
  createConnectorServiceListDatabaseSchemas,
  createConnectorServiceListTables,
  createConnectorServiceGetTable,
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
