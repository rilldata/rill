import {
  ChartField,
  buildVegaLiteSpec,
} from "@rilldata/web-common/features/charts/templates/build-template";
import { TDDChartMap } from "@rilldata/web-common/features/charts/types";
import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import { View, VisualizationSpec } from "svelte-vega";
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

function isSignalEqual(currentValues, newValues) {
  if (currentValues.length !== newValues.length) {
    return false;
  }

  for (let i = 0; i < currentValues.length; i++) {
    if (currentValues[i] !== newValues[i]) {
      return false;
    }
  }
  return true;
}

function hasChartDimensionParam(
  chartType: TDDAlternateCharts,
  hasComparison: boolean,
) {
  if (!hasComparison) {
    return false;
  } else if (chartType === TDDChart.STACKED_AREA) {
    return false;
  }

  return true;
}

export function updateVegaOnTableHover(
  viewVL: View | undefined,
  chartType: TDDAlternateCharts,
  hasComparison: boolean,
  time: Date | null | undefined,
  dimensionValue: string | null | undefined,
) {
  if (!viewVL) {
    return;
  }

  const epochTime = time ? time.getTime() : null;
  const hasDimensionParam = hasChartDimensionParam(chartType, hasComparison);
  const values = epochTime
    ? hasDimensionParam
      ? [epochTime, dimensionValue]
      : [epochTime]
    : null;

  const newValue = epochTime
    ? {
        unit: "",
        fields: viewVL.signal("hover_tuple_fields"),
        values,
      }
    : null;
  const currentValues = (viewVL.signal("hover_tuple") || { values: [] }).values;
  const newValues = values || [];
  if (isSignalEqual(currentValues, newValues)) {
    return;
  }
  viewVL.signal("hover_tuple", newValue);
  viewVL.runAsync();
}
