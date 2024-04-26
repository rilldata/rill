import {
  createConnectorServiceOLAPListTables,
  createRuntimeServiceGetInstance,
} from "../../runtime-client";
import { OLAP_DRIVERS_WITHOUT_MODELING } from "./olap-config";

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
  connectorInstanceId: string | undefined,
  olapConnector: string | undefined,
) {
  return createConnectorServiceOLAPListTables(
    {
      instanceId: connectorInstanceId,
      connector: olapConnector,
    },
    {
      query: {
        enabled: !!connectorInstanceId && !!olapConnector,
      },
    },
  );
}
