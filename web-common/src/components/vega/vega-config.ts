import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import {
  MainAreaColorGradientDark,
  MainAreaColorGradientLight,
  MainLineColor,
} from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { Config } from "vega-lite";

/**
 * Resolves a CSS variable to its computed value
 * This is necessary for canvas rendering where CSS variables need to be resolved
 * Checks scoped theme boundary first, then falls back to document root
 */
function resolveCSSVariable(cssVar: string, fallback?: string): string {
  if (typeof window === "undefined") return fallback || cssVar;
  
  // Extract the variable name from var() syntax
  const varName = cssVar.replace("var(", "").replace(")", "").split(",")[0].trim();
  
  // First check if there's a dashboard-theme-boundary element (scoped themes)
  const themeBoundary = document.querySelector(".dashboard-theme-boundary");
  if (themeBoundary) {
    const scopedValue = getComputedStyle(themeBoundary as HTMLElement).getPropertyValue(varName);
    if (scopedValue && scopedValue.trim()) {
      return scopedValue.trim();
    }
  }
  
  // Fall back to document root for global variables
  const computed = getComputedStyle(document.documentElement).getPropertyValue(varName);
  
  if (computed && computed.trim()) {
    return computed.trim();
  }
  
  // If fallback is provided and is also a CSS variable, resolve it
  if (fallback && fallback.startsWith("var(")) {
    return resolveCSSVariable(fallback);
  }
  
  return fallback || cssVar;
}

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
  const surfaceColor = isDarkMode ? colors.dark.surface : colors.light.surface;
  
  // Resolve colors at render time for canvas rendering
  const lineColor = resolveCSSVariable("var(--color-theme-600)", "var(--color-primary-600)");
  const barColor = resolveCSSVariable("var(--color-theme-400)", "var(--color-primary-400)");
  const areaGradientLight = resolveCSSVariable("var(--color-theme-50)", "var(--color-primary-50)");
  const areaGradientDark = resolveCSSVariable("var(--color-theme-300)", "var(--color-primary-300)");

  return {
    autosize: {
      type: "fit-x",
    },
    background: surfaceColor,
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
      // Resolve qualitative palette colors for categorical data
      category: COMPARIONS_COLORS.map((color) => 
        color.startsWith("var(") ? resolveCSSVariable(color) : color
      ),
      // Resolve sequential palette colors for heatmaps
      heatmap: [
        resolveCSSVariable("var(--color-sequential-1)"),
        resolveCSSVariable("var(--color-sequential-2)"),
        resolveCSSVariable("var(--color-sequential-3)"),
        resolveCSSVariable("var(--color-sequential-4)"),
        resolveCSSVariable("var(--color-sequential-5)"),
        resolveCSSVariable("var(--color-sequential-6)"),
        resolveCSSVariable("var(--color-sequential-7)"),
        resolveCSSVariable("var(--color-sequential-8)"),
        resolveCSSVariable("var(--color-sequential-9)"),
      ],
    },
    numberFormat: "s",
    tooltipFormat: {
      numberFormat: "d",
    },
    customFormatTypes: true,
  };
};
