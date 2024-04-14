import { VisualizationSpec } from "svelte-vega";
import { ChartTypes } from "../types";
import { buildArea } from "./area";
import { buildGroupedBar } from "./grouped-bar";
import { buildLine } from "./line";
import { buildSimpleBar } from "./simple-bar";
import { buildStackedArea } from "./stacked-area";
import { buildStackedBar } from "./stacked-bar";

const BAR_LIKE_CHARTS = [
  ChartTypes.BAR,
  ChartTypes.GROUPED_BAR,
  ChartTypes.STACKED_BAR,
];
const LINE_LIKE_CHARTS = [
  ChartTypes.LINE,
  ChartTypes.AREA,
  ChartTypes.STACKED_AREA,
];
export function buildVegaLiteSpec(
  chartType: ChartTypes,
  timeFields: string[],
  quantitativeFields: string[],
  nominalFields: string[] = [],
): VisualizationSpec {
  if (!timeFields.length) throw "No time fields found";
  const hasNominalFields = nominalFields.length > 0;

  if (BAR_LIKE_CHARTS.includes(chartType)) {
    if (!hasNominalFields) {
      return buildSimpleBar(timeFields[0], quantitativeFields[0]);
    } else if (chartType === ChartTypes.GROUPED_BAR) {
      return buildGroupedBar(
        timeFields[0],
        quantitativeFields[0],
        nominalFields[0],
      );
    } else if (chartType === ChartTypes.STACKED_BAR) {
      return buildStackedBar(
        timeFields[0],
        quantitativeFields[0],
        nominalFields[0],
      );
    } else return buildSimpleBar(timeFields[0], quantitativeFields[0]);
  } else if (LINE_LIKE_CHARTS.includes(chartType)) {
    if (chartType == ChartTypes.AREA) {
      return buildArea(timeFields[0], quantitativeFields[0]);
    } else if (chartType == ChartTypes.LINE) {
      return buildLine(timeFields[0], quantitativeFields[0], nominalFields[0]);
    } else if (chartType == ChartTypes.STACKED_AREA) {
      return buildStackedArea(
        timeFields[0],
        quantitativeFields[0],
        nominalFields[0],
      );
    } else return buildArea(timeFields[0], quantitativeFields[0]);
  } else {
    throw new Error(`Chart type "${chartType}" not supported.`);
  }
}
