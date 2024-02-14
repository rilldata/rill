import { createRuntimeServiceGetInstance } from "../../runtime-client";
import { OLAP_DRIVERS_WITHOUT_MODELING } from "./config";

export function useIsModelingSupportedForOlapDriver(instanceId: string) {
  return createRuntimeServiceGetInstance(instanceId, {
    query: {
      select: (data) => {
        const olapConnector = data.instance?.olapConnector as string;
        return !OLAP_DRIVERS_WITHOUT_MODELING.includes(olapConnector);
      },
    },
  });
}
