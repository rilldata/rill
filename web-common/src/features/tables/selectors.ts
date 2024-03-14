import { Readable, derived } from "svelte/store";
import {
  createConnectorServiceOLAPListTables,
  createRuntimeServiceGetInstance,
} from "../../runtime-client";
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

export function useTableNames(
  runtimeInstanceId: string,
  connectorInstanceId: string | undefined,
  olapConnector: string | undefined,
): Readable<string[]> {
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
    ([$sources, $models, $tables]) => {
      if (
        !$sources.data ||
        !$models.data ||
        !$tables.data ||
        !$tables.data.tables
      ) {
        return [];
      }

      // Filter out Rill-managed tables (Sources and Models)
      const sourceNames = $sources.data;
      const modelNames = $models.data;
      const filteredTables = $tables.data.tables?.filter(
        (table) =>
          !sourceNames.includes(table.name as string) &&
          !modelNames.includes(table.name as string),
      );

      // Return the fully qualified table names
      return (
        filteredTables?.map((table) => table.database + "." + table.name) || []
      );
    },
  );
}
