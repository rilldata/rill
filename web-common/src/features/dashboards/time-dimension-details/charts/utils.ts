import {
  ChartField,
  buildVegaLiteSpec,
} from "@rilldata/web-common/features/charts/templates/build-template";
import { TDDChartMap } from "@rilldata/web-common/features/charts/types";
import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import {
  MainAreaColorGradientDark,
  MainLineColor,
} from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { VisualizationSpec } from "svelte-vega";
import { TDDAlternateCharts, TDDChart } from "../types";

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
  const temporalFields: ChartField[] = [{ name: "ts", label: "Time" }];
  const measureFields: ChartField[] = [
    { name: expandedMeasureName, label: measureLabel },
  ];

  let nominalFields: ChartField[] = [];
  if (isDimensional) {
    nominalFields = [
      {
        name: "dimension",
        label: dimensionName || "dimension",
        values: comparedValues,
      },
    ];
  } else if (isTimeComparison) {
    nominalFields = [
      {
        name: "key",
        label: "Comparing",
        values: [expandedMeasureName, `comparison.${expandedMeasureName}`],
      },
    ];

    // For time comparison, time field contains `ts` or `comparison.ts` based on data
    temporalFields[0].tooltipName = "time";
    // For time comparison, measure field contains `<measureName>` or
    // `comparison.<measureName>` based on data
    measureFields[0].name = "measure";
  }

  const builderChartType = TDDChartMap[chartType];

  const spec = buildVegaLiteSpec(
    builderChartType,
    temporalFields,
    measureFields,
    nominalFields,
  );

  return spec;
}

function patchSpecForTimeComparison(
  sanitizedSpec,
  chartType: TDDAlternateCharts,
  timeUnit: string,
  measureName: string,
  yEncoding,
  colorEncoding,
) {
  yEncoding.field = "measure";

  sanitizedSpec.transform = [
    // Sanitize and transform comparison data in the right time format
    { timeUnit: timeUnit, field: "comparison\\.ts", as: "comparison_ts" },
    // Expand datum to have a key field to differentiate between current and comparison data
    { fold: ["ts", "comparison_ts"] },
    // Add a measure field to hold the right measure value
    {
      calculate: `(datum['key'] === 'comparison_ts' ? datum['comparison.${measureName}'] : datum['${measureName}'])`,
      as: "measure",
    },
    // Add a time field to hold the right time value
    {
      calculate:
        "(datum['key'] === 'comparison_ts' ? datum['comparison_ts'] : datum['ts'])",
      as: "time",
    },
  ];

  colorEncoding.scale = {
    domain: ["ts", "comparison_ts"],
    range: [MainLineColor, MainAreaColorGradientDark],
  };

  if (chartType === TDDChart.STACKED_AREA) {
    /**
     * For stacked area charts, we don't need to pivot transform as the
     * comparison values are already in the right format.
     */

    // TODO: This is error prone, find a better way to do this without
    // mutating the template
    const stackedAreaPivotLayer = sanitizedSpec.layer[2];

    if (stackedAreaPivotLayer) {
      delete stackedAreaPivotLayer.transform;
    }
  }
}

export function patchSpecForTDD(
  spec,
  chartType: TDDAlternateCharts,
  timeGrain: V1TimeGrain,
  xMin: Date,
  xMax: Date,
  isTimeComparison: boolean,
  measureName: string,
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
  yEncoding.axis = { title: "", formatType: "measureFormatter" };

  // Set timeUnit for x-axis using timeGrain
  const timeUnit = timeGrainToVegaTimeUnitMap[timeGrain];
  xEncoding.timeUnit = timeUnit;

  if (isTimeComparison) {
    patchSpecForTimeComparison(
      sanitizedSpec,
      chartType,
      timeUnit,
      measureName,
      yEncoding,
      colorEncoding,
    );
  }

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
