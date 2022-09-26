import type { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import type { DimensionDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type {
  MetricsViewMetaResponse,
  MetricsViewRequestFilter,
} from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
import { getMapFromArray } from "@rilldata/web-local/common/utils/arrayUtils";
import { fetchUrl } from "../fetch-url";
import { useQuery } from "@sveltestack/svelte-query";
import type { UseQueryOptions } from "@sveltestack/svelte-query/dist/types";

// GET /api/v1/metrics-views/{view-name}/meta

export const getMetricsViewMetadata = async (
  config: RootConfig,
  metricViewId: string
): Promise<MetricsViewMetaResponse> => {
  const json = await fetchUrl(
    config.server.exploreUrl,
    `metrics-views/${metricViewId}/meta`,
    "GET"
  );
  json.id = metricViewId;
  return json;
};

export const MetaId = `v1/metrics-view/meta`;

export const getMetaQueryKey = (metricViewId: string) => {
  return [MetaId, metricViewId];
};

export const useMetaQuery = <T = MetricsViewMetaResponse>(
  config: RootConfig,
  metricViewId: string,
  selector?: (meta: MetricsViewMetaResponse) => T
) => {
  const metaQueryKey = getMetaQueryKey(metricViewId);
  const metaQueryFn = () => getMetricsViewMetadata(config, metricViewId);
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

export const useMetaMeasure = (
  config: RootConfig,
  metricViewId: string,
  measureId: string
) =>
  useMetaQuery<MeasureDefinitionEntity>(config, metricViewId, (meta) =>
    meta.measures?.find((measure) => measure.id === measureId)
  );
export const useMetaMeasureNames = (
  config: RootConfig,
  metricViewId: string,
  measureIds: Array<string>
) =>
  useMetaQuery<Array<string>>(config, metricViewId, (meta) => {
    const measureIdMap = getMapFromArray(
      meta.measures ?? [],
      (measure) => measure.id
    );
    return (
      measureIds?.map((measureId) => measureIdMap.get(measureId).sqlName) ?? []
    );
  });

export const useMetaDimension = (
  config: RootConfig,
  metricViewId: string,
  dimensionId: string
) =>
  useMetaQuery<DimensionDefinitionEntity>(config, metricViewId, (meta) =>
    meta.dimensions?.find((dimension) => dimension.id === dimensionId)
  );

export const useMetaMappedFilters = (
  config: RootConfig,
  metricViewId: string,
  filters: MetricsViewRequestFilter
) =>
  useMetaQuery<MetricsViewRequestFilter>(config, metricViewId, (meta) => {
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
