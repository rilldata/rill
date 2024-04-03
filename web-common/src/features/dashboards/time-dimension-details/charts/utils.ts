import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { VisualizationSpec } from "svelte-vega";

export function sanitizeSpecForTDD(
  spec: VisualizationSpec,
  timeGrain: V1TimeGrain,
  xMin: Date,
  xMax: Date,
  chartType: string,
): VisualizationSpec {
  if (!spec) return spec;

  const sanitizedSpec: VisualizationSpec = { ...spec };
  let xEncoding;
  let yEncoding;
  if (sanitizedSpec.encoding) {
    xEncoding = sanitizedSpec?.encoding?.x;
    yEncoding = sanitizedSpec?.encoding?.y;

    xEncoding.scale = {
      domain: [xMin.toISOString(), xMax.toISOString()],
    };
  } else if (sanitizedSpec?.layer?.[0].encoding) {
    xEncoding = sanitizedSpec?.layer?.[0].encoding?.x;
    yEncoding = sanitizedSpec?.layer?.[0].encoding?.y;
  }

  // Set extents for x-axis
  xEncoding.scale = {
    domain: [xMin.toISOString(), xMax.toISOString()],
  };

  // Remove titles from axes
  xEncoding.axis = { ticks: false, title: "" };
  yEncoding.axis = { title: "" };

  if (chartType === "bar" || chartType === "stacked bar") {
    // Set timeUnit for x-axis using timeGrain
    switch (timeGrain) {
      case V1TimeGrain.TIME_GRAIN_SECOND:
        xEncoding.timeUnit = "yearmonthdatehoursminutesseconds";
        break;
      case V1TimeGrain.TIME_GRAIN_MINUTE:
        xEncoding.timeUnit = "yearmonthdatehoursminutes";
        break;
      case V1TimeGrain.TIME_GRAIN_HOUR:
        xEncoding.timeUnit = "yearmonthdatehours";
        break;
      case V1TimeGrain.TIME_GRAIN_DAY:
        xEncoding.timeUnit = "yearmonthdate";
        break;
      case V1TimeGrain.TIME_GRAIN_WEEK:
        xEncoding.timeUnit = "yearweek";
        break;
      case V1TimeGrain.TIME_GRAIN_MONTH:
        xEncoding.timeUnit = "yearmonth";
        break;
      case V1TimeGrain.TIME_GRAIN_QUARTER:
        xEncoding.timeUnit = "yearquarter";
        break;
      case V1TimeGrain.TIME_GRAIN_YEAR:
        xEncoding.timeUnit = "year";
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
