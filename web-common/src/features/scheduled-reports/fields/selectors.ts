import type { FieldType } from "@rilldata/web-common/features/canvas/inspector/types.ts";
import { ephemeralMeasureToSpecMeasure } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping.ts";
import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types.ts";
import {
  getMeasureDisplayName,
  getDimensionDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import {
  getAvailableTimeGrains,
  translateV1TimeGrain,
} from "@rilldata/web-common/lib/time/new-grains";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived } from "svelte/store";

export function getFieldsForExplore(
  client: RuntimeClient,
  exploreName: string,
  ephemeralMeasures: EphemeralMeasureDef[] | undefined,
) {
  return derived(useExploreValidSpec(client, exploreName), ($validSpecResp) => {
    const metricsViewSpec = $validSpecResp.data?.metricsView ?? {};
    const exploreSpec = $validSpecResp.data?.explore ?? {};

    // Ephemeral measures have no spec entry; synthesize one so the report form
    // shows their display name and lets them be re-added as columns.
    const measures = [
      ...(metricsViewSpec.measures?.filter((measureSpec) =>
        exploreSpec.measures?.some(
          (exploreMeasure) => exploreMeasure === measureSpec.name,
        ),
      ) ?? []),
      ...(ephemeralMeasures ?? []).map(ephemeralMeasureToSpecMeasure),
    ];
    const dimensions =
      metricsViewSpec.dimensions?.filter((dimensionSpec) =>
        exploreSpec.dimensions?.some(
          (exploreDimension) => exploreDimension === dimensionSpec.name,
        ),
      ) ?? [];
    const allowedTimeGrains = getAvailableTimeGrains(metricsViewSpec);

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
        label: `Time ${translateV1TimeGrain(grain)}`,
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
  });
}
