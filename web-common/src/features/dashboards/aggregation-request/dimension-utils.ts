import type { PivotChipData } from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import {
  type V1MetricsViewAggregationDimension,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

export function getDimensionForTimeField(
  timeDimension: string,
  selectedTimezone: string,
  pivotChipData: PivotChipData,
  humanReadableAlias: boolean,
) {
  return <V1MetricsViewAggregationDimension>{
    name: timeDimension,
    timeGrain: pivotChipData.id as V1TimeGrain,
    timeZone: selectedTimezone,
    alias: humanReadableAlias
      ? `Time ${pivotChipData.title}`
      : `${timeDimension}_rill_${pivotChipData.id}`,
  };
}

export function getDimensionNameFromAggregationDimension(
  dimension: V1MetricsViewAggregationDimension,
) {
  if (!dimension.timeGrain) return dimension.name!;
  return `${dimension.name}_rill_${dimension.timeGrain}`;
}

const timeDimensionNameRegex = /^(.*)_rill_(.*)$/;
export function getTimestampAndGrainFromTimeDimension(colName: string) {
  const match = timeDimensionNameRegex.exec(colName);
  if (!match) {
    return {
      timeCol: colName,
      grain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
    };
  }

  const [, timeCol, grain] = match;
  return {
    timeCol,
    grain,
  };
}
