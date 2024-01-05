import { useMainEntityFiles } from "@rilldata/web-common/features/entity-management/file-selectors";
import {
  ResourceKind,
  useFilteredResourceNames,
  useFilteredResources,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewTimeRange,
  V1MetricsViewFilter,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryOptions } from "@tanstack/svelte-query";

export function useDashboardNames(instanceId: string) {
  return useFilteredResourceNames(instanceId, ResourceKind.MetricsView);
}

export function useDashboardFileNames(instanceId: string) {
  return useMainEntityFiles(instanceId, "dashboards");
}

export function useDashboard(instanceId: string, metricViewName: string) {
  return useResource(instanceId, metricViewName, ResourceKind.MetricsView);
}

/**
 * Gets the valid metrics view spec. Only to be used in displaying a dashboard.
 * Use {@link useDashboard} in the metrics view editor and other use cases.
 */
export const useMetaQuery = <T = V1MetricsViewSpec>(
  instanceId: string,
  metricViewName: string,
  selector?: (meta: V1MetricsViewSpec) => T,
) => {
  return useResource<T>(
    instanceId,
    metricViewName,
    ResourceKind.MetricsView,
    (data) =>
      selector
        ? selector(data.metricsView?.state?.validSpec)
        : (data.metricsView?.state?.validSpec as T),
  );
};

// TODO: cleanup usage of useModelHasTimeSeries and useModelAllTimeRange
export const useModelHasTimeSeries = (
  instanceId: string,
  metricViewName: string,
) => useMetaQuery(instanceId, metricViewName, (meta) => !!meta?.timeDimension);

export function useModelAllTimeRange(
  instanceId: string,
  metricsViewName: string,
  options?: {
    query?: CreateQueryOptions;
  },
) {
  const { query: queryOptions } = options ?? {};

  return createQueryServiceMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    {},
    {
      query: {
        select: (data) => {
          if (!data.timeRangeSummary?.min || !data.timeRangeSummary?.max)
            return undefined;
          return {
            name: TimeRangePreset.ALL_TIME,
            start: new Date(data.timeRangeSummary.min),
            end: new Date(data.timeRangeSummary.max),
          };
        },
        ...queryOptions,
      },
    },
  );
}

export const useMetaMeasure = (
  instanceId: string,
  metricViewName: string,
  measureName: string,
) =>
  useMetaQuery(
    instanceId,
    metricViewName,
    (meta) => meta?.measures?.find((measure) => measure.name === measureName),
  );

export const useMetaDimension = (
  instanceId: string,
  metricViewName: string,
  dimensionName: string,
) =>
  useMetaQuery(instanceId, metricViewName, (meta) => {
    const dim = meta?.dimensions?.find(
      (dimension) => dimension.name === dimensionName,
    );
    return {
      ...dim,
      // this is for backwards compatibility when we used `name` as `column`
      column: dim.column ?? dim.name,
    };
  });

/**
 * Returns a copy of a V1MetricsViewFilter that does not include
 * the filters for the specified dimension name.
 */
export const getFiltersForOtherDimensions = (
  filters: V1MetricsViewFilter,
  dimensionName?: string,
) => {
  if (!filters) return { include: [], exclude: [] };

  const filter: V1MetricsViewFilter = {
    include:
      filters.include
        ?.filter((dimensionValues) => dimensionName !== dimensionValues.name)
        .map((dimensionValues) => ({
          name: dimensionValues.name,
          in: dimensionValues.in,
        })) ?? [],
    exclude:
      filters.exclude
        ?.filter((dimensionValues) => dimensionName !== dimensionValues.name)
        .map((dimensionValues) => ({
          name: dimensionValues.name,
          in: dimensionValues.in,
        })) ?? [],
  };
  return filter;
};

export const useGetDashboardsForModel = (
  instanceId: string,
  modelName: string,
) => {
  return useFilteredResources(instanceId, ResourceKind.MetricsView, (data) =>
    data.resources.filter((res) => res.metricsView?.spec?.table === modelName),
  );
};
