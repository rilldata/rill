import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import {
  MainAreaColorGradientDark,
  MainAreaColorGradientLight,
  MainLineColor,
} from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { Config } from "vega-lite";

const BarFill = "var(--color-primary-400)";

const defaultMarkColor = MainLineColor;
const gridColor = "#d1d5db"; // gray-300
const axisLabelColor = "#4b5563"; // gray-600

export const getRillTheme: (isCustomDashboard: boolean) => Config = (
  isCustomDashboard,
) => ({
  autosize: {
    type: "fit-x",
  },
  mark: {
    tooltip: isCustomDashboard ? true : false,
  },
  arc: { fill: defaultMarkColor },
  area: {
    line: { stroke: MainLineColor, strokeWidth: 1 },
    stroke: null,
    fillOpacity: 0.8,
    color: {
      x1: 1,
      y1: 1,
      x2: 1,
      y2: 0,
      gradient: "linear",
      stops: [
        {
          offset: 0,
          color: MainAreaColorGradientLight,
        },
        {
          offset: 1,
          color: MainAreaColorGradientDark,
        },
      ],
    },
  },
  bar: { fill: BarFill, opacity: 0.8 },
  line: { stroke: defaultMarkColor, strokeWidth: 1.5, strokeOpacity: 1 },
  path: { stroke: defaultMarkColor },
  rect: { fill: defaultMarkColor },
  shape: { stroke: defaultMarkColor },
  symbol: { fill: defaultMarkColor },

  legend: {
    orient: "top",
    labelFontSize: 11,
    labelColor: axisLabelColor,
    titleColor: axisLabelColor,
    labelFontWeight: 400,
    titleFontWeight: 500,
    titleFontSize: 12,
  },
  axisY: {
    orient: "left",
    gridColor: gridColor,
    gridDash: [2],
    tickColor: gridColor,
    domain: false,
    tickSize: 0,
    labelFont: "Inter, sans-serif",
    labelFontSize: 11,
    labelFontWeight: 400,
    labelColor: axisLabelColor,
    labelPadding: 10,
    titleColor: axisLabelColor,
    titleFont: "Inter, sans-serif",
    titleFontSize: 12,
    titleFontWeight: 500,
    titlePadding: 10,
    labelOverlap: false,
  },
  axisX: {
    gridColor: gridColor,
    gridDash: [2],
    tickColor: gridColor,
    tickSize: 0,
    domain: false,
    labelFont: "Inter, sans-serif",
    labelFontSize: 11,
    labelFontWeight: 400,
    labelPadding: 12,
    labelOverlap: "parity",
    labelSeparation: 5,
    labelColor: axisLabelColor,
    titleColor: axisLabelColor,
    titleFont: "Inter, sans-serif",
    titleFontSize: 12,
    titleFontWeight: 500,
    titlePadding: 10,
  },
  view: {
    strokeWidth: 0,
  },
  range: {
    category: COMPARIONS_COLORS,
  },
  numberFormat: "s",
  tooltipFormat: {
    numberFormat: "d",
  },
  customFormatTypes: true,
});
