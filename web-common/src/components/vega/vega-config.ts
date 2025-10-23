import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import chroma from "chroma-js";
import type { Config } from "vega-lite";

/**
 * Resolves a CSS variable to its computed value for canvas rendering
 * This is necessary because canvas rendering requires concrete color values, not CSS variables
 * 
 * For palette variables (--color-theme-600), explicitly resolves to light/dark variant
 * For theme variables (--color-sequential-1), reads the computed value considering dark mode
 */
function resolveCSSVariable(cssVar: string, isDarkMode: boolean, fallback?: string): string {
  if (typeof window === "undefined") return fallback || cssVar;
  
  // Extract the variable name from var() syntax
  const varName = cssVar.replace("var(", "").replace(")", "").split(",")[0].trim();
  
  // For theme palette variables (--color-theme-600, --color-primary-500, etc), 
  // these use light-dark() CSS function, so resolve to explicit light/dark variant
  const palettePattern = /^--color-(theme|primary|secondary|theme-secondary)-(\d+)$/;
  const match = varName.match(palettePattern);
  
  if (match) {
    const [, colorType, shade] = match;
    const modeVariant = isDarkMode ? `--color-${colorType}-dark-${shade}` : `--color-${colorType}-light-${shade}`;
    
    // Try scoped theme boundary first
    const themeBoundary = document.querySelector(".dashboard-theme-boundary");
    if (themeBoundary) {
      const scopedValue = getComputedStyle(themeBoundary as HTMLElement).getPropertyValue(modeVariant);
      if (scopedValue && scopedValue.trim()) {
        return scopedValue.trim();
      }
    }
    
    // Fall back to document root
    const computed = getComputedStyle(document.documentElement).getPropertyValue(modeVariant);
    if (computed && computed.trim()) {
      return computed.trim();
    }
  }
  
  // For other variables (--color-sequential-1, --primary, custom vars from theme),
  // read directly - the CSS rules handle light/dark switching via :root vs :root.dark selectors
  // We just need to ensure we're reading when the .dark class state matches isDarkMode
  const themeBoundary = document.querySelector(".dashboard-theme-boundary");
  if (themeBoundary) {
    const scopedValue = getComputedStyle(themeBoundary as HTMLElement).getPropertyValue(varName);
    if (scopedValue && scopedValue.trim()) {
      return scopedValue.trim();
    }
  }
  
  // Fall back to document root
  const computed = getComputedStyle(document.documentElement).getPropertyValue(varName);
  if (computed && computed.trim()) {
    return computed.trim();
  }
  
  // If fallback is provided and is also a CSS variable, resolve it
  if (fallback && fallback.startsWith("var(")) {
    return resolveCSSVariable(fallback, isDarkMode);
  }
  
  return fallback || cssVar;
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
  let lineColor, barColor, areaGradientLight, areaGradientDark;
  
  if (theme?.primary) {
    // Use theme's primary color directly
    const primary = chroma(theme.primary);
    lineColor = primary.darken(0.2).css();
    barColor = primary.css();
    areaGradientLight = primary.brighten(2).css();
    areaGradientDark = primary.darken(0.5).css();
  } else {
    // Fallback: resolve from CSS variables (for standalone charts without theme context)
    lineColor = resolveCSSVariable("var(--color-theme-600)", isDarkMode, "var(--color-primary-600)");
    barColor = resolveCSSVariable("var(--color-theme-400)", isDarkMode, "var(--color-primary-400)");
    areaGradientLight = resolveCSSVariable("var(--color-theme-50)", isDarkMode, "var(--color-primary-50)");
    areaGradientDark = resolveCSSVariable("var(--color-theme-300)", isDarkMode, "var(--color-primary-300)");
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
          color.startsWith("var(") ? resolveCSSVariable(color, isDarkMode) : color
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
        const defaultColors = [
          resolveCSSVariable("var(--color-sequential-1)", isDarkMode),
          resolveCSSVariable("var(--color-sequential-2)", isDarkMode),
          resolveCSSVariable("var(--color-sequential-3)", isDarkMode),
          resolveCSSVariable("var(--color-sequential-4)", isDarkMode),
          resolveCSSVariable("var(--color-sequential-5)", isDarkMode),
          resolveCSSVariable("var(--color-sequential-6)", isDarkMode),
          resolveCSSVariable("var(--color-sequential-7)", isDarkMode),
          resolveCSSVariable("var(--color-sequential-8)", isDarkMode),
          resolveCSSVariable("var(--color-sequential-9)", isDarkMode),
        ];
        
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
