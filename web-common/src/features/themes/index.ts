/**
 * Theme system exports with data visualization palettes
 * 
 * This module provides:
 * - Theme color management (primary, secondary)
 * - Three data visualization palette types (sequential, diverging, qualitative)
 * - Reactive stores for accessing current theme colors
 * - Utility functions for color manipulation
 */

// Theme color actions
export {
  updateThemeVariables,
  createDarkVariation,
  setVariables,
} from "./actions";

// Palette color actions
export {
  setSequentialColor,
  setDivergingColor,
  setQualitativeColor,
  setPaletteColors,
  clearPaletteColor,
  clearAllPaletteColors,
  updatePaletteColorsFromTheme,
  type PaletteType,
} from "./actions";

// Palette color stores and utilities
export {
  // Sequential palette (9 colors for ordered data)
  sequentialColors,
  sequentialColorsArray,
  getSequentialColor,
  getSequentialColorsSnapshot,
  getSequentialColorsAsHex,
  
  // Diverging palette (11 colors for data with midpoint)
  divergingColors,
  divergingColorsArray,
  getDivergingColor,
  getDivergingColorsSnapshot,
  getDivergingColorsAsHex,
  
  // Qualitative palette (12 colors for categorical data)
  qualitativeColors,
  qualitativeColorsArray,
  getQualitativeColor,
  getQualitativeColorsSnapshot,
  getQualitativeColorsAsHex,
  
  // Utility
  colorToHex,
  
  // Types
  type SequentialColors,
  type DivergingColors,
  type QualitativeColors,
  type AllPaletteColors,
} from "./palette-store";

// Theme control (light/dark mode)
export { themeControl } from "./theme-control";

// Color configuration
export {
  TailwindColorSpacing,
  TailwindColors,
  defaultPrimaryColors,
  defaultSecondaryColors,
  type LightnessMap,
  type ThemeColorKind,
} from "./color-config";

// Theme selectors
export { useTheme } from "./selectors";

// Color definitions
export { allColors } from "./colors";
