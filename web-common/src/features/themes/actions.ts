/**
 * Theme Actions
 * 
 * Main functions for updating theme variables, including CSS-based themes
 * and legacy color-based themes.
 */

import { TailwindColorSpacing } from "./color-config.ts";
import type { V1ThemeSpec } from "../../../../web-common/src/runtime-client/index.ts";
import chroma, { type Color } from "chroma-js";
import { generateColorPalette } from "./palette-generator.ts";
import { featureFlags } from "../feature-flags.ts";
import { get } from "svelte/store";
import { generatePalette, DEFAULT_STEP_COUNT, DEFAULT_GAMMA } from "./color-generation.ts";
import { sanitizeAndExtractSafeVariables, extractColorVariables } from "./css-sanitizer.ts";
import { setPaletteColors, clearAllPaletteColors } from "./palette-colors.ts";

// Re-export from color-generation for backward compatibility
export { createDarkVariation, BLACK, WHITE, MODE } from "./color-generation.ts";

// Re-export palette functions
export {
  setSequentialColor,
  setDivergingColor,
  setQualitativeColor,
  setPaletteColors,
  clearPaletteColor,
  clearAllPaletteColors,
  type PaletteType,
} from "./palette-colors";

// Constants
const CUSTOM_THEME_STYLE_ID = "rill-custom-theme";

// Type definitions
type ColorMatch = {
  lightColor: string | null;
  darkColor: string | null;
};

type ColorPalette = {
  light: Color[];
  dark: Color[];
};

type ColorType = string;

/**
 * Sets CSS variables for a color type (theme, theme-secondary, etc.)
 */
export function setVariables(
  root: HTMLElement,
  type: string,
  mode: "dark" | "light",
  colors?: Color[],
): void {
  if (!colors) {
    // Only remove properties if we're working with a scoped element (not document root)
    // This prevents removing default theme colors from the global scope
    if (root !== document.documentElement) {
      TailwindColorSpacing.forEach((_, i) => {
        root.style.removeProperty(
          `--color-${type}-${mode}-${TailwindColorSpacing[i]}`,
        );
      });
    }
  } else {
    colors.forEach((color, i) => {
      root.style.setProperty(
        `--color-${type}-${mode}-${TailwindColorSpacing[i]}`,
        color.css("oklch"),
      );
    });
  }
}

/**
 * Updates theme variables based on the provided theme specification
 * @param theme - The theme specification to apply
 * @param scopeElement - Optional element to scope the theme to (defaults to document root)
 */
export function updateThemeVariables(
  theme: V1ThemeSpec | undefined,
  scopeElement?: HTMLElement | null,
): void {
  // Use provided scope element or fall back to document root
  const root = scopeElement || document.documentElement;
  const { darkMode } = featureFlags;
  const allowNewPalette = get(darkMode);

  // Handle new CSS property first (takes precedence over legacy colors)
  if (theme?.css) {
    injectCustomCSS(theme.css, root);
    return; // CSS themes override legacy color themes
  }

  // If no theme or no CSS, remove any existing custom CSS
  removeExistingCustomCSS();

  // Handle legacy color properties (this clears them if theme is undefined)
  updateLegacyColors(theme, root, allowNewPalette);
  
  // Update palette colors if provided (this clears them if theme is undefined)
  updatePaletteColorsFromTheme(theme, root);
}

/**
 * Updates legacy color properties (primary and secondary)
 */
function updateLegacyColors(
  theme: V1ThemeSpec | undefined,
  root: HTMLElement,
  allowNewPalette: boolean,
): void {
  updatePrimaryColor(theme, root, allowNewPalette);
  updateSecondaryColor(theme, root, allowNewPalette);
}

/**
 * Updates primary color variables
 */
