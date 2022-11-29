import type { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import type { DimensionDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type {
  MetricsViewMetaResponse,
  MetricsViewRequestFilter,
} from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
import { getMapFromArray } from "@rilldata/web-local/common/utils/arrayUtils";
import type { UseQueryOptions } from "@sveltestack/svelte-query/dist/types";
import {
  useRuntimeServiceGetCatalogEntry,
  V1GetCatalogEntryResponse,
  V1MetricsView,
} from "@rilldata/web-common/runtime-client";

export const MetaId = `v1/metrics-view/meta`;

export const getMetaQueryKey = (metricViewId: string) => {
  return [MetaId, metricViewId];
};

export const useMetaQuery = (instanceId: string, metricViewName: string) => {
  return useRuntimeServiceGetCatalogEntry(instanceId, metricViewName, {
    query: {
      enabled: !!metricViewName,
      select: (data) => data?.entry?.metricsView,
    },
  });
};

export const useCatalogQuery = <T = V1MetricsView>(
  instanceId: string,
  metricViewName: string,
  selector?: (meta: V1MetricsView) => T
) => {
  return useRuntimeServiceGetCatalogEntry(instanceId, metricViewName, {
    query: {
      enabled: !!metricViewName,
      ...(selector
        ? { select: (data) => selector(data?.entry?.metricsView) }
        : {}),
    },
  });
};

export const useMetaMeasure = (
  instanceId: string,
  metricViewName: string,
  measureName: string
) =>
  useCatalogQuery(instanceId, metricViewName, (meta) =>
    meta.measures?.find((measure) => measure.name === measureName)
  );

export const useMetaDimension = (
  instanceId: string,
  metricViewName: string,
  dimensionName: string
) =>
  useCatalogQuery(instanceId, metricViewName, (meta) =>
    meta.dimensions?.find((dimension) => dimension.name === dimensionName)
  );

export const useMetaMappedFilters = (
  instanceId: string,
  metricViewName: string,
  filters: MetricsViewRequestFilter,
  dimensionName?: string
) =>
  useCatalogQuery<MetricsViewRequestFilter>(instanceId, metricViewName, (_) => {
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
  });
