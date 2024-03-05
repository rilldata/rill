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
