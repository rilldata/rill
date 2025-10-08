/**
 * Palette Integration for Canvas Charts
 * 
 * This file provides utilities and documentation for using color palettes
 * in Canvas charts.
 */

/**
 * Available palette schemes for quantitative color encodings
 * 
 * Use these scheme names in your chart's colorRange configuration:
 * 
 * @example
 * ```typescript
 * {
 *   color: {
 *     type: "quantitative",
 *     field: "temperature",
 *     colorRange: {
 *       mode: "scheme",
 *       scheme: "sequential"  // or "diverging"
 *     }
 *   }
 * }
 * ```
 */
export type PaletteScheme = "sequential" | "diverging";

/**
 * Chart type to recommended palette mapping
 */
export const CHART_PALETTE_RECOMMENDATIONS = {
  // Sequential palette: for ordered data that progresses from low to high
  sequential: {
    chartTypes: ["heatmap"],
    useCase: "Heat maps, density maps, choropleth maps, progress indicators",
    examples: [
      "Temperature ranges",
      "Population density",
      "Sales over time",
      "Performance scores",
    ],
  },
  
  // Diverging palette: for data with a critical midpoint
  diverging: {
    chartTypes: ["heatmap"],
    useCase: "Data that emphasizes deviation from a midpoint",
    examples: [
      "Profit vs Loss",
      "Above/Below average",
      "Temperature anomalies",
      "Sentiment analysis",
    ],
  },
  
  // Qualitative palette: for categorical data (automatically used)
  qualitative: {
    chartTypes: ["bar", "line", "pie", "donut", "funnel", "combo"],
    useCase: "Categorical data without inherent ordering",
    examples: [
      "Product categories",
      "Geographic regions",
      "Department names",
      "User segments",
    ],
    note: "Automatically applied to categorical color fields (dimensions)",
  },
} as const;

/**
 * Get the sequential palette colors as an array of CSS variable strings
 * Useful for programmatic color generation
 */
export function getSequentialPaletteRange(): string[] {
  return [
    "var(--color-sequential-1)",
    "var(--color-sequential-2)",
    "var(--color-sequential-3)",
    "var(--color-sequential-4)",
    "var(--color-sequential-5)",
    "var(--color-sequential-6)",
    "var(--color-sequential-7)",
    "var(--color-sequential-8)",
    "var(--color-sequential-9)",
  ];
}

/**
 * Get the diverging palette colors as an array of CSS variable strings
 */
export function getDivergingPaletteRange(): string[] {
  return [
    "var(--color-diverging-1)",
    "var(--color-diverging-2)",
    "var(--color-diverging-3)",
    "var(--color-diverging-4)",
    "var(--color-diverging-5)",
    "var(--color-diverging-6)",
    "var(--color-diverging-7)",
    "var(--color-diverging-8)",
    "var(--color-diverging-9)",
    "var(--color-diverging-10)",
    "var(--color-diverging-11)",
  ];
}

/**
 * Get the qualitative palette colors as an array of CSS variable strings
 */
export function getQualitativePaletteRange(): string[] {
  return [
    "var(--color-qualitative-1)",
    "var(--color-qualitative-2)",
    "var(--color-qualitative-3)",
    "var(--color-qualitative-4)",
    "var(--color-qualitative-5)",
    "var(--color-qualitative-6)",
    "var(--color-qualitative-7)",
    "var(--color-qualitative-8)",
    "var(--color-qualitative-9)",
    "var(--color-qualitative-10)",
    "var(--color-qualitative-11)",
    "var(--color-qualitative-12)",
  ];
}

/**
 * Recommended color scheme based on chart type and data characteristics
 */
export function getRecommendedColorScheme(
  chartType: string,
  dataType: "quantitative" | "nominal" | "ordinal" | "temporal",
  hasNaturalMidpoint = false,
): PaletteScheme | "qualitative" | null {
  // For quantitative color encodings (like heatmaps)
  if (dataType === "quantitative") {
    // If data has a natural midpoint (0, average, etc.), use diverging
    if (hasNaturalMidpoint) {
      return "diverging";
    }
    // Otherwise use sequential for ordered data
    return "sequential";
  }
  
  // For categorical data (nominal/ordinal), use qualitative
  if (dataType === "nominal" || dataType === "ordinal") {
    return "qualitative";
  }
  
  return null;
}

