import { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import { TableInfo } from "../../../proto/gen/rill/runtime/v1/connectors_pb";
import {
  V1TableInfo,
  createConnectorServiceOLAPListTables,
  createRuntimeServiceAnalyzeConnectors,
  createRuntimeServiceGetInstance,
} from "../../../runtime-client";
import { OLAP_DRIVERS_WITHOUT_MODELING } from "./olap-config";

export function useIsModelingSupportedForCurrentOlapDriver(instanceId: string) {
  return derived(
    [
      createRuntimeServiceGetInstance(instanceId, { sensitive: true }),
      createRuntimeServiceAnalyzeConnectors(instanceId),
    ],
    ([$instanceQuery, $connectorsQuery]) => {
      const { instance: { olapConnector: olapConnectorName = "" } = {} } =
        $instanceQuery.data || {};
      const { connectors = [] } = $connectorsQuery.data || {};

      const olapConnector = connectors.find(
        (connector) => connector.name === olapConnectorName,
      );

      const olapDriverName = olapConnector?.driver?.name ?? "";
      return !OLAP_DRIVERS_WITHOUT_MODELING.includes(olapDriverName);
    },
  );
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
              ?.map((tableInfo: TableInfo) => tableInfo.database)
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
): CreateQueryResult<V1TableInfo[]> {
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
