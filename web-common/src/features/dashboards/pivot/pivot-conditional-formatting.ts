import {
  getDivergingColorsAsHex,
  getSequentialColorsAsHex,
} from "@rilldata/web-common/features/themes/palette-store";
import chroma from "chroma-js";
import type { PivotMeasureFormatting } from "./types";

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

export interface CellFormatter {
  background: (value: number) => string;
  textColor: (value: number) => string;
}

// Below this background luminance, switch to light text for legibility. Mirrors
// the threshold used by the canvas heatmap (charts/heatmap/spec.ts).
const LIGHT_TEXT_LUMINANCE = 0.4;

/**
 * Build a background/text-color formatter for a single measure given its value
 * domain and formatting config. Handles both heatmap (solid gradient color) and
 * data-bar (proportional in-cell bar) modes.
 */
export function makeCellFormatter(
  domain: [number, number],
  fmt: PivotMeasureFormatting,
): CellFormatter {
  const [min, max] = domain;
  const stops = getSchemeStops(fmt.scheme, fmt.reverse);

  if (fmt.mode === "data_bar") {
    // Bar length is proportional to magnitude relative to the largest absolute
    // value, mirroring the leaderboard bar (LeaderboardRow.svelte).
    const absMax = Math.max(Math.abs(min), Math.abs(max)) || 1;
    // Use the scheme's most saturated stop, lightened so right-aligned numbers
    // stay legible over the bar.
    const barColor = chroma(stops[stops.length - 1])
      .luminance(0.72)
      .hex();
    return {
      background: (value: number) => {
        const pct = Math.min(100, (Math.abs(value) / absMax) * 100);
        return `linear-gradient(to right, ${barColor} ${pct}%, transparent ${pct}%)`;
      },
      textColor: () => "inherit",
    };
  }

  // Heatmap: interpolate a solid background color across the domain. Guard
  // against a degenerate domain (all values equal) where chroma cannot scale.
  const scale = chroma
    .scale(stops)
    .domain(min === max ? [min, min + 1] : [min, max]);
  return {
    background: (value: number) => scale(value).hex(),
    textColor: (value: number) =>
      scale(value).luminance() < LIGHT_TEXT_LUMINANCE ? "#ffffff" : "inherit",
  };
}
