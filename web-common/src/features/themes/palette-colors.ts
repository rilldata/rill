/**
 * Palette Color Management
 *
 * Functions for setting and clearing palette colors (sequential, diverging, qualitative)
 * for data visualization.
 */

import { type Color } from "chroma-js";
import {
  generatePalette,
  DEFAULT_STEP_COUNT,
  DEFAULT_GAMMA,
} from "./color-generation";
import { getChroma } from "./theme-utils";

/**
 * Palette type definitions for data visualization
 */
export type PaletteType = "sequential" | "diverging" | "qualitative";

/**
 * Sets a single sequential color (1-9)
 * Sequential colors are for ordered data that progresses from low to high
 */
export function setSequentialColor(
  index: number,
  color: string | Color,
  scopeElement?: HTMLElement,
): void {
  if (index < 1 || index > 9) {
    throw new Error("Sequential color index must be between 1 and 9");
  }

  const root = scopeElement || document.documentElement;
  const chromaColor = typeof color === "string" ? getChroma(color) : color;
  const { light, dark } = generatePalette(
    chromaColor,
    false,
    DEFAULT_STEP_COUNT,
    DEFAULT_GAMMA,
  );

  // Convert all colors to HSL for internal representation
  root.style.setProperty(
    `--color-sequential-light-${index}`,
    light[5].css("hsl"),
  );
  root.style.setProperty(
    `--color-sequential-dark-${index}`,
    dark[5].css("hsl"),
  );
}

/**
 * Sets a single diverging color (1-11)
 * Diverging colors emphasize mid-range values and extremes in different hues
 */
export function setDivergingColor(
  index: number,
  color: string | Color,
  scopeElement?: HTMLElement,
): void {
  if (index < 1 || index > 11) {
    throw new Error("Diverging color index must be between 1 and 11");
  }

  const root = scopeElement || document.documentElement;
  const chromaColor = typeof color === "string" ? getChroma(color) : color;
  const { light, dark } = generatePalette(
    chromaColor,
    false,
    DEFAULT_STEP_COUNT,
    DEFAULT_GAMMA,
  );

  // Convert all colors to HSL for internal representation
  root.style.setProperty(
    `--color-diverging-light-${index}`,
    light[5].css("hsl"),
  );
  root.style.setProperty(`--color-diverging-dark-${index}`, dark[5].css("hsl"));
}

/**
 * Sets a single qualitative color (1-24)
 * Qualitative colors are for categorical data without inherent ordering
 */
export function setQualitativeColor(
  index: number,
  color: string | Color,
  scopeElement?: HTMLElement,
): void {
  if (index < 1 || index > 24) {
    throw new Error("Qualitative color index must be between 1 and 24");
  }

  const root = scopeElement || document.documentElement;
  const chromaColor = typeof color === "string" ? getChroma(color) : color;
  const { light, dark } = generatePalette(
    chromaColor,
    false,
    DEFAULT_STEP_COUNT,
    DEFAULT_GAMMA,
  );

  // Convert all colors to HSL for internal representation
  root.style.setProperty(
    `--color-qualitative-light-${index}`,
    light[5].css("hsl"),
  );
  root.style.setProperty(
    `--color-qualitative-dark-${index}`,
    dark[5].css("hsl"),
  );
}

/**
 * Sets multiple colors for a specific palette type
 */
export function setPaletteColors(
  type: PaletteType,
  colors: (string | Color)[],
  scopeElement?: HTMLElement,
): void {
  if (type === "sequential") {
    if (colors.length > 9)
      throw new Error("Maximum of 9 sequential colors allowed");
    colors.forEach((color, index) =>
      setSequentialColor(index + 1, color, scopeElement),
    );
  } else if (type === "diverging") {
    if (colors.length > 11)
      throw new Error("Maximum of 11 diverging colors allowed");
    colors.forEach((color, index) =>
      setDivergingColor(index + 1, color, scopeElement),
    );
  } else {
    if (colors.length > 24)
      throw new Error("Maximum of 24 qualitative colors allowed");
    colors.forEach((color, index) =>
      setQualitativeColor(index + 1, color, scopeElement),
    );
  }
}

/**
 * Clears a specific color from a palette type
 */
export function clearPaletteColor(
  type: PaletteType,
  index: number,
  scopeElement?: HTMLElement,
): void {
  const root = scopeElement || document.documentElement;

  // Only remove properties if we're working with a scoped element (not document root)
  if (root === document.documentElement) return;

  if (type === "sequential") {
    if (index < 1 || index > 9)
      throw new Error("Sequential index must be between 1 and 9");
    root.style.removeProperty(`--color-sequential-light-${index}`);
    root.style.removeProperty(`--color-sequential-dark-${index}`);
  } else if (type === "diverging") {
    if (index < 1 || index > 11)
      throw new Error("Diverging index must be between 1 and 11");
    root.style.removeProperty(`--color-diverging-light-${index}`);
    root.style.removeProperty(`--color-diverging-dark-${index}`);
  } else {
    if (index < 1 || index > 24)
      throw new Error("Qualitative index must be between 1 and 24");
    root.style.removeProperty(`--color-qualitative-light-${index}`);
    root.style.removeProperty(`--color-qualitative-dark-${index}`);
  }
}

/**
 * Clears all colors for a specific palette type
 */
export function clearAllPaletteColors(
  type?: PaletteType,
  scopeElement?: HTMLElement,
): void {
  const root = scopeElement || document.documentElement;

  // Only remove properties if we're working with a scoped element (not document root)
  if (root === document.documentElement) return;

  if (!type || type === "sequential") {
    for (let i = 1; i <= 9; i++) {
      root.style.removeProperty(`--color-sequential-light-${i}`);
      root.style.removeProperty(`--color-sequential-dark-${i}`);
    }
  }

  if (!type || type === "diverging") {
    for (let i = 1; i <= 11; i++) {
      root.style.removeProperty(`--color-diverging-light-${i}`);
      root.style.removeProperty(`--color-diverging-dark-${i}`);
    }
  }

  if (!type || type === "qualitative") {
    for (let i = 1; i <= 24; i++) {
      root.style.removeProperty(`--color-qualitative-light-${i}`);
      root.style.removeProperty(`--color-qualitative-dark-${i}`);
    }
  }
}
