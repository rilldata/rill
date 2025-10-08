export const MEASURE_CONFIG = {
  breakpoint: 960,
  chart: {
    height: 125,
    fullHeight: 220,
  },
  bigNumber: {
    widthWithChart: 140,
    widthWithoutChart: {
      1: "250px",
      2: "400px",
      3: "580px",
    },
  },
};

/**
 * Comparison colors using the qualitative palette
 * For categorical distinction when comparing different dimension values
 */
export const COMPARIONS_COLORS = [
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

export const SELECTED_NOT_COMPARED_COLOR = "#d1d5db";
