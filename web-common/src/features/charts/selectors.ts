import { getRouteFromName } from "@rilldata/web-common/features/entity-management/entity-mappers";
import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { useMainEntityFiles } from "../entity-management/file-selectors";

export function useChartFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "charts");
}

export function useChartRoutes(instanceId: string) {
  return useMainEntityFiles(instanceId, "charts", (name) =>
    getRouteFromName(name, EntityType.Chart),
  );
}

export const useChart = (instanceId: string, chartName: string) => {
  return useResource(instanceId, chartName, ResourceKind.Chart);
};
