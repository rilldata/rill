import { derived, writable, type Readable } from "svelte/store";
import { themeControl } from "./theme-control";
import chroma from "chroma-js";

/**
 * Palette types for data visualization
 */
export type PaletteType = "sequential" | "diverging" | "qualitative";

/**
 * Sequential palette colors (9 colors) - for ordered data
 */
export interface SequentialColors {
  sequential1: string;
  sequential2: string;
  sequential3: string;
  sequential4: string;
  sequential5: string;
  sequential6: string;
  sequential7: string;
  sequential8: string;
  sequential9: string;
}

/**
 * Diverging palette colors (11 colors) - for data that diverges from a midpoint
 */
export interface DivergingColors {
  diverging1: string;
  diverging2: string;
  diverging3: string;
  diverging4: string;
  diverging5: string;
  diverging6: string;
  diverging7: string;
  diverging8: string;
  diverging9: string;
  diverging10: string;
  diverging11: string;
}

/**
 * Qualitative palette colors (24 colors) - for categorical data
 */
export interface QualitativeColors {
  qualitative1: string;
  qualitative2: string;
  qualitative3: string;
  qualitative4: string;
  qualitative5: string;
  qualitative6: string;
  qualitative7: string;
  qualitative8: string;
  qualitative9: string;
  qualitative10: string;
  qualitative11: string;
  qualitative12: string;
  qualitative13: string;
  qualitative14: string;
  qualitative15: string;
  qualitative16: string;
  qualitative17: string;
  qualitative18: string;
  qualitative19: string;
  qualitative20: string;
  qualitative21: string;
  qualitative22: string;
  qualitative23: string;
  qualitative24: string;
}

/**
 * All palette colors combined
 */
export interface AllPaletteColors {
  sequential: SequentialColors;
  diverging: DivergingColors;
  qualitative: QualitativeColors;
}

// Store instances
const sequentialStore = writable<SequentialColors>({
  sequential1: "",
  sequential2: "",
  sequential3: "",
  sequential4: "",
  sequential5: "",
  sequential6: "",
  sequential7: "",
  sequential8: "",
  sequential9: "",
});

const divergingStore = writable<DivergingColors>({
  diverging1: "",
  diverging2: "",
  diverging3: "",
  diverging4: "",
  diverging5: "",
  diverging6: "",
  diverging7: "",
  diverging8: "",
  diverging9: "",
  diverging10: "",
  diverging11: "",
});

const qualitativeStore = writable<QualitativeColors>({
  qualitative1: "",
  qualitative2: "",
  qualitative3: "",
  qualitative4: "",
  qualitative5: "",
  qualitative6: "",
  qualitative7: "",
  qualitative8: "",
  qualitative9: "",
  qualitative10: "",
  qualitative11: "",
  qualitative12: "",
  qualitative13: "",
  qualitative14: "",
  qualitative15: "",
  qualitative16: "",
  qualitative17: "",
  qualitative18: "",
  qualitative19: "",
  qualitative20: "",
  qualitative21: "",
  qualitative22: "",
  qualitative23: "",
  qualitative24: "",
});

/**
 * Reads the computed CSS variable value
 * Checks scoped theme boundary first (Canvas/dashboards), then falls back to document root
 */
function getCSSVariableValue(variableName: string): string {
  if (typeof window === "undefined") return "";

  // Check for scoped theme boundary first (Canvas/dashboards use this for theme isolation)
  const themeBoundary = document.querySelector(".dashboard-theme-boundary");
  if (themeBoundary) {
    const scopedValue = getComputedStyle(themeBoundary as HTMLElement)
      .getPropertyValue(variableName)
      .trim();
    if (scopedValue) return scopedValue;
  }

  // Fall back to document root
  return getComputedStyle(document.documentElement)
    .getPropertyValue(variableName)
    .trim();
}

/**
 * Generic helper to build palette colors object from CSS variables
 */
function buildPaletteColors<T>(prefix: string, count: number): T {
  const colors: Record<string, string> = {};
  for (let i = 1; i <= count; i++) {
    const key = `${prefix}${i}`;
    colors[key] = getCSSVariableValue(`--color-${prefix}-${i}`);
  }
  return colors as T;
}

/**
 * Updates sequential palette colors from CSS variables
 */
function updateSequentialColors() {
  sequentialStore.set(buildPaletteColors<SequentialColors>("sequential", 9));
}

/**
 * Updates diverging palette colors from CSS variables
 */
function updateDivergingColors() {
  divergingStore.set(buildPaletteColors<DivergingColors>("diverging", 11));
}

/**
 * Updates qualitative palette colors from CSS variables
 */
function updateQualitativeColors() {
  qualitativeStore.set(
    buildPaletteColors<QualitativeColors>("qualitative", 24),
  );
}

