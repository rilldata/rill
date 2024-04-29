import { buildVegaLiteSpec } from "@rilldata/web-common/features/charts/templates/build-template";
import { TDDChartMap } from "@rilldata/web-common/features/charts/types";
import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { VisualizationSpec } from "svelte-vega";
import { TDDAlternateCharts } from "../types";

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

export function getVegaSpecForTDD(
  chartType: TDDAlternateCharts,
  expandedMeasureName: string,
  measureLabel: string,
  isTimeComparison: boolean,
  isDimensional: boolean,
  dimensionName: string | undefined,
  comparedValues: (string | null)[] | undefined,
): VisualizationSpec {
  const temporalFields = [{ name: "ts", label: "Time" }];
  const measureFields = [{ name: expandedMeasureName, label: measureLabel }];

  if (isTimeComparison) {
    measureFields.push({ name: "comparison.ts", label: "Compared Time" });
  }
  const nominalFields = [
    {
      name: "dimension",
      label: dimensionName || "dimension",
      values: comparedValues,
    },
  ];

  const builderChartType = TDDChartMap[chartType];

  const spec = buildVegaLiteSpec(
    builderChartType,
    temporalFields,
    measureFields,
    isDimensional ? nominalFields : [],
  );

  return spec;
}

export function sanitizeSpecForTDD(
  spec,
  timeGrain: V1TimeGrain,
  xMin: Date,
  xMax: Date,
  selectedDimensionValues: (string | null)[] = [],
): VisualizationSpec {
  if (!spec) return spec;

  /**
   * Sub level types are not being exported from the vega-lite package.
   * This makes it hard to modify the specs without breaking typescript
   * interface. For now we have removed the types for the spec and will
   * add them back when we have the support for it.
   * More at https://github.com/vega/vega-lite/issues/9222
   */

  const sanitizedSpec = structuredClone(spec);
  let xEncoding;
  let yEncoding;
  let colorEncoding;
  if (sanitizedSpec.encoding) {
    xEncoding = sanitizedSpec.encoding.x;
    yEncoding = sanitizedSpec.encoding.y;

    colorEncoding = sanitizedSpec.encoding?.color;

    xEncoding.scale = {
      domain: [xMin.toISOString(), xMax.toISOString()],
    };
  }

  if (!xEncoding || !yEncoding) {
    return sanitizedSpec;
  }

  const selectedValuesLength = selectedDimensionValues.length;
  if (colorEncoding && selectedValuesLength) {
    colorEncoding.scale = {
      domain: selectedDimensionValues,
      range: COMPARIONS_COLORS.slice(0, selectedValuesLength),
    };
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

  // Set timeUnit for x-axis using timeGrain
  const timeUnit = timeGrainToVegaTimeUnitMap[timeGrain];
  xEncoding.timeUnit = timeUnit;

  return sanitizedSpec;
}

const timeGrainToVegaTimeUnitMap = {
  [V1TimeGrain.TIME_GRAIN_SECOND]: "yearmonthdatehoursminutesseconds",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "yearmonthdatehoursminutes",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "yearmonthdatehours",
  [V1TimeGrain.TIME_GRAIN_DAY]: "yearmonthdate",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "yearweek",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "yearmonth",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "yearquarter",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "year",
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "yearmonthdate",
};
