import { useMainEntityFiles } from "../entity-management/file-selectors";

export function useChartFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "charts");
}
