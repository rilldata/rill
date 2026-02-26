import type { ColorRangeMapping } from "@rilldata/web-common/features/components/charts/types";
import {
  getSequentialColorsAsHex,
  getDivergingColorsAsHex,
} from "@rilldata/web-common/features/themes/palette-store";
import chroma, { type Color } from "chroma-js";
import * as d3sc from "d3-scale-chromatic";

type MapTheme = { primary: Color; secondary: Color };

/**
 * Mapping of Vega/D3 scheme names to d3-scale-chromatic interpolators.
 */
const schemeInterpolators: Record<string, ((t: number) => string) | undefined> =
  {
    tealblues: d3sc.interpolateGnBu,
    viridis: d3sc.interpolateViridis,
    magma: d3sc.interpolateMagma,
    inferno: d3sc.interpolateInferno,
    plasma: d3sc.interpolatePlasma,
    cividis: d3sc.interpolateCividis,
    blues: d3sc.interpolateBlues,
    teals: d3sc.interpolateGnBu,
    greens: d3sc.interpolateGreens,
    greys: d3sc.interpolateGreys,
    oranges: d3sc.interpolateOranges,
    purples: d3sc.interpolatePurples,
    reds: d3sc.interpolateReds,
    turbo: d3sc.interpolateTurbo,
    spectral: d3sc.interpolateSpectral,
  };

/**
 * Resolves a theme color reference ("primary"/"secondary") or hex to a hex string.
 */
export function resolveStaticColor(color: string, theme: MapTheme): string {
  if (color === "primary") return theme.primary.hex();
  if (color === "secondary") return theme.secondary.hex();
  return color;
}

/**
 * Resolves a ColorRangeMapping to an array of hex color strings
 * for generating Mapbox interpolation stops.
 */
export function resolveColorRange(
  colorRange: ColorRangeMapping,
  theme: MapTheme,
  steps = 7,
): string[] {
  if (colorRange.mode === "gradient") {
    const start = resolveStaticColor(colorRange.start, theme);
    const end = resolveStaticColor(colorRange.end, theme);
    return chroma.scale([start, end]).mode("lab").colors(steps);
  }

  // Scheme mode
  if (colorRange.scheme === "sequential") {
    return getSequentialColorsAsHex();
  }
  if (colorRange.scheme === "diverging") {
    return getDivergingColorsAsHex();
  }

  // Named D3/Vega scheme
  const interpolator = schemeInterpolators[colorRange.scheme as string];
  if (interpolator) {
    return Array.from({ length: steps }, (_, i) => {
      const t = i / (steps - 1);
      return chroma(interpolator(t)).hex();
    });
  }

  // Fallback to tealblues
  return Array.from({ length: steps }, (_, i) => {
    const t = i / (steps - 1);
    return chroma(d3sc.interpolateGnBu(t)).hex();
  });
}

/**
 * Computes the min and max values of a numeric property across GeoJSON features.
 */
export function computeMinMax(
  features: GeoJSON.Feature[],
  property: string,
): [number, number] {
  let min = Infinity;
  let max = -Infinity;
  for (const f of features) {
    const val = f.properties?.[property];
    if (typeof val === "number" && isFinite(val)) {
      min = Math.min(min, val);
      max = Math.max(max, val);
    }
  }
  if (min === Infinity) return [0, 1];
  if (min === max) return [min, min + 1];
  return [min, max];
}

/**
 * Builds a Mapbox GL interpolate expression for coloring features
 * by a numeric property using the given color stops.
 */
export function buildColorExpression(
  property: string,
  min: number,
  max: number,
  colors: string[],
): unknown[] {
  const stops: (number | string)[] = [];
  for (let i = 0; i < colors.length; i++) {
    const t = min + (i / (colors.length - 1)) * (max - min);
    stops.push(t, colors[i]);
  }

  return [
    "interpolate",
    ["linear"],
    ["coalesce", ["get", property], min],
    ...stops,
  ];
}

/**
 * Builds a Mapbox GL interpolate expression for sizing circles
 * by a numeric property.
 */
export function buildSizeExpression(
  property: string,
  min: number,
  max: number,
  minRadius = 4,
  maxRadius = 20,
): unknown[] {
  return [
    "interpolate",
    ["linear"],
    ["coalesce", ["get", property], min],
    min,
    minRadius,
    max,
    maxRadius,
  ];
}
