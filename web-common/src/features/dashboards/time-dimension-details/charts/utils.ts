import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { TopLevelSpec } from "vega-lite";

export function sanitizeSpecForTDD(
  spec: TopLevelSpec,
  timeGrain: V1TimeGrain,
): TopLevelSpec {
  if (!spec) return spec;

  const sanitizedSpec = { ...spec };

  if (sanitizedSpec.encoding) {
    // Remove titles from axes
    sanitizedSpec.encoding.x.axis = { title: "" };
    sanitizedSpec.encoding.y.axis = { title: "" };

    // Set timeUnit for x-axis using timeGrain

    switch (timeGrain) {
      case V1TimeGrain.TIME_GRAIN_SECOND:
        sanitizedSpec.encoding.x.timeUnit = "yearmonthdatehoursminutesseconds";
        break;
      case V1TimeGrain.TIME_GRAIN_MINUTE:
        sanitizedSpec.encoding.x.timeUnit = "yearmonthdatehoursminutes";
        break;
      case V1TimeGrain.TIME_GRAIN_HOUR:
        sanitizedSpec.encoding.x.timeUnit = "yearmonthdatehours";
        break;
      case V1TimeGrain.TIME_GRAIN_DAY:
        sanitizedSpec.encoding.x.timeUnit = "yearmonthdate";
        break;
      case V1TimeGrain.TIME_GRAIN_WEEK:
        sanitizedSpec.encoding.x.timeUnit = "yearweek";
        break;
      case V1TimeGrain.TIME_GRAIN_MONTH:
        sanitizedSpec.encoding.x.timeUnit = "yearmonth";
        break;
      case V1TimeGrain.TIME_GRAIN_QUARTER:
        sanitizedSpec.encoding.x.timeUnit = "yearquarter";
        break;
      case V1TimeGrain.TIME_GRAIN_YEAR:
        sanitizedSpec.encoding.x.timeUnit = "year";
        break;
    }
  }

  return sanitizedSpec;
}
export function reduceDimensionData(dimensionData: DimensionDataItem[]) {
  return dimensionData
    .map((dimension) =>
      dimension.data.map((datum) => ({
        dimension: dimension.value,
        ...datum,
      })),
    )
    .flat();
}
