import {
  useRuntimeServiceGetCatalogEntry,
  useRuntimeServiceGetTimeRangeSummary,
  useRuntimeServiceListCatalogEntries,
  useRuntimeServiceListFiles,
  V1MetricsView,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { TimeRangeName } from "./time-controls/time-control-types";

export function useDashboardNames() {
  return useRuntimeServiceListFiles(
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
  metricViewName: string,
  selector?: (meta: V1MetricsView) => T
) => {
  return useRuntimeServiceGetCatalogEntry(metricViewName, {
    query: {
      select: (data) =>
        selector
          ? selector(data?.entry?.metricsView)
          : data?.entry?.metricsView,
    },
  });
};

export const useModelHasTimeSeries = (metricViewName: string) =>
  useMetaQuery(metricViewName, (meta) => !!meta?.timeDimension);

export function useModelAllTimeRange(modelName: string, timeDimension: string) {
  return useRuntimeServiceGetTimeRangeSummary(
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
            name: TimeRangeName.AllTime,
            start: new Date(data.timeRangeSummary.min),
            end: new Date(data.timeRangeSummary.max),
          };
        },
      },
    }
  );
}

export const useMetaMeasure = (metricViewName: string, measureName: string) =>
  useMetaQuery(metricViewName, (meta) =>
    meta.measures?.find((measure) => measure.name === measureName)
  );

export const useMetaDimension = (
  metricViewName: string,
  dimensionName: string
) =>
  useMetaQuery(metricViewName, (meta) =>
    meta.dimensions?.find((dimension) => dimension.name === dimensionName)
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

export const useGetDashboardsForModel = (modelName: string) => {
  return useRuntimeServiceListCatalogEntries(
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
