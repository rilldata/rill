import type { FieldType } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";
import { parseDocument } from "yaml";

export function getParsedDocument(fileArtifact: FileArtifact) {
  const { editorContent } = fileArtifact;
  return derived([editorContent], ([$localContent]) => {
    return parseDocument($localContent ?? "");
  });
}

export function useMetricFieldData(
  ctx: CanvasStore,
  metricViewName: string,
  type: FieldType[],
  searchableItems: string[] | undefined = undefined,
  searchValue = "",
  excludedValues: string[] | undefined = undefined,
) {
  const { metricsView, timeControls } = ctx.canvasEntity;

  const metricsViewQuery = metricsView.getMetricsViewFromName(metricViewName);

  return derived(
    [metricsViewQuery, timeControls.minTimeGrain],
    ([$metricsViewQuery, minTimeGrain]) => {
      const metricsViewSpec = $metricsViewQuery.metricsView;
      let items: string[] = [];
      const displayMap: Record<string, { label: string; type: FieldType }> = {};

      const measures = metricsViewSpec?.measures ?? [];
      const dimensions = metricsViewSpec?.dimensions ?? [];
      const timeDimension = metricsViewSpec?.timeDimension;

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

      const timeGrainOptions = Object.keys(TIME_GRAIN)
        .filter(
          (grain: V1TimeGrain) =>
            minTimeGrain === undefined ||
            minTimeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED ||
            !isGrainBigger(minTimeGrain, grain),
        )
        .map((grain) => {
          return {
            grain: grain,
            label: `Time ${TIME_GRAIN[grain].label}`,
            id: `${metricsViewSpec?.timeDimension}_rill_${grain}`,
          };
        });

      if (type.includes("time") && timeDimension) {
        items = items.concat(timeGrainOptions.map((tgo) => tgo.id));
        Object.assign(
          displayMap,
          Object.fromEntries(
            timeGrainOptions.map((tgo) => [
              tgo.id,
              { label: tgo.label, type: "time" },
            ]),
          ),
        );
      }

      const filteredItems = (
        searchableItems && searchValue ? searchableItems : items
      ).filter((item) => {
        // Exclude items that are in the excludedValues list
        if (excludedValues?.includes(item)) {
          return false;
        }

        const matches =
          displayMap[item]?.label
            ?.toLowerCase()
            .includes(searchValue.toLowerCase()) ||
          item.toLowerCase().includes(searchValue.toLowerCase());
        return matches;
      });

      return { items, displayMap, filteredItems };
    },
  );
}
