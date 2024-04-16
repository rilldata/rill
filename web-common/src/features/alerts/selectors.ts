import { useMainEntityFiles } from "@rilldata/web-common/features/entity-management/file-selectors";

export function useAlertFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "alerts");
}
