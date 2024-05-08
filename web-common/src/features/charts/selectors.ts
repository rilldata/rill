import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";

export const useChart = (instanceId: string, chartName: string) => {
  return useResource(instanceId, chartName, ResourceKind.Component);
};