/**
 * Updates all palette colors
 */
function updateAllPaletteColors() {
  updateSequentialColors();
  updateDivergingColors();
  updateQualitativeColors();
}

/**
 * Initialize the palette colors store
 */
function initializePaletteStore() {
  if (typeof window === "undefined") return;

  // Initial update
  updateAllPaletteColors();

  // Update when theme changes (light/dark mode)
  // This handles the primary case of palette updates
  themeControl.subscribe(() => {
    setTimeout(updateAllPaletteColors, 0);
  });

  // Watch for programmatic style changes on document element
  // This catches setPaletteColor() and similar programmatic updates
  const observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      if (
        mutation.type === "attributes" &&
        mutation.attributeName === "style"
      ) {
        updateAllPaletteColors();
        break;
      }
    }
  });

  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ["style"],
  });
}

// Initialize on module load (client-side only)
if (typeof window !== "undefined") {
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initializePaletteStore);
  } else {
    initializePaletteStore();
  }
}

/**
 * Reactive store for sequential colors (9 colors for ordered data)
 * Sequential palettes are designed for data that progresses from low to high values
 */
export const sequentialColors: Readable<SequentialColors> =
  sequentialStore as Readable<SequentialColors>;

/**
 * Sequential colors as an array
 */
export const sequentialColorsArray: Readable<string[]> = derived(
  sequentialColors,
  ($colors) => Object.values($colors) as string[],
);

/**
 * Reactive store for diverging colors (11 colors for data with a midpoint)
 * Diverging palettes emphasize deviations from a critical midpoint value
 */
export const divergingColors: Readable<DivergingColors> =
  divergingStore as Readable<DivergingColors>;

/**
 * Diverging colors as an array
 */
export const divergingColorsArray: Readable<string[]> = derived(
  divergingColors,
  ($colors) => Object.values($colors) as string[],
);

/**
 * Reactive store for qualitative colors (24 colors for categorical data)
 * Qualitative palettes are designed for categorical data without inherent ordering
 */
export const qualitativeColors: Readable<QualitativeColors> =
  qualitativeStore as Readable<QualitativeColors>;

/**
 * Qualitative colors as an array
 */
export const qualitativeColorsArray: Readable<string[]> = derived(
  qualitativeColors,
  ($colors) => Object.values($colors) as string[],
);

/**
 * Generic helper to get a specific palette color by index
 */
function getPaletteColor<T>(
  store: Readable<T>,
  prefix: string,
  index: number,
  max: number,
): Readable<string> {
  if (index < 1 || index > max) {
    throw new Error(`${prefix} color index must be between 1 and ${max}`);
  }
  return derived(
    store,
    ($colors) => ($colors as Record<string, string>)[`${prefix}${index}`],
  );
}

/**
 * Get a specific sequential color by index (1-9)
 */
export function getSequentialColor(index: number): Readable<string> {
  return getPaletteColor(sequentialColors, "sequential", index, 9);
}

/**
 * Get a specific diverging color by index (1-11)
 */
export function getDivergingColor(index: number): Readable<string> {
  return getPaletteColor(divergingColors, "diverging", index, 11);
}

/**
 * Get a specific qualitative color by index (1-24)
 */
export function getQualitativeColor(index: number): Readable<string> {
  return getPaletteColor(qualitativeColors, "qualitative", index, 24);
}

/**
 * Get current sequential colors as a plain object (non-reactive)
 */
export function getSequentialColorsSnapshot(): SequentialColors {
  return buildPaletteColors<SequentialColors>("sequential", 9);
}

/**
 * Get current diverging colors as a plain object (non-reactive)
 */
export function getDivergingColorsSnapshot(): DivergingColors {
  return buildPaletteColors<DivergingColors>("diverging", 11);
}

/**
 * Get current qualitative colors as a plain object (non-reactive)
 */
export function getQualitativeColorsSnapshot(): QualitativeColors {
  return buildPaletteColors<QualitativeColors>("qualitative", 24);
}

/**
 * Convert a CSS color value to hex format
 */
export function colorToHex(cssColor: string): string {
  try {
    return chroma(cssColor).hex();
  } catch {
    return "#000000";
  }
}

/**
 * Get sequential colors as hex values
 */
export function getSequentialColorsAsHex(): string[] {
  const colors = getSequentialColorsSnapshot();
  return Object.values(colors).map(colorToHex);
}

/**
 * Get diverging colors as hex values
 */
export function getDivergingColorsAsHex(): string[] {
  const colors = getDivergingColorsSnapshot();
  return Object.values(colors).map(colorToHex);
}

/**
 * Get qualitative colors as hex values
 */
export function getQualitativeColorsAsHex(): string[] {
  const colors = getQualitativeColorsSnapshot();
  return Object.values(colors).map(colorToHex);
}
