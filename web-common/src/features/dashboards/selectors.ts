import {
  ResourceKind,
  useFilteredEntityNames,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewTimeRange,
  createRuntimeServiceGetCatalogEntry,
  createRuntimeServiceListCatalogEntries,
  V1MetricsView,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryOptions } from "@tanstack/svelte-query";

export function useDashboardNames(instanceId: string) {
  return useFilteredEntityNames(instanceId, ResourceKind.MetricsView);
}

export const useMetaQuery = <T = V1MetricsView>(
  instanceId: string,
  metricViewName: string,
  selector?: (meta: V1MetricsView) => T
) => {
  return createRuntimeServiceGetCatalogEntry(instanceId, metricViewName, {
    query: {
      select: (data) =>
        selector
          ? selector(data?.entry?.metricsView)
          : data?.entry?.metricsView,
    },
  });
};

export const useModelHasTimeSeries = (
  instanceId: string,
  metricViewName: string
) => useMetaQuery(instanceId, metricViewName, (meta) => !!meta?.timeDimension);

export function useModelAllTimeRange(
  instanceId: string,
  metricsViewName: string,
  options?: {
    query?: CreateQueryOptions;
  }
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
    }
  );
}

export const useMetaMeasure = (
  instanceId: string,
  metricViewName: string,
  measureName: string
) =>
  useMetaQuery(instanceId, metricViewName, (meta) =>
    meta?.measures?.find((measure) => measure.name === measureName)
  );

export const useMetaDimension = (
  instanceId: string,
  metricViewName: string,
  dimensionName: string
) =>
  useMetaQuery(instanceId, metricViewName, (meta) => {
    const dim = meta?.dimensions?.find(
      (dimension) => dimension.name === dimensionName
    );
    return {
      ...dim,
      // this is for backwards compatibility when we used `name` as `column`
      column: dim.column ?? dim.name,
    };
  });

/**
 * Returns a copy of the filter without the passed in dimension filters.
 */
export const getFilterForDimension = (
  filters: V1MetricsViewFilter,
  dimensionName?: string
) => {
  if (!filters) return undefined;
  return {
    include: filters.include
      .filter((dimensionValues) => dimensionName !== dimensionValues.name)
      .map((dimensionValues) => ({
        name: dimensionValues.name,
        in: dimensionValues.in,
      })),
    exclude: filters.exclude
      .filter((dimensionValues) => dimensionName !== dimensionValues.name)
      .map((dimensionValues) => ({
        name: dimensionValues.name,
        in: dimensionValues.in,
      })),
  };
};

export const useGetDashboardsForModel = (
  instanceId: string,
  modelName: string
) => {
  return createRuntimeServiceListCatalogEntries(
    instanceId,
    { type: "OBJECT_TYPE_METRICS_VIEW" },
    {
      query: {
        select(data) {
          return data?.entries?.filter(
            (entry) => entry?.metricsView?.model === modelName
          );
        },
      },
    }
  );
};
