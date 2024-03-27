import { getRouteFromName } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { useMainEntityFiles } from "../entity-management/file-selectors";

export function useCustomDashboardFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "custom-dashboards");
}

export function useCustomDashboardRoutes(instanceId: string) {
  return useMainEntityFiles(instanceId, "custom-dashboards", (name) =>
    getRouteFromName(name, EntityType.Dashboard),
  );
}
