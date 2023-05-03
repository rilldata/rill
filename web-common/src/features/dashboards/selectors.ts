import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  RpcStatus,
  V1MetricsView,
  V1MetricsViewFilter,
  createQueryServiceColumnTimeRange,
  createRuntimeServiceGetCatalogEntry,
  createRuntimeServiceListCatalogEntries,
  createRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type {
  CreateQueryOptions,
  QueryObserverResult,
} from "@tanstack/svelte-query";

export function useDashboardNames(instanceId: string) {
  return createRuntimeServiceListFiles(
    instanceId,
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
  return createRuntimeServiceGetCatalogEntry(instanceId, metricViewName, {
    query: {
      select: (data) =>
        selector
          ? selector(data?.entry?.metricsView)
          : data?.entry?.metricsView,
    },
  });
};

/**
 * This selector returns the best available string for each measure,
 * using the "label" if available but falling back to the expression
 * if needed.
 *
 * @param metaQuery: QueryObserverResult<V1MetricsView, RpcStatus>
 * @returns string[]
 */
export const selectBestMeasureStrings = (
  metaQuery: QueryObserverResult<V1MetricsView, RpcStatus>
): string[] => {
  if (metaQuery && metaQuery.isSuccess && !metaQuery.isRefetching) {
    return metaQuery.data?.measures?.map((m) => m.label || m.expression) ?? [];
  }
  return [];
};

/**
 * This selector returns the measure key, which can be used to
 * lookup measures across sessions, for example in stateful URLs
 * 
 * FIXME:
 * For now we are using the user supplied `expression` for measure
 * keys because that is the only field that must exist for the
 *  measure to appear in the dashboard.
 * This may lead to problems if there are ever duplicate
 * expressions so Hamilton has started discussions with Benjamin about
 * adding unique IDS that could be used to replace these temporary keys. 
 * Once those become available the fields below should be updated.

 * @param metaQuery: QueryObserverResult<V1MetricsView, RpcStatus>
 * @returns string[]
 */
export const selectMeasureKeys = (
  metaQuery: QueryObserverResult<V1MetricsView, RpcStatus>
): string[] => {
  if (metaQuery && metaQuery.isSuccess && !metaQuery.isRefetching) {
    return metaQuery.data?.measures?.map((m) => m.expression) ?? [];
  }
  return [];
};

/**
 * This selector returns the best available string for each dimension,
 * using the "label" if available but falling back to the name of
 * the categorical column (which must be present) if needed
 * @param metaQuery: QueryObserverResult<V1MetricsView, RpcStatus>
 * @returns string[]
 */
export const selectBestDimensionStrings = (
  metaQuery: QueryObserverResult<V1MetricsView, RpcStatus>
): string[] => {
  if (metaQuery && metaQuery.isSuccess && !metaQuery.isRefetching)
    return metaQuery.data?.dimensions?.map((d) => d.label || d.name) ?? [];
};

/**
 * This selector returns the dimension key, which can be used to
 * lookup dimensions across sessions, for example in stateful URLs
 * 
 * FIXME:
 * For now we are using the user supplied `name` for dimension
 * keys because that is the only field that must exist for the
 * dimension to appear in the dashboard

 * This may lead to problems if there are ever duplicates `names`,
 * so Hamilton has started discussions with Benjamin about
 * adding unique IDS that could be used to replace these temporary keys. 
 * Once those become available the fields below should be updated.

 * @param metaQuery: QueryObserverResult<V1MetricsView, RpcStatus>
 * @returns string[]
 */
export const selectDimensionKeys = (
  metaQuery: QueryObserverResult<V1MetricsView, RpcStatus>
): string[] => {
  if (metaQuery && metaQuery.isSuccess && !metaQuery.isRefetching) {
    return metaQuery.data?.dimensions?.map((d) => d.name) ?? [];
  }
  return [];
};

export const useModelHasTimeSeries = (
  instanceId: string,
  metricViewName: string
) => useMetaQuery(instanceId, metricViewName, (meta) => !!meta?.timeDimension);

export function useModelAllTimeRange(
  instanceId: string,
  modelName: string,
  timeDimension: string,
  options?: {
    query?: CreateQueryOptions;
  }
) {
  const { query: queryOptions } = options ?? {};

  return createQueryServiceColumnTimeRange(
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