function updatePrimaryColor(
  theme: V1ThemeSpec | undefined,
  root: HTMLElement,
  allowNewPalette: boolean,
): void {
  if (theme?.primaryColor) {
    const chromaColor = chroma.rgb(
      (theme.primaryColor.red ?? 1) * 256,
      (theme.primaryColor.green ?? 1) * 256,
      (theme.primaryColor.blue ?? 1) * 256,
    );

    const originalLightPalette = generateColorPalette(chromaColor);
    const { light, dark } = generatePalette(chromaColor, false);

    setVariables(
      root,
      "theme",
      "light",
      allowNewPalette ? light : originalLightPalette,
    );

    setVariables(root, "theme", "dark", dark);
  } else {
    setVariables(root, "theme", "light");
    setVariables(root, "theme", "dark");
  }
}

/**
 * Updates secondary color variables
 */
function updateSecondaryColor(
  theme: V1ThemeSpec | undefined,
  root: HTMLElement,
  allowNewPalette: boolean,
): void {
  if (theme?.secondaryColor) {
    const chromaColor = chroma.rgb(
      (theme.secondaryColor.red ?? 1) * 256,
      (theme.secondaryColor.green ?? 1) * 256,
      (theme.secondaryColor.blue ?? 1) * 256,
    );

    const originalLightPalette = generateColorPalette(chromaColor);
    const { light, dark } = generatePalette(chromaColor, false);

    setVariables(
      root,
      "theme-secondary",
      "light",
      allowNewPalette ? light : originalLightPalette,
    );
    setVariables(root, "theme-secondary", "dark", dark);
  } else {
    setVariables(root, "theme-secondary", "light");
    setVariables(root, "theme-secondary", "dark");
  }
}

/**
 * Injects custom CSS from theme definition
 * Scopes CSS to .dashboard-theme-boundary to prevent affecting global Rill chrome
 * Only extracts and injects known safe CSS variables to prevent XSS attacks
 */
function injectCustomCSS(css: string, scopeElement: HTMLElement): void {
  removeExistingCustomCSS();
  
  // Sanitize CSS and scope it to the dashboard boundary
  const sanitizedCSS = sanitizeAndExtractSafeVariables(css, ".dashboard-theme-boundary");
  
  // Only inject if we have safe variables
  if (sanitizedCSS) {
    createAndInjectStyle(sanitizedCSS);
  }
  
  // Apply primary/secondary color variables to legacy theme system
  applySimpleColorVariables(css, scopeElement);
}

/**
 * Removes any existing custom theme CSS
 */
function removeExistingCustomCSS(): void {
  const existingStyle = document.getElementById(CUSTOM_THEME_STYLE_ID);
  if (existingStyle) {
    existingStyle.remove();
  }
}

/**
 * Creates and injects new style element
 */
function createAndInjectStyle(css: string): void {
  const style = document.createElement('style');
  style.id = CUSTOM_THEME_STYLE_ID;
  style.textContent = css;
  document.head.appendChild(style);
}

/**
 * Applies simple color variables (--primary, --secondary) to the existing theme system
 */
function applySimpleColorVariables(css: string, scopeElement: HTMLElement): void {
  const colorMatches = extractColorVariables(css);
  
  applyPrimaryColorVariables(colorMatches, scopeElement);
  applySecondaryColorVariables(colorMatches, scopeElement);
}

/**
 * Applies primary color variables to the theme system
 */
function applyPrimaryColorVariables(
  colorMatches: { primary: ColorMatch; secondary: ColorMatch },
  root: HTMLElement,
): void {
  const { lightColor, darkColor } = colorMatches.primary;
  
  if (!lightColor && !darkColor) return;
  
  try {
    const palettes = generatePalettesFromColors(lightColor, darkColor);
    
    if (palettes) {
      applyPaletteToVariables(root, "theme", palettes);
      applyPaletteToVariables(root, "primary", palettes);
    }
  } catch (error) {
    console.error('Failed to generate palette from primary colors:', { lightColor, darkColor }, error);
  }
}

