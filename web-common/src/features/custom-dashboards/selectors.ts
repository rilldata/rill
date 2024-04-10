import { getRouteFromName } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { useMainEntityFiles } from "../entity-management/file-selectors";
import {
  ResourceKind,
  useResource,
} from "../entity-management/resource-selectors";

export function useCustomDashboardFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "custom-dashboards");
}

export function useCustomDashboard(instanceId: string, name: string) {
  return useResource(instanceId, name, ResourceKind.Dashboard);
}

export function useCustomDashboardRoutes(instanceId: string) {
  return useMainEntityFiles(instanceId, "custom-dashboards", (name) =>
    getRouteFromName(name, EntityType.Dashboard),
  );
}
