import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { derived } from "svelte/store";
import { parseDocument } from "yaml";

export type FieldType = "measure" | "dimension" | "time";

export function getParsedDocument(fileArtifact: FileArtifact) {
  const { editorContent } = fileArtifact;
  return derived([editorContent], ([$localContent]) => {
    return parseDocument($localContent ?? "");
  });
}

export function useMetricFieldData(
  ctx: StateManagers,
  metricViewName: string,
  type: FieldType[],
  searchableItems: string[] | undefined = undefined,
  searchValue = "",
) {
  const { spec } = ctx.canvasEntity;

  const metricViewSpec = spec.getMetricsViewFromName(metricViewName);

  return derived([metricViewSpec], ([metricViewSpec]) => {
    let items: string[] = [];
    const displayMap: Record<string, { label: string; type: FieldType }> = {};

    const measures = metricViewSpec?.measures ?? [];
    const dimensions = metricViewSpec?.dimensions ?? [];
    const timeDimension = metricViewSpec?.timeDimension;

    if (type.includes("measure")) {
      items = measures.map((m) => m.name as string);
      Object.assign(
        displayMap,
        Object.fromEntries(
          measures.map((item) => [
            item.name as string,
            { label: getMeasureDisplayName(item), type: "measure" },
          ]),
        ),
      );
    }
    if (type.includes("dimension")) {
      items = items.concat(
        dimensions?.map((d) => d.name || (d.column as string)) ?? [],
      );
      Object.assign(
        displayMap,
        Object.fromEntries(
          dimensions.map((item) => [
            item.name || (item.column as string),
            { label: getDimensionDisplayName(item), type: "dimension" },
          ]),
        ),
      );
    }
    if (type.includes("time") && timeDimension) {
      items.push(timeDimension);
      Object.assign(displayMap, {
        [timeDimension]: { label: "Time", type: "time" },
      });
    }
    const filteredItems = (
      searchableItems && searchValue ? searchableItems : items
    ).filter((item) => {
      const matches =
        displayMap[item]?.label
          ?.toLowerCase()
          .includes(searchValue.toLowerCase()) ||
        item.toLowerCase().includes(searchValue.toLowerCase());
      return matches;
    });

    return { items, displayMap, filteredItems };
  });
}
