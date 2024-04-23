import { useMainEntityFiles } from "@rilldata/web-common/features/entity-management/file-selectors";

export function useAPIFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "apis");
}
