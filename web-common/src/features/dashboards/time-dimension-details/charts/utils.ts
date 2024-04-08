import { buildVegaLiteSpec } from "@rilldata/web-common/features/charts/templates/build-template";
import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { VisualizationSpec } from "svelte-vega";
import { TDDChart, TDDCustomCharts } from "../types";

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

export function getVegaSpec(
  chartType: TDDCustomCharts,
  expandedMeasureName: string,
  isDimensional: boolean,
): VisualizationSpec {
  const temporalFields = ["ts_position"];
  const measureFields = [expandedMeasureName];

  const spec = buildVegaLiteSpec(
    chartType,
    temporalFields,
    measureFields,
    isDimensional ? ["dimension"] : [],
  );

  return spec;
}

export function sanitizeSpecForTDD(
  spec: VisualizationSpec,
  timeGrain: V1TimeGrain,
  xMin: Date,
  xMax: Date,
  chartType: TDDCustomCharts,
): VisualizationSpec {
  if (!spec) return spec;

  const sanitizedSpec: VisualizationSpec = structuredClone(spec);
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

  const timeLabelFormat = TIME_GRAIN[timeGrain]?.d3format as string;
  // Remove titles from axes
  xEncoding.axis = {
    ticks: false,
    orient: "top",
    title: "",
    formatType: "time",
    format: timeLabelFormat,
  };
  yEncoding.axis = { title: "" };

  if (
    chartType === TDDChart.STACKED_BAR ||
    chartType === TDDChart.GROUPED_BAR
  ) {
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
