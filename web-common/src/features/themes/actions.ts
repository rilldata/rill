/**
 * Theme Actions
 * 
 * Main functions for updating theme variables from the new theme structure
 * and legacy color-based themes.
 */

import { TailwindColorSpacing } from "./color-config.ts";
import type { V1ThemeSpec } from "../../../../web-common/src/runtime-client/index.ts";
import chroma, { type Color } from "chroma-js";
import { generateColorPalette } from "./palette-generator.ts";
import { featureFlags } from "../feature-flags.ts";
import { get } from "svelte/store";
import { generatePalette, DEFAULT_STEP_COUNT, DEFAULT_GAMMA, createDarkVariation as createDarkVariationFn } from "./color-generation.ts";
import { 
  sanitizeThemeVariables, 
  themeVariablesToCSS 
} from "./css-sanitizer.ts";
import { defaultPrimaryPalette, defaultSecondaryPalette } from "./colors.ts";

// Cache default dark palettes (computed once on module load)
const defaultPrimaryDarkPalette = createDarkVariationFn(defaultPrimaryPalette);
const defaultSecondaryDarkPalette = createDarkVariationFn(defaultSecondaryPalette);

// Constants
const CUSTOM_THEME_STYLE_ID = "rill-custom-theme";

// Type definitions
type ColorPalette = {
  light: Color[];
  dark: Color[];
};

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
      // Convert all colors to HSL for internal representation
      root.style.setProperty(
        `--color-${type}-${mode}-${TailwindColorSpacing[i]}`,
        color.css("hsl"),
      );
    });
  }
}

/**
 * Sets intermediate CSS variables that fall back to light/dark mode variants
 * This allows variables like --color-theme-600 to work in scoped contexts
 */
function setIntermediateVariables(
  root: HTMLElement,
  type: string,
): void {
  TailwindColorSpacing.forEach((spacing) => {
    // Set intermediate variable with fallback logic:
    // In light mode: uses light variant, in dark mode: uses dark variant
    root.style.setProperty(
      `--color-${type}-${spacing}`,
      `light-dark(var(--color-${type}-light-${spacing}), var(--color-${type}-dark-${spacing}))`,
    );
  });
}

/**
 * Updates theme variables based on the provided theme specification
 * Supports both new (light/dark.variables) and legacy (primaryColor/secondaryColor) formats
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

  // Priority 1: New structure - theme.light.variables / theme.dark.variables
  // Check if new theme structure is being used (has light or dark with variables, primary, or secondary)
  const hasNewStructure = Boolean(
    theme?.light?.variables || theme?.dark?.variables ||
    theme?.light?.primary || theme?.dark?.primary ||
    theme?.light?.secondary || theme?.dark?.secondary
  );

  if (hasNewStructure && theme) {
    injectThemeVariables(theme, root);
    // Also handle legacy primary/secondary within the new structure
    handleLegacyPrimarySecondaryInNewStructure(theme, root);
    return;
  }

  // If no theme or no modern properties, remove any existing custom CSS
  removeExistingCustomCSS();

  // Priority 2: Legacy color properties (primaryColor, secondaryColor)
  updateLegacyColors(theme, root, allowNewPalette);
}

/**
 * Injects theme variables from the new theme.light.variables / theme.dark.variables structure
 * This is the main handler for the new backend theme format
 */
function injectThemeVariables(
  theme: V1ThemeSpec,
  scopeElement: HTMLElement,
): void {
  removeExistingCustomCSS();

  // Sanitize light and dark variables
  const lightVariables = sanitizeThemeVariables(theme.light?.variables);
  const darkVariables = sanitizeThemeVariables(theme.dark?.variables);

  // Convert to CSS and inject
  const scopeSelector = scopeElement === document.documentElement 
    ? undefined 
    : ".dashboard-theme-boundary";
  
  const css = themeVariablesToCSS(lightVariables, darkVariables, scopeSelector);
  
  if (css) {
    createAndInjectStyle(css);
  }
}

/**
 * Handles legacy primary/secondary color properties within the new theme structure
 * (theme.light.primary, theme.light.secondary, theme.dark.primary, theme.dark.secondary)
 */
function handleLegacyPrimarySecondaryInNewStructure(
  theme: V1ThemeSpec,
  root: HTMLElement,
): void {
  const lightPrimary = theme.light?.primary;
  const darkPrimary = theme.dark?.primary;
  const lightSecondary = theme.light?.secondary;
  const darkSecondary = theme.dark?.secondary;

  // Handle primary colors
  if (lightPrimary || darkPrimary) {
    try {
      const palettes = generatePalettesFromColorStrings(
        lightPrimary,
        darkPrimary,
        defaultPrimaryPalette,
        defaultPrimaryDarkPalette
      );
      applyPaletteToVariables(root, "theme", palettes);
      applyPaletteToVariables(root, "primary", palettes);
    } catch (error) {
      console.error('Failed to generate palette from primary colors in new structure:', { lightPrimary, darkPrimary }, error);
    }
  }

  // Handle secondary colors
  if (lightSecondary || darkSecondary) {
    try {
      const palettes = generatePalettesFromColorStrings(
        lightSecondary,
        darkSecondary,
        defaultSecondaryPalette,
        defaultSecondaryDarkPalette
      );
      applyPaletteToVariables(root, "theme-secondary", palettes);
      applyPaletteToVariables(root, "secondary", palettes);
    } catch (error) {
      console.error('Failed to generate palette from secondary colors in new structure:', { lightSecondary, darkSecondary }, error);
    }
  }
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
    setIntermediateVariables(root, "theme");
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
    setIntermediateVariables(root, "theme-secondary");
  } else {
    setVariables(root, "theme-secondary", "light");
    setVariables(root, "theme-secondary", "dark");
  }
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
 * Generates palettes from light and dark color strings (for new theme structure)
 */
function generatePalettesFromColorStrings(
  lightColor: string | undefined,
  darkColor: string | undefined,
  defaultLightPalette: Color[],
  defaultDarkPalette: Color[],
): ColorPalette {
  let lightPalette: Color[];
  let darkPalette: Color[];
  
  // Generate palette from lightColor if provided
  let generatedPalette: { light: Color[]; dark: Color[] } | null = null;
  if (lightColor) {
    const color = chroma(lightColor);
    generatedPalette = generatePalette(color, false, DEFAULT_STEP_COUNT, DEFAULT_GAMMA);
    lightPalette = generatedPalette.light;
  } else {
    // Use default light palette when not defined
    lightPalette = defaultLightPalette;
  }
  
  // Handle dark palette
  if (darkColor) {
    const color = chroma(darkColor);
    const palette = generatePalette(color, false, DEFAULT_STEP_COUNT, DEFAULT_GAMMA);
    darkPalette = palette.dark;
  } else if (generatedPalette) {
    // If light was generated but dark wasn't specified, reuse the generated dark variant
    darkPalette = generatedPalette.dark;
  } else {
    // Use default dark palette when neither is defined
    darkPalette = defaultDarkPalette;
  }
  
  return { light: lightPalette, dark: darkPalette };
}

/**
 * Applies a palette to CSS variables
 */
function applyPaletteToVariables(
  root: HTMLElement,
  type: string,
  palettes: ColorPalette,
): void {
  setVariables(root, type, "light", palettes.light);
  setVariables(root, type, "dark", palettes.dark);
  // Set intermediate variables that automatically switch between light/dark
  setIntermediateVariables(root, type);
}
