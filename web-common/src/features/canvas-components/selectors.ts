import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";

export const useComponent = (instanceId: string, componentName: string) => {
  return useResource(instanceId, componentName, ResourceKind.Component);
};
