import {
  MainAreaColorGradientDark,
  MainAreaColorGradientLight,
  MainLineColor,
} from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { Config } from "vega-lite";

const BarFill = "var(--color-primary-400)";
const CategoryColors = [
  "#5FA9B9",
  "#875FB9",
  "#306B59",
  "#594159",
  "#E29FE3",
  "#B7DAF0",
  "#FFCBDF",
];

const defaultMarkColor = MainLineColor;
const gridColor = "#d1d5db"; // gray-300
const axisLabelColor = "#374151"; // gray-700

export const getRillTheme: () => Config = () => ({
  mark: {
    tooltip: true,
  },
  arc: { fill: defaultMarkColor },
  area: {
    line: { stroke: MainLineColor, strokeWidth: 1 },
    stroke: null,
    fillOpacity: 0.7,
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
  bar: { fill: BarFill, opacity: 0.7 },
  line: { stroke: defaultMarkColor, strokeWidth: 1.5, strokeOpacity: 1 },
  path: { stroke: defaultMarkColor },
  rect: { fill: defaultMarkColor },
  shape: { stroke: defaultMarkColor },
  symbol: { fill: defaultMarkColor },

  axisY: {
    gridColor: gridColor,
    gridDash: [2],
    tickColor: gridColor,
    domain: false,
    labelFont: "Inter, sans-serif",
    labelFontSize: 10,
    labelFontWeight: 500,
    labelColor: axisLabelColor,
    labelPadding: 5,
    titleColor: axisLabelColor,
    titleFont: "Inter, sans-serif",
    titleFontSize: 12,
    titleFontWeight: "bold",
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
    labelFontSize: 10,
    labelFontWeight: 500,
    labelPadding: 5,
    labelColor: axisLabelColor,
    titleColor: axisLabelColor,
    titleFont: "Inter, sans-serif",
    titleFontSize: 12,
    titleFontWeight: "bold",
    titlePadding: 10,
  },
  view: {
    strokeWidth: 0,
  },
  range: {
    category: CategoryColors,
  },
});
