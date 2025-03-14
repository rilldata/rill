import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import {
  type V1TableInfo,
  createConnectorServiceOLAPListTables,
  createRuntimeServiceAnalyzeConnectors,
  createRuntimeServiceGetInstance,
} from "../../../runtime-client";
import { featureFlags } from "../../feature-flags";
import { OLAP_DRIVERS_WITHOUT_MODELING } from "./olap-config";

export function useIsModelingSupportedForOlapDriver(
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

export function useIsModelingSupportedForDefaultOlapDriver(instanceId: string) {
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
              ?.map((table) => table.database)
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
