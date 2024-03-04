import { useMainEntityFiles } from "../entity-management/file-selectors";

export function useCustomDashboardFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "custom-dashboards");
}
