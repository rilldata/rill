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
  const { localContent, remoteContent } = fileArtifact;
  return derived(
    [localContent, remoteContent],
    ([$localContent, $remoteContent]) => {
      return parseDocument($localContent ?? $remoteContent ?? "");
    },
  );
}
