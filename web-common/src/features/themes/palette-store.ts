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
});

/**
 * Reads the computed CSS variable value from the document element
 */
function getCSSVariableValue(variableName: string): string {
  if (typeof window === "undefined") return "";
  return getComputedStyle(document.documentElement)
    .getPropertyValue(variableName)
    .trim();
}

/**
 * Updates sequential palette colors from CSS variables
 */
function updateSequentialColors() {
  const colors: SequentialColors = {
    sequential1: getCSSVariableValue("--color-sequential-1"),
    sequential2: getCSSVariableValue("--color-sequential-2"),
    sequential3: getCSSVariableValue("--color-sequential-3"),
    sequential4: getCSSVariableValue("--color-sequential-4"),
    sequential5: getCSSVariableValue("--color-sequential-5"),
    sequential6: getCSSVariableValue("--color-sequential-6"),
    sequential7: getCSSVariableValue("--color-sequential-7"),
    sequential8: getCSSVariableValue("--color-sequential-8"),
    sequential9: getCSSVariableValue("--color-sequential-9"),
  };
  sequentialStore.set(colors);
}

/**
 * Updates diverging palette colors from CSS variables
 */
function updateDivergingColors() {
  const colors: DivergingColors = {
    diverging1: getCSSVariableValue("--color-diverging-1"),
    diverging2: getCSSVariableValue("--color-diverging-2"),
    diverging3: getCSSVariableValue("--color-diverging-3"),
    diverging4: getCSSVariableValue("--color-diverging-4"),
    diverging5: getCSSVariableValue("--color-diverging-5"),
    diverging6: getCSSVariableValue("--color-diverging-6"),
    diverging7: getCSSVariableValue("--color-diverging-7"),
    diverging8: getCSSVariableValue("--color-diverging-8"),
    diverging9: getCSSVariableValue("--color-diverging-9"),
    diverging10: getCSSVariableValue("--color-diverging-10"),
    diverging11: getCSSVariableValue("--color-diverging-11"),
  };
  divergingStore.set(colors);
}

/**
 * Updates qualitative palette colors from CSS variables
 */
function updateQualitativeColors() {
  const colors: QualitativeColors = {
    qualitative1: getCSSVariableValue("--color-qualitative-1"),
    qualitative2: getCSSVariableValue("--color-qualitative-2"),
    qualitative3: getCSSVariableValue("--color-qualitative-3"),
    qualitative4: getCSSVariableValue("--color-qualitative-4"),
    qualitative5: getCSSVariableValue("--color-qualitative-5"),
    qualitative6: getCSSVariableValue("--color-qualitative-6"),
    qualitative7: getCSSVariableValue("--color-qualitative-7"),
    qualitative8: getCSSVariableValue("--color-qualitative-8"),
    qualitative9: getCSSVariableValue("--color-qualitative-9"),
    qualitative10: getCSSVariableValue("--color-qualitative-10"),
    qualitative11: getCSSVariableValue("--color-qualitative-11"),
    qualitative12: getCSSVariableValue("--color-qualitative-12"),
    qualitative13: getCSSVariableValue("--color-qualitative-13"),
    qualitative14: getCSSVariableValue("--color-qualitative-14"),
    qualitative15: getCSSVariableValue("--color-qualitative-15"),
    qualitative16: getCSSVariableValue("--color-qualitative-16"),
    qualitative17: getCSSVariableValue("--color-qualitative-17"),
    qualitative18: getCSSVariableValue("--color-qualitative-18"),
    qualitative19: getCSSVariableValue("--color-qualitative-19"),
    qualitative20: getCSSVariableValue("--color-qualitative-20"),
    qualitative21: getCSSVariableValue("--color-qualitative-21"),
    qualitative22: getCSSVariableValue("--color-qualitative-22"),
    qualitative23: getCSSVariableValue("--color-qualitative-23"),
    qualitative24: getCSSVariableValue("--color-qualitative-24"),
  };
  qualitativeStore.set(colors);
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
  ($colors) => [
    $colors.sequential1,
    $colors.sequential2,
    $colors.sequential3,
    $colors.sequential4,
    $colors.sequential5,
    $colors.sequential6,
    $colors.sequential7,
    $colors.sequential8,
    $colors.sequential9,
  ],
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
  ($colors) => [
    $colors.diverging1,
    $colors.diverging2,
    $colors.diverging3,
    $colors.diverging4,
    $colors.diverging5,
    $colors.diverging6,
    $colors.diverging7,
    $colors.diverging8,
    $colors.diverging9,
    $colors.diverging10,
    $colors.diverging11,
  ],
);

/**
 * Reactive store for qualitative colors (12 colors for categorical data)
 * Qualitative palettes are designed for categorical data without inherent ordering
 */
export const qualitativeColors: Readable<QualitativeColors> =
  qualitativeStore as Readable<QualitativeColors>;

/**
 * Qualitative colors as an array
 */
