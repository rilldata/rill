import { useMainEntityFiles } from "@rilldata/web-common/features/entity-management/file-selectors";
import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";

export function useTheme(instanceId: string, name: string) {
  return useResource(instanceId, name, ResourceKind.Theme);
}

export function useThemeFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "themes");
}