/**
 * Applies secondary color variables to the theme system
 */
function applySecondaryColorVariables(
  colorMatches: { primary: ColorMatch; secondary: ColorMatch },
  root: HTMLElement,
): void {
  const { lightColor, darkColor } = colorMatches.secondary;
  
  if (!lightColor && !darkColor) return;
  
  try {
    const palettes = generatePalettesFromColors(lightColor, darkColor);
    
    if (palettes) {
      applyPaletteToVariables(root, "theme-secondary", palettes);
      applyPaletteToVariables(root, "secondary", palettes);
    }
  } catch (error) {
    console.error('Failed to generate palette from secondary colors:', { lightColor, darkColor }, error);
  }
}

/**
 * Generates palettes from light and dark colors
 */
function generatePalettesFromColors(
  lightColor: string | null,
  darkColor: string | null,
): ColorPalette | null {
  let lightPalette: ColorPalette | undefined;
  let darkPalette: ColorPalette | undefined;
  
  if (lightColor) {
    const color = chroma(lightColor);
    lightPalette = generatePalette(color, false, DEFAULT_STEP_COUNT, DEFAULT_GAMMA);
  }
  
  if (darkColor) {
    const color = chroma(darkColor);
    darkPalette = generatePalette(color, false, DEFAULT_STEP_COUNT, DEFAULT_GAMMA);
  }
  
  // Use the same palette for both if only one color is provided
  if (lightColor && !darkColor) {
    darkPalette = lightPalette;
  } else if (darkColor && !lightColor) {
    lightPalette = darkPalette;
  }
  
  return lightPalette && darkPalette ? { light: lightPalette.light, dark: darkPalette.dark } : null;
}

/**
 * Applies a palette to CSS variables
 */
function applyPaletteToVariables(
  root: HTMLElement,
  type: ColorType,
  palettes: ColorPalette,
): void {
  setVariables(root, type, "light", palettes.light);
  setVariables(root, type, "dark", palettes.dark);
}

/**
 * Updates palette colors from a theme specification
 */
export function updatePaletteColorsFromTheme(
  theme: V1ThemeSpec | undefined,
  scopeElement?: HTMLElement,
): void {
  const themeWithPalettes = theme as V1ThemeSpec & {
    sequentialColors?: Array<{ red?: number; green?: number; blue?: number }>;
    divergingColors?: Array<{ red?: number; green?: number; blue?: number }>;
    qualitativeColors?: Array<{ red?: number; green?: number; blue?: number }>;
  };

  // Clear all if no palette colors defined
  if (!themeWithPalettes?.sequentialColors && 
      !themeWithPalettes?.divergingColors && 
      !themeWithPalettes?.qualitativeColors) {
    clearAllPaletteColors(undefined, scopeElement);
    return;
  }

  try {
    if (themeWithPalettes.sequentialColors) {
      const colors = themeWithPalettes.sequentialColors.map(convertThemeColorToChroma);
      setPaletteColors("sequential", colors, scopeElement);
    }

    if (themeWithPalettes.divergingColors) {
      const colors = themeWithPalettes.divergingColors.map(convertThemeColorToChroma);
      setPaletteColors("diverging", colors, scopeElement);
    }

    if (themeWithPalettes.qualitativeColors) {
      const colors = themeWithPalettes.qualitativeColors.map(convertThemeColorToChroma);
      setPaletteColors("qualitative", colors, scopeElement);
    }
  } catch (error) {
    console.error('Failed to set palette colors from theme:', error);
    clearAllPaletteColors(undefined, scopeElement);
  }
}

/**
 * Helper function to convert theme color to chroma color
 */
function convertThemeColorToChroma(color: { red?: number; green?: number; blue?: number }): Color {
  return chroma.rgb(
    (color.red ?? 0) * 256,
    (color.green ?? 0) * 256,
    (color.blue ?? 0) * 256,
  );
}