export const qualitativeColorsArray: Readable<string[]> = derived(
  qualitativeColors,
  ($colors) => [
    $colors.qualitative1,
    $colors.qualitative2,
    $colors.qualitative3,
    $colors.qualitative4,
    $colors.qualitative5,
    $colors.qualitative6,
    $colors.qualitative7,
    $colors.qualitative8,
    $colors.qualitative9,
    $colors.qualitative10,
    $colors.qualitative11,
    $colors.qualitative12,
  ],
);

/**
 * Get a specific sequential color by index (1-9)
 */
export function getSequentialColor(index: number): Readable<string> {
  if (index < 1 || index > 9) {
    throw new Error("Sequential color index must be between 1 and 9");
  }
  return derived(sequentialColors, ($colors) => {
    const key = `sequential${index}` as keyof SequentialColors;
    return $colors[key];
  });
}

/**
 * Get a specific diverging color by index (1-11)
 */
export function getDivergingColor(index: number): Readable<string> {
  if (index < 1 || index > 11) {
    throw new Error("Diverging color index must be between 1 and 11");
  }
  return derived(divergingColors, ($colors) => {
    const key = `diverging${index}` as keyof DivergingColors;
    return $colors[key];
  });
}

/**
 * Get a specific qualitative color by index (1-12)
 */
export function getQualitativeColor(index: number): Readable<string> {
  if (index < 1 || index > 12) {
    throw new Error("Qualitative color index must be between 1 and 12");
  }
  return derived(qualitativeColors, ($colors) => {
    const key = `qualitative${index}` as keyof QualitativeColors;
    return $colors[key];
  });
}

/**
 * Get current sequential colors as a plain object (non-reactive)
 */
export function getSequentialColorsSnapshot(): SequentialColors {
  return {
    sequential1: getCSSVariableValue("--color-sequential-1"),
    sequential2: getCSSVariableValue("--color-sequential-2"),
    sequential3: getCSSVariableValue("--color-sequential-3"),
    sequential4: getCSSVariableValue("--color-sequential-4"),
    sequential5: getCSSVariableValue("--color-sequential-5"),
    sequential6: getCSSVariableValue("--color-sequential-6"),
    sequential7: getCSSVariableValue("--color-sequential-7"),
    sequential8: getCSSVariableValue("--color-sequential-8"),
    sequential9: getCSSVariableValue("--color-sequential-9"),
  };
}

/**
 * Get current diverging colors as a plain object (non-reactive)
 */
export function getDivergingColorsSnapshot(): DivergingColors {
  return {
    diverging1: getCSSVariableValue("--color-diverging-1"),
    diverging2: getCSSVariableValue("--color-diverging-2"),
    diverging3: getCSSVariableValue("--color-diverging-3"),
    diverging4: getCSSVariableValue("--color-diverging-4"),
    diverging5: getCSSVariableValue("--color-diverging-5"),
    diverging6: getCSSVariableValue("--color-diverging-6"),
    diverging7: getCSSVariableValue("--color-diverging-7"),
    diverging8: getCSSVariableValue("--color-diverging-8"),
    diverging9: getCSSVariableValue("--color-diverging-9"),
    diverging10: getCSSVariableValue("--color-diverging-10"),
    diverging11: getCSSVariableValue("--color-diverging-11"),
  };
}

/**
 * Get current qualitative colors as a plain object (non-reactive)
 */
export function getQualitativeColorsSnapshot(): QualitativeColors {
  return {
    qualitative1: getCSSVariableValue("--color-qualitative-1"),
    qualitative2: getCSSVariableValue("--color-qualitative-2"),
    qualitative3: getCSSVariableValue("--color-qualitative-3"),
    qualitative4: getCSSVariableValue("--color-qualitative-4"),
    qualitative5: getCSSVariableValue("--color-qualitative-5"),
    qualitative6: getCSSVariableValue("--color-qualitative-6"),
    qualitative7: getCSSVariableValue("--color-qualitative-7"),
    qualitative8: getCSSVariableValue("--color-qualitative-8"),
    qualitative9: getCSSVariableValue("--color-qualitative-9"),
    qualitative10: getCSSVariableValue("--color-qualitative-10"),
    qualitative11: getCSSVariableValue("--color-qualitative-11"),
    qualitative12: getCSSVariableValue("--color-qualitative-12"),
    qualitative13: getCSSVariableValue("--color-qualitative-13"),
    qualitative14: getCSSVariableValue("--color-qualitative-14"),
    qualitative15: getCSSVariableValue("--color-qualitative-15"),
    qualitative16: getCSSVariableValue("--color-qualitative-16"),
    qualitative17: getCSSVariableValue("--color-qualitative-17"),
    qualitative18: getCSSVariableValue("--color-qualitative-18"),
    qualitative19: getCSSVariableValue("--color-qualitative-19"),
    qualitative20: getCSSVariableValue("--color-qualitative-20"),
    qualitative21: getCSSVariableValue("--color-qualitative-21"),
    qualitative22: getCSSVariableValue("--color-qualitative-22"),
    qualitative23: getCSSVariableValue("--color-qualitative-23"),
    qualitative24: getCSSVariableValue("--color-qualitative-24"),
  };
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
