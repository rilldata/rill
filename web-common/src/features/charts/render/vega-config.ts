import {
  MainAreaColorGradientDark,
  MainAreaColorGradientLight,
  MainLineColor,
} from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { Config } from "vega-lite";

const markColor = MainLineColor;
const axisColor = "#E5E7EB";
const axisLabelColor = "#4b5563"; // gray-600

export const getRillTheme: () => Config = () => ({
  arc: { fill: markColor },
  area: {
    line: { stroke: markColor },
    stroke: null,
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
  bar: { fill: markColor },
  line: { stroke: markColor },
  path: { stroke: markColor },
  rect: { fill: markColor },
  shape: { stroke: markColor },
  symbol: { fill: markColor },

  axisY: {
    gridColor: axisColor,
    gridDash: [2],
    tickColor: axisColor,
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
    gridColor: axisColor,
    gridDash: [2],
    tickColor: axisColor,
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
});
