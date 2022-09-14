import type {
  MetricsViewMetaResponse,
  MetricsViewRequestFilter,
} from "$common/rill-developer-service/MetricsViewActions";
import { getMapFromArray } from "$common/utils/arrayUtils";

export const selectDimensionFromMeta = (
  meta: MetricsViewMetaResponse,
  dimensionId: string
) => meta?.dimensions?.find((dimension) => dimension.id === dimensionId);

export const selectMappedFilterFromMeta = (
  meta: MetricsViewMetaResponse,
  filters: MetricsViewRequestFilter
): MetricsViewRequestFilter => {
  const dimensionIdMap = getMapFromArray(
    meta.dimensions,
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
};

export const selectMeasureFromMeta = (
  meta: MetricsViewMetaResponse,
  measureId: string
) => meta?.measures?.find((measure) => measure.id === measureId);

export const selectMeasureNamesFromMeta = (
  meta: MetricsViewMetaResponse,
  measureIds: Array<string>
): Array<string> => {
  const measureIdMap = getMapFromArray(meta.measures, (measure) => measure.id);
  return measureIds.map((measureId) => measureIdMap.get(measureId).sqlName);
};
