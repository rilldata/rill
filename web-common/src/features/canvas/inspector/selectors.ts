import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import { useMetricsViewValidSpec } from "@rilldata/web-common/features/dashboards/selectors";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { MetricsViewSpecMeasureType } from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";
import { parseDocument } from "yaml";

export const useAllDimensionFromMetric = (
  instanceId: string,
  metricsViewName: string,
) =>
  useMetricsViewValidSpec(
    instanceId,
    metricsViewName,
    (meta) => meta?.dimensions,
  );

export const useAllSimpleMeasureFromMetric = (
  instanceId: string,
  metricsViewName: string,
) =>
  useMetricsViewValidSpec(instanceId, metricsViewName, (meta) =>
    meta?.measures?.filter(
      (m) =>
        !m.window &&
        m.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON,
    ),
  );

export function getParsedDocument(fileArtifact: FileArtifact) {
  const { editorContent, remoteContent } = fileArtifact;
  return derived(
    [editorContent, remoteContent],
    ([$localContent, $remoteContent]) => {
      return parseDocument($localContent ?? $remoteContent ?? "");
    },
  );
}

export function useMetricFieldData(
  instanceId: string,
  metricName: string,
  type: "measure" | "dimension",
  searchableItems: string[] | undefined,
  searchValue: string,
) {
  const allDimensions = useAllDimensionFromMetric(instanceId, metricName);
  const allFilteredMeasures = useAllSimpleMeasureFromMetric(
    instanceId,
    metricName,
  );

  return derived(
    [allDimensions, allFilteredMeasures],
    ([$allDimensions, $allFilteredMeasures]) => {
      let items: string[] = [];
      let displayMap: Record<string, string> = {};

      if (type === "measure") {
        const itemsData = $allFilteredMeasures?.data ?? [];
        items = itemsData?.map((m) => m.name as string) ?? [];
        displayMap = Object.fromEntries(
          itemsData.map((item) => [
            item.name as string,
            getMeasureDisplayName(item),
          ]),
        );
      } else {
        const itemsData = $allDimensions?.data ?? [];
        items = itemsData?.map((d) => d.name || (d.column as string)) ?? [];
        displayMap = Object.fromEntries(
          itemsData.map((item) => [
            item.name || (item.column as string),
            getDimensionDisplayName(item),
          ]),
        );
      }

      const filteredItems = (
        searchableItems && searchValue ? searchableItems : items
      ).filter((item) => {
        const matches =
          displayMap[item]?.toLowerCase().includes(searchValue.toLowerCase()) ||
          item.toLowerCase().includes(searchValue.toLowerCase());
        return matches;
      });

      return { items, displayMap, filteredItems };
    },
  );
}
