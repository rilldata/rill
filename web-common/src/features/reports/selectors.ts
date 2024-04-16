import { useMainEntityFiles } from "@rilldata/web-common/features/entity-management/file-selectors";

export function useReportFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "reports");
}
