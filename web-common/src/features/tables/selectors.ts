import { debounce } from "@rilldata/web-common/lib/create-debouncer";
import { Readable, derived } from "svelte/store";
import {
  V1TableInfo,
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

export function useTables(
  runtimeInstanceId: string,
  connectorInstanceId: string | undefined,
  olapConnector: string | undefined,
): Readable<V1TableInfo[]> {
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
      // add a debounce to make sure quick changes and invalidations do not cause a flicker.
      // these quick changes can happen when a table/model is renamed
      debounce(() => {
        if (
          !$sources.data ||
          !$models.data ||
          !$tables.data ||
          !$tables.data.tables
        ) {
          set([]);
          return;
        }

        // Filter out Rill-managed tables (Sources and Models)
        const sourceNames = $sources.data;
        const modelNames = $models.data;
        const filteredTables = $tables.data.tables?.filter(
          (table) =>
            !sourceNames.includes(table.name as string) &&
            !modelNames.includes(table.name as string),
        );

        set(filteredTables);
      }, 500);
    },
  );
}
