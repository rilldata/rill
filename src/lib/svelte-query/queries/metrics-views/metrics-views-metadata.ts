import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type {
  MetricsViewMetaResponse,
  MetricsViewRequestFilter,
} from "$common/rill-developer-service/MetricsViewActions";
import { getMapFromArray } from "$common/utils/arrayUtils";
import { fetchUrl } from "$lib/svelte-query/queries/fetch-url";
import { useQuery } from "@sveltestack/svelte-query";
import type { UseQueryOptions } from "@sveltestack/svelte-query/dist/types";

// GET /api/v1/metrics-views/{view-name}/meta

export const getMetricsViewMetadata = async (
  metricViewId: string
): Promise<MetricsViewMetaResponse> => {
  const json = await fetchUrl(`metrics-views/${metricViewId}/meta`, "GET");
  json.id = metricViewId;
  return json;
};

export const MetaId = `v1/metrics-view/meta`;

export const getMetaQueryKey = (metricViewId: string) => {
  return [MetaId, metricViewId];
};

export const useMetaQuery = <T = MetricsViewMetaResponse>(
  metricViewId: string,
  selector?: (meta: MetricsViewMetaResponse) => T
) => {
  const metaQueryKey = getMetaQueryKey(metricViewId);
  const metaQueryFn = () => getMetricsViewMetadata(metricViewId);
  const metaQueryOptions: UseQueryOptions<MetricsViewMetaResponse, Error, T> = {
    enabled: !!metricViewId,
    ...(selector ? { select: selector } : {}),
  };
  return useQuery<MetricsViewMetaResponse, Error, T>(
    metaQueryKey,
    metaQueryFn,
    metaQueryOptions
  );
};

export const useMeasureFromMetaQuery = (
  metricViewId: string,
  measureId: string
) =>
  useMetaQuery<MeasureDefinitionEntity>(metricViewId, (meta) =>
    meta.measures?.find((measure) => measure.id === measureId)
  );
export const useMeasureNamesFromMetaQuery = (
  metricViewId: string,
  measureIds: Array<string>
) =>
  useMetaQuery<Array<string>>(metricViewId, (meta) => {
    const measureIdMap = getMapFromArray(
      meta.measures ?? [],
      (measure) => measure.id
    );
    return (
      measureIds?.map((measureId) => measureIdMap.get(measureId).sqlName) ?? []
    );
  });

export const useDimensionFromMetaQuery = (
  metricViewId: string,
  dimensionId: string
) =>
  useMetaQuery<DimensionDefinitionEntity>(metricViewId, (meta) =>
    meta.dimensions?.find((dimension) => dimension.id === dimensionId)
  );

export const useMappedFiltersFromMetaQuery = (
  metricViewId: string,
  filters: MetricsViewRequestFilter
) =>
  useMetaQuery<MetricsViewRequestFilter>(metricViewId, (meta) => {
    if (!filters) return undefined;
    const dimensionIdMap = getMapFromArray(
      meta.dimensions ?? [],
      (dimension) => dimension.id
    );
    return {
      include: filters.include.map((dimensionValues) => ({
        name: dimensionIdMap.get(dimensionValues.name).dimensionColumn,
        values: dimensionValues.values,
      })),
      exclude: filters.exclude.map((dimensionValues) => ({
        name: dimensionIdMap.get(dimensionValues.name).dimensionColumn,
        values: dimensionValues.values,
      })),
    };
  });
