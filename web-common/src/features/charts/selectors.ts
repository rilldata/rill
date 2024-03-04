import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Chart } from "@rilldata/web-common/runtime-client";
import { useMainEntityFiles } from "../entity-management/file-selectors";

export function useChartFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "charts");
}

export const useChart = (instanceId: string, chartName: string) => {
  return useResource(instanceId, chartName, ResourceKind.Chart);
};

export const useChartView = <T = V1Chart>(
  instanceId: string,
  chartName: string,
  selector?: (meta: V1Chart) => T,
) => {
  return useResource<T>(instanceId, chartName, ResourceKind.Chart, (data) =>
    selector ? selector(data.chart) : (data.chart as T),
  );
};
