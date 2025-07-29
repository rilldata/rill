import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import {
  type V1TableInfo,
  type V1OlapTableInfo,
  createConnectorServiceListDatabaseSchemas,
  createConnectorServiceListTables,
  createConnectorServiceGetTable,
  createConnectorServiceOLAPListTables,
  createRuntimeServiceAnalyzeConnectors,
  createRuntimeServiceGetInstance,
} from "../../runtime-client";
import { featureFlags } from "../feature-flags";
import { OLAP_DRIVERS_WITHOUT_MODELING } from "./olap/olap-config";

/**
 * LEGACY OLAP SELECTORS
 * These use the legacy OLAP-specific APIs and should be migrated to the new generic APIs above
 */

export function useIsModelingSupportedForOlapDriverOLAP(
  instanceId: string,
  driver: string,
) {
  const { clickhouseModeling } = featureFlags;
  return derived(
    [createRuntimeServiceAnalyzeConnectors(instanceId), clickhouseModeling],
    ([$connectorsQuery, $clickhouseModeling]) => {
      const { connectors = [] } = $connectorsQuery.data || {};
      const olapConnector = connectors.find(
        (connector) => connector.name === driver,
      );
      const olapDriverName = olapConnector?.driver?.name ?? "";

      if (olapDriverName === "clickhouse") {
        return $clickhouseModeling;
      }

      return !OLAP_DRIVERS_WITHOUT_MODELING.includes(olapDriverName);
    },
  );
}

export function useIsModelingSupportedForDefaultOlapDriverOLAP(
  instanceId: string,
) {
  const { clickhouseModeling } = featureFlags;
  return derived(
    [
      createRuntimeServiceGetInstance(instanceId, { sensitive: true }),
      createRuntimeServiceAnalyzeConnectors(instanceId),
      clickhouseModeling,
    ],
    ([$instanceQuery, $connectorsQuery, $clickhouseModeling]) => {
      const { instance: { olapConnector: olapConnectorName = "" } = {} } =
        $instanceQuery.data || {};
      const { connectors = [] } = $connectorsQuery.data || {};

      const olapConnector = connectors.find(
        (connector) => connector.name === olapConnectorName,
      );

      const olapDriverName = olapConnector?.driver?.name ?? "";

      if (olapDriverName === "clickhouse") {
        return $clickhouseModeling;
      }

      return !OLAP_DRIVERS_WITHOUT_MODELING.includes(olapDriverName);
    },
  );
}

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
