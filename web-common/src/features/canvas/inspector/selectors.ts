import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { derived } from "svelte/store";
import { parseDocument } from "yaml";

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
  ctx: StateManagers,
  metricViewName: string,
  type: "measure" | "dimension",
  searchableItems: string[] | undefined,
  searchValue: string,
) {
  const { spec } = ctx.canvasEntity;
  const allDimensions = spec.getDimensionsForMetricView(metricViewName);
  const allMeasures = spec.getSimpleMeasuresForMetricView(metricViewName);

  return derived([allDimensions, allMeasures], ([dimensions, measures]) => {
    let items: string[] = [];
    let displayMap: Record<string, string> = {};

    if (type === "measure") {
      items = measures?.map((m) => m.name as string) ?? [];
      displayMap = Object.fromEntries(
        measures.map((item) => [
          item.name as string,
          getMeasureDisplayName(item),
        ]),
      );
    } else {
      items = dimensions?.map((d) => d.name || (d.column as string)) ?? [];
      displayMap = Object.fromEntries(
        dimensions.map((item) => [
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
  });
}
