import {
  MainAreaColorGradientDark,
  MainAreaColorGradientLight,
  MainLineColor,
} from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { Config } from "vega-lite";

const BarFill = "var(--color-primary-400)";
const CategoryColors = [
  "#75DAFF",
  "#5FA9B9",
  "#306B59",
  "#3125AE",
  "#757EFF",
  "#875FB9",
  "#F0A76A",
  "#948476",
  "#594159",
  "#5774A1",
  "#B7DAF0",
  "#E29FE3",
  "#FFCBDF",
  "#BFF7E3",
  "#FFFACB",
];

const defaultMarkColor = MainLineColor;
const gridColor = "#d1d5db"; // gray-300
const axisLabelColor = "#4b5563"; // gray-600

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
    orient: "right",
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
    labelFontSize: 11,
    labelFontWeight: 400,
    labelPadding: 12,
    labelOverlap: "parity",
    labelSeparation: 5,
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
