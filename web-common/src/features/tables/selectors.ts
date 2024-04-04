import { debounce } from "@rilldata/web-common/lib/create-debouncer";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import { Readable, derived } from "svelte/store";
import {
  V1OLAPListTablesResponse,
  V1TableInfo,
  createConnectorServiceOLAPListTables,
  createRuntimeServiceGetInstance,
} from "../../runtime-client";
import type { HTTPError } from "../../runtime-client/fetchWrapper";
import {
  ResourceKind,
  useFilteredResourceNames,
} from "../entity-management/resource-selectors";
import { OLAP_DRIVERS_WITHOUT_MODELING } from "./config";

export function useIsModelingSupportedForCurrentOlapDriver(instanceId: string) {
  return createRuntimeServiceGetInstance(instanceId, {
    query: {
      select: (data) => {
        const olapConnector = data.instance?.olapConnector as string;
        return !OLAP_DRIVERS_WITHOUT_MODELING.includes(olapConnector);
      },
    },
  });
}

export function useTables(
  runtimeInstanceId: string,
  connectorInstanceId: string | undefined,
  olapConnector: string | undefined,
): Readable<V1TableInfo[]> {
  function setTables(
    sources: QueryObserverResult<string[], HTTPError>,
    models: QueryObserverResult<string[], HTTPError>,
    tables: QueryObserverResult<V1OLAPListTablesResponse, HTTPError>,
    set: (value: unknown) => void,
  ) {
    if (!sources.data || !models.data || !tables.data || !tables.data.tables) {
      set([]);
      return;
    }

    const sourceNames = sources.data;
    const modelNames = models.data;
    // Filter out Rill-managed tables (Sources and Models)
    const filteredTables = tables.data.tables.filter((table) => {
      const tableName = table.name as string;
      return (
        !sourceNames.includes(tableName) && !modelNames.includes(tableName)
      );
    });

    set(filteredTables);
  }

  // Debounce the table list calculation to make sure quick changes and invalidations do not cause a flicker.
  // These quick changes can happen when a table/model is renamed.
  const debouncedSetTables = debounce(setTables, 450);

  return derived(
    [
      useFilteredResourceNames(runtimeInstanceId, ResourceKind.Source),
      useFilteredResourceNames(runtimeInstanceId, ResourceKind.Model),
      createConnectorServiceOLAPListTables(
        {
          instanceId: connectorInstanceId,
          connector: olapConnector,
        },
        {
          query: {
            enabled: !!connectorInstanceId && !!olapConnector,
          },
        },
      ),
    ],
    ([$sources, $models, $tables], set) => {
      debouncedSetTables($sources, $models, $tables, set);
    },
  );
}

export function makeFullyQualifiedTableName(
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
) {
  switch (connector) {
    case "clickhouse":
      return `${databaseSchema}.${table}`;
    case "druid":
      return `${databaseSchema}.${table}`;
    case "duckdb":
      return `${database}.${databaseSchema}.${table}`;
    default:
      throw new Error(`Unsupported OLAP connector: ${connector}`);
  }
}
