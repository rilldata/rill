import {
  getDivergingColorsAsHex,
  getSequentialColorsAsHex,
} from "@rilldata/web-common/features/themes/palette-store";
import chroma from "chroma-js";
import type { PivotFormatRule, PivotMeasureFormatting } from "./types";

/**
 * Conditional formatting color schemes available per measure. Theme-aware
 * schemes resolve their stops from the active theme's palettes at call time, so
 * `getStops` is a function rather than a static array.
 */
export interface PivotFormattingScheme {
  key: string;
  label: string;
  getStops: () => string[];
}

export const PIVOT_FORMATTING_SCHEMES: PivotFormattingScheme[] = [
  {
    key: "theme-sequential",
    label: "Theme",
    getStops: () => getSequentialColorsAsHex(),
  },
  {
    key: "theme-diverging",
    label: "Theme (diverging)",
    getStops: () => getDivergingColorsAsHex(),
  },
  {
    key: "greens",
    label: "Greens",
    getStops: () => ["#f7fcf5", "#74c476", "#00441b"],
  },
  {
    key: "blues",
    label: "Blues",
    getStops: () => ["#f7fbff", "#6baed6", "#08306b"],
  },
  {
    key: "redYellowGreen",
    label: "Red–Yellow–Green",
    getStops: () => ["#d73027", "#ffffbf", "#1a9850"],
  },
];

const DEFAULT_SCHEME_KEY = "theme-sequential";

function getScheme(schemeKey: string): PivotFormattingScheme {
  return (
    PIVOT_FORMATTING_SCHEMES.find((s) => s.key === schemeKey) ??
    PIVOT_FORMATTING_SCHEMES.find((s) => s.key === DEFAULT_SCHEME_KEY)!
  );
}

/**
 * Resolve a scheme's ordered color stops, reversing them when requested.
 */
export function getSchemeStops(schemeKey: string, reverse = false): string[] {
  const stops = getScheme(schemeKey).getStops();
  return reverse ? [...stops].reverse() : stops;
}

/**
 * Accumulate per-measure [min, max] domains from a flat list of measure values.
 * Measures with no finite values are omitted.
 */
export function computeMeasureDomains(
  entries: Iterable<{ measureName: string; value: number }>,
): Map<string, [number, number]> {
  const domains = new Map<string, [number, number]>();
  for (const { measureName, value } of entries) {
    if (!Number.isFinite(value)) continue;
    const existing = domains.get(measureName);
    if (!existing) {
      domains.set(measureName, [value, value]);
    } else {
      if (value < existing[0]) existing[0] = value;
      if (value > existing[1]) existing[1] = value;
    }
  }
  return domains;
}

// Semantic rule colors, modeled on the classic Excel conditional-formatting
// fills so thresholds read instantly as good/bad/caution. Each pairs a light
// fill with a legible dark text color.
export interface PivotRuleColor {
  key: string;
  label: string;
  fill: string;
  text: string;
}

export const PIVOT_RULE_COLORS: PivotRuleColor[] = [
  { key: "positive", label: "Positive", fill: "#c6efce", text: "#006100" },
  { key: "negative", label: "Negative", fill: "#ffc7ce", text: "#9c0d07" },
  { key: "warning", label: "Warning", fill: "#ffeb9c", text: "#9c6500" },
  { key: "info", label: "Info", fill: "#cfe2f3", text: "#0b5394" },
  { key: "neutral", label: "Neutral", fill: "#e5e7eb", text: "#374151" },
];

/**
 * Resolve a rule's stored color (semantic key or raw hex, with or without "#")
 * to fill and text hex colors. Unknown/invalid values fall back to the neutral
 * semantic color so a malformed URL never breaks rendering.
 */
export function resolveRuleColor(color: string): {
  fill: string;
  text: string;
} {
  const semantic = PIVOT_RULE_COLORS.find((c) => c.key === color);
  if (semantic) return { fill: semantic.fill, text: semantic.text };
  const hex = color.startsWith("#") ? color : `#${color}`;
  if (!chroma.valid(hex)) {
    const neutral = PIVOT_RULE_COLORS.find((c) => c.key === "neutral")!;
    return { fill: neutral.fill, text: neutral.text };
  }
  return {
    fill: hex,
    text:
      chroma(hex).luminance() < LIGHT_TEXT_LUMINANCE ? "#ffffff" : "inherit",
  };
}

export function ruleMatches(rule: PivotFormatRule, value: number): boolean {
  switch (rule.operator) {
    case "gt":
      return value > rule.value;
    case "gte":
      return value >= rule.value;
    case "lt":
      return value < rule.value;
    case "lte":
      return value <= rule.value;
    case "eq":
      return value === rule.value;
    case "between":
      return value >= rule.value && value <= (rule.value2 ?? rule.value);
  }
}

export interface CellFormat {
  background: string;
  color: string;
}

// Returns the style for a cell value, or null when the value should render
// unformatted (e.g. no threshold rule matched).
export type CellFormatter = (value: number) => CellFormat | null;

// Below this background luminance, switch to light text for legibility. Mirrors
// the threshold used by the canvas heatmap (charts/heatmap/spec.ts).
const LIGHT_TEXT_LUMINANCE = 0.4;

/**
 * Build a cell formatter for a single measure. Scale modes (heatmap, data_bar)
 * color across the measure's value domain; rules mode applies the first
 * matching threshold rule and needs no domain.
 */
export function makeCellFormatter(
  fmt: PivotMeasureFormatting,
  domain?: [number, number],
  lowerIsBetter = false,
): CellFormatter {
  if (fmt.mode === "rules") {
    return (value: number) => {
      const match = fmt.rules.find((rule) => ruleMatches(rule, value));
      if (!match) return null;
      const { fill, text } = resolveRuleColor(match.color);
      return { background: fill, color: text };
    };
  }

  const [min, max] = domain ?? [0, 1];
  // Data bars encode magnitude only, so direction doesn't apply to them.
  const stops = getSchemeStops(
    fmt.scheme,
    fmt.mode === "heatmap" && lowerIsBetter,
  );

  if (fmt.mode === "data_bar") {
    // Bar length is proportional to magnitude relative to the largest absolute
    // value, mirroring the leaderboard bar (LeaderboardRow.svelte).
    const absMax = Math.max(Math.abs(min), Math.abs(max)) || 1;
    // Use the scheme's most saturated stop, lightened so right-aligned numbers
    // stay legible over the bar.
    const barColor = chroma(stops[stops.length - 1])
      .luminance(0.72)
      .hex();
    return (value: number) => {
      const pct = Math.min(100, (Math.abs(value) / absMax) * 100);
      return {
        background: `linear-gradient(to right, ${barColor} ${pct}%, transparent ${pct}%)`,
        color: "inherit",
      };
    };
  }

  // Heatmap: interpolate a solid background color across the domain. Guard
  // against a degenerate domain (all values equal) where chroma cannot scale.
  const scale = chroma
    .scale(stops)
    .domain(min === max ? [min, min + 1] : [min, max]);
  return (value: number) => ({
    background: scale(value).hex(),
    color:
      scale(value).luminance() < LIGHT_TEXT_LUMINANCE ? "#ffffff" : "inherit",
  });
}
