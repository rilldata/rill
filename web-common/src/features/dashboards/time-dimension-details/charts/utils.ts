import { TDDChartMap } from "@rilldata/web-common/components/vega/types";
import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import {
  type ChartField,
  buildVegaLiteSpec,
} from "@rilldata/web-common/features/templates/charts/build-template";
import type { View, VisualizationSpec } from "svelte-vega";
import { type TDDAlternateCharts, TDDChart } from "../types";

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

export function getVegaLiteSpecForTDD(
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
    {
      name: expandedMeasureName,
      label: measureLabel,
      formatterFunction: "measureFormatter",
    },
  ];

  let nominalFields: ChartField[] = [];
  if (
    isDimensional &&
    isTimeComparison &&
    chartType !== TDDChart.STACKED_AREA
  ) {
    nominalFields = [
      {
        name: "dimension",
        label: dimensionName || "dimension",
        values: comparedValues,
      },
    ];

    measureFields.push({
      name: `comparison.${expandedMeasureName}`,
      label: `comparison.${measureLabel}`,
      formatterFunction: "measureFormatter",
    });
  } else if (isDimensional) {
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

function isSignalEqual(currentValues: unknown[], newValues: unknown[]) {
  if (!Array.isArray(currentValues) || !Array.isArray(newValues)) {
    return false;
  }

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

export function updateChartOnTableCellHover(
  viewVL: View | undefined,
  chartType: TDDAlternateCharts,
  isTimeComparison: boolean,
  isDimensionComparison: boolean,
  time: Date | null | undefined,
  dimensionValue: string | null | undefined,
) {
  if (!viewVL) {
    return;
  }

  const hasComparison = isTimeComparison || isDimensionComparison;
  const epochTime = time ? time.getTime() : null;
  const hasDimensionParam = hasChartDimensionParam(chartType, hasComparison);

  if (hasDimensionParam && isTimeComparison && dimensionValue) {
    if (dimensionValue.startsWith("Previous")) dimensionValue = "comparison_ts";
    else dimensionValue = "ts";
  }
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

function checkLayerForBrush(layer) {
  if (layer.params && Array.isArray(layer.params)) {
    for (const param of layer.params) {
      if (param.name === "brush") {
        return true;
      }
    }
  }
  return false;
}

// Check if vega lite spec has brush params
// If so, we need to compile vega lite spec to vega spec
export function hasBrushParam(spec) {
  if (spec && spec.layer && Array.isArray(spec.layer)) {
    // Layered and Multi-view
    // https://vega.github.io/vega-lite/docs/spec.html#layered-and-multi-view-specifications
    for (const layer of spec.layer) {
      if (checkLayerForBrush(layer)) {
        return true;
      }
    }
  } else if (spec && spec.params && Array.isArray(spec.params)) {
    // Single view
    // https://vega.github.io/vega-lite/docs/spec.html#single
    if (checkLayerForBrush(spec)) {
      return true;
    }
  }

  return false;
}
