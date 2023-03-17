import {
  useRuntimeServiceGetCatalogEntry,
  useQueryServiceColumnTimeRange,
  useRuntimeServiceListCatalogEntries,
  useRuntimeServiceListFiles,
  V1MetricsView,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { TimeRangePreset } from "./time-controls/utils/time-types";

export function useDashboardNames(repoId: string) {
  return useRuntimeServiceListFiles(
    repoId,
    {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        // refetchInterval: 1000,
        select: (data) =>
          data.paths
            ?.filter((path) => path.includes("dashboards/"))
            .map((path) =>
              path.replace("/dashboards/", "").replace(".yaml", "")
            )
            // sort alphabetically case-insensitive
            .sort((a, b) =>
              a.localeCompare(b, undefined, { sensitivity: "base" })
            ),
      },
    }
  );
}

export const useMetaQuery = <T = V1MetricsView>(
  instanceId: string,
  metricViewName: string,
  selector?: (meta: V1MetricsView) => T
) => {
  return useRuntimeServiceGetCatalogEntry(instanceId, metricViewName, {
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
  modelName: string,
  timeDimension: string
) {
  return useQueryServiceColumnTimeRange(
    instanceId,
    modelName,
    {
      columnName: timeDimension,
    },
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
  useMetaQuery(instanceId, metricViewName, (meta) =>
    meta?.dimensions?.find((dimension) => dimension.name === dimensionName)
  );

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
  return useRuntimeServiceListCatalogEntries(
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
