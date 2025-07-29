import type { CreateQueryResult } from "@tanstack/svelte-query";
import { createQuery } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import {
  createConnectorServiceOLAPListTables,
  createRuntimeServiceGetInstance,
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
  type RpcStatus,
  type V1GetResourceResponse,
  type V1OlapTableInfo,
} from "../../../runtime-client";
import type { ErrorType } from "../../../runtime-client/http-client";
import { ResourceKind } from "../../entity-management/resource-selectors";

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
 * Check if modeling is supported for a specific connector based on its properties
 */
export function useIsModelingSupportedForConnector(
  instanceId: string,
  connectorName: string,
): CreateQueryResult<boolean, ErrorType<RpcStatus>> {
  return createQuery(
    createModelingSupportQueryOptions(instanceId, connectorName),
  );
}

/**
 * Check if modeling is supported for the default OLAP connector
 */
export function useIsModelingSupportedForDefaultOlapDriver(
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

export function useDatabases(instanceId: string, connector: string) {
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

export function useDatabaseSchemas(
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

export function useTables(
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
