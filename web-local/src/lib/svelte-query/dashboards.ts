import {
  useRuntimeServiceGetCatalogEntry,
  useRuntimeServiceListFiles,
  V1MetricsView,
  V1MetricsViewRequestFilter,
} from "@rilldata/web-common/runtime-client";

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

export const useMetaMeasure = (
  instanceId: string,
  metricViewName: string,
  measureName: string
) =>
  useMetaQuery(instanceId, metricViewName, (meta) =>
    meta.measures?.find((measure) => measure.name === measureName)
  );

export const useMetaDimension = (
  instanceId: string,
  metricViewName: string,
  dimensionName: string
) =>
  useMetaQuery(instanceId, metricViewName, (meta) =>
    meta.dimensions?.find((dimension) => dimension.name === dimensionName)
  );

/**
 * Returns a copy of the filter without the passed in dimension filters.
 */
export const getFilterForDimension = (
  filters: V1MetricsViewRequestFilter,
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
