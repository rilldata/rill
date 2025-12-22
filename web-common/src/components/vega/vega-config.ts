import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import { getSequentialColorsAsHex } from "@rilldata/web-common/features/themes/palette-store";
import { themeManager } from "@rilldata/web-common/features/themes/theme-manager";
import { getChroma } from "@rilldata/web-common/features/themes/theme-utils";
import type { Config } from "vega-lite";

function resolveCSSVariable(
  cssVar: string,
  isDarkMode: boolean,
  fallback?: string,
): string {
  return themeManager.resolveCSSVariable(cssVar, isDarkMode, fallback);
}

// Light and dark mode color values for canvas compatibility
const colors = {
  light: {
    grid: "#e5e7eb", // gray-200
    axisLabel: "#6b7280", // gray-600
  },
  dark: {
    grid: "#374151", // gray-700
    axisLabel: "#d1d5db", // gray-300
  },
};

export const getRillTheme: (
  isCanvasDashboard: boolean,
  isDarkMode?: boolean,
  theme?: Record<string, string>,
) => Config = (isCanvasDashboard, isDarkMode = false, theme) => {
  const gridColor = isDarkMode ? colors.dark.grid : colors.light.grid;
  const axisLabelColor = isDarkMode
    ? colors.dark.axisLabel
    : colors.light.axisLabel;

  // Use provided theme if available, otherwise resolve from CSS variables
  let lineColor: string,
    barColor: string,
    areaGradientLight: string,
    areaGradientDark: string;

  if (theme?.primary) {
    // Use theme's primary color directly
    const primary = getChroma(theme.primary);
    lineColor = primary.darken(0.2).css();
    barColor = primary.css();
    areaGradientLight = primary.brighten(2).css();
    areaGradientDark = primary.darken(0.5).css();
  } else {
    // Fallback: resolve from CSS variables (for standalone charts without theme context)
    lineColor = resolveCSSVariable(
      "var(--color-theme-600)",
      isDarkMode,
      "var(--color-primary-600)",
    );
    barColor = resolveCSSVariable(
      "var(--color-theme-400)",
      isDarkMode,
      "var(--color-primary-400)",
    );
    areaGradientLight = resolveCSSVariable(
      "var(--color-theme-50)",
      isDarkMode,
      "var(--color-primary-50)",
    );
    areaGradientDark = resolveCSSVariable(
      "var(--color-theme-300)",
      isDarkMode,
      "var(--color-primary-300)",
    );
  }

  return {
    autosize: {
      type: "fit-x",
    },
    background: "transparent",
    mark: {
      tooltip: isCanvasDashboard,
    },
    arc: { fill: lineColor },
    area: {
      line: { stroke: lineColor, strokeWidth: 1 },
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
            color: areaGradientLight,
          },
          {
            offset: 1,
            color: areaGradientDark,
          },
        ],
      },
    },
    bar: {
      fill: barColor,
      ...(!isCanvasDashboard && { opacity: 0.8 }),
    },
    line: { stroke: lineColor, strokeWidth: 1.5, strokeOpacity: 1 },
    path: { stroke: lineColor },
    rect: { fill: lineColor },
    shape: { stroke: lineColor },
    symbol: { fill: lineColor },

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
      category: (() => {
        const defaultColors = COMPARIONS_COLORS.map((color) =>
          color.startsWith("var(")
            ? resolveCSSVariable(color, isDarkMode)
            : color,
        );

        if (!theme) return defaultColors;

        const themeColors = [
          theme["color-qualitative-1"],
          theme["color-qualitative-2"],
          theme["color-qualitative-3"],
          theme["color-qualitative-4"],
          theme["color-qualitative-5"],
          theme["color-qualitative-6"],
          theme["color-qualitative-7"],
          theme["color-qualitative-8"],
          theme["color-qualitative-9"],
          theme["color-qualitative-10"],
          theme["color-qualitative-11"],
          theme["color-qualitative-12"],
          theme["color-qualitative-13"],
          theme["color-qualitative-14"],
          theme["color-qualitative-15"],
          theme["color-qualitative-16"],
          theme["color-qualitative-17"],
          theme["color-qualitative-18"],
          theme["color-qualitative-19"],
          theme["color-qualitative-20"],
          theme["color-qualitative-21"],
          theme["color-qualitative-22"],
          theme["color-qualitative-23"],
          theme["color-qualitative-24"],
        ].filter(Boolean);

        return themeColors.length > 0 ? themeColors : defaultColors;
      })(),
      heatmap: (() => {
        // Use palette store which respects scoped themes and converts to hex
        const defaultColors = getSequentialColorsAsHex();

        if (!theme) return defaultColors;

        const themeColors = [
          theme["color-sequential-1"],
          theme["color-sequential-2"],
          theme["color-sequential-3"],
          theme["color-sequential-4"],
          theme["color-sequential-5"],
          theme["color-sequential-6"],
          theme["color-sequential-7"],
          theme["color-sequential-8"],
          theme["color-sequential-9"],
        ].filter(Boolean);

        return themeColors.length > 0 ? themeColors : defaultColors;
      })(),
    },
    numberFormat: "s",
    tooltipFormat: {
      numberFormat: "d",
    },
    customFormatTypes: true,
  };
};
