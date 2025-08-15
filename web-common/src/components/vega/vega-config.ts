import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import {
  MainAreaColorGradientDark,
  MainAreaColorGradientLight,
  MainLineColor,
} from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { Config } from "vega-lite";

const BarFill = "var(--color-primary-400)";

const defaultMarkColor = MainLineColor;

// Light and dark mode color values for canvas compatibility
const colors = {
  light: {
    grid: "#e5e7eb", // gray-200
    axisLabel: "#6b7280", // gray-600
    surface: "white",
  },
  dark: {
    grid: "#374151", // gray-700
    axisLabel: "#d1d5db", // gray-300
    surface: "oklch(0.153 0.007 264.364)", // gray-50
  },
};

export const getRillTheme: (
  isCanvasDashboard: boolean,
  isDarkMode?: boolean,
) => Config = (isCanvasDashboard, isDarkMode = false) => {
  const gridColor = isDarkMode ? colors.dark.grid : colors.light.grid;
  const axisLabelColor = isDarkMode
    ? colors.dark.axisLabel
    : colors.light.axisLabel;

  return {
    autosize: {
      type: "fit-x",
    },
    background: isDarkMode ? colors.dark.surface : colors.light.surface,
    mark: {
      tooltip: isCanvasDashboard,
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
    bar: {
      fill: BarFill,
      ...(!isCanvasDashboard && { opacity: 0.8 }),
    },
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
      ...(!isCanvasDashboard && {
        gridDash: [2],
      }),
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
      ...(isCanvasDashboard && {
        grid: false,
      }),
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
      heatmap: {
        scheme: "tealblues", // TODO: Generate this from theme
      },
    },
    numberFormat: "s",
    tooltipFormat: {
      numberFormat: "d",
    },
    customFormatTypes: true,
  };
};
