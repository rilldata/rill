import type { FieldType } from "@rilldata/web-common/features/canvas/inspector/types.ts";
import {
  getMeasureDisplayName,
  getDimensionDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";

export function getFieldsForExplore(instanceId: string, exploreName: string) {
  return derived(
    useExploreValidSpec(instanceId, exploreName),
    ($validSpecResp) => {
      const metricsViewSpec = $validSpecResp.data?.metricsView ?? {};
      const minTimeGrain = metricsViewSpec.smallestTimeGrain;
      const exploreSpec = $validSpecResp.data?.explore ?? {};

      const measures =
        metricsViewSpec.measures?.filter((measureSpec) =>
          exploreSpec.measures?.some(
            (exploreMeasure) => exploreMeasure === measureSpec.name,
          ),
        ) ?? [];
      const dimensions =
        metricsViewSpec.dimensions?.filter((dimensionSpec) =>
          exploreSpec.dimensions?.some(
            (exploreDimension) => exploreDimension === dimensionSpec.name,
          ),
        ) ?? [];
      const allowedTimeGrains = Object.keys(TIME_GRAIN).filter(
        (grain: V1TimeGrain) =>
          minTimeGrain === undefined ||
          minTimeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED ||
          !isGrainBigger(minTimeGrain, grain),
      );

      const displayMap: Record<string, { label: string; type: FieldType }> = {};
      const allowedRows: string[] = [];
      const allowedColumns: string[] = [];

      measures.forEach((measure) => {
        displayMap[measure.name!] = {
          label: getMeasureDisplayName(measure),
          type: "measure",
        };
        allowedColumns.push(measure.name!);
      });

      dimensions.forEach((dimension) => {
        displayMap[dimension.name!] = {
          label: getDimensionDisplayName(dimension),
          type: "dimension",
        };
        allowedRows.push(dimension.name!);
        allowedColumns.push(dimension.name!);
      });

      allowedTimeGrains.forEach((grain) => {
        const id = `${metricsViewSpec.timeDimension}_rill_${grain}`;
        displayMap[id] = {
          label: `Time ${V1TimeGrainToDateTimeUnit[grain]}`,
          type: "time",
        };
        allowedRows.push(id);
        allowedColumns.push(id);
      });

      return {
        displayMap,
        allowedRows,
        allowedColumns,
      };
    },
  );
}
