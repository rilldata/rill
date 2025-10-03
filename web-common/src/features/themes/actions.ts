import { TailwindColorSpacing } from "./color-config.ts";
import type { V1ThemeSpec } from "../../../../web-common/src/runtime-client/index.ts";
import chroma, { type Color } from "chroma-js";
import { clamp } from "../../../../web-common/src/lib/clamp.ts";
import { generateColorPalette } from "./palette-generator.ts";
import { featureFlags } from "../feature-flags.ts";
import { get } from "svelte/store";

// Constants
export const BLACK = chroma("black");
export const WHITE = chroma("white");
export const MODE = "oklab";

// Color generation constants
const DEFAULT_STEP_COUNT = 11;
const DEFAULT_GAMMA = 1.12;
const GRAY_DESATURATION_THRESHOLD = 0.05;
const SPECTRUM_COLORS = 102;
const DARK_GAMMA = 1.15;
const LIGHT_GAMMA = 0.6;
const DARK_LUMINANCE_THRESHOLD = 0.25;
const LIGHT_LUMINANCE_THRESHOLD = 0.965;
const CONTRAST_THRESHOLDS = {
  MIN_CONTRAST: 1.06,
  MAX_CONTRAST: 18.5,
  MID_CONTRAST: 5,
  DARK_700_CONTRAST: 10.8,
  DARK_300_CONTRAST: 2,
  LIGHT_CONTRAST: 1.065,
  DARK_LIGHT_CONTRAST: 1.2,
} as const;

// CSS injection constants
const CUSTOM_THEME_STYLE_ID = 'rill-custom-theme';
const COLOR_SCALE_FACTOR = 256;

// Type definitions
type ColorMode = "dark" | "light";
type ColorType = "theme" | "theme-secondary" | "primary" | "secondary";

interface ColorPalette {
  light: Color[];
  dark: Color[];
}

interface ColorMatch {
  lightColor: string | null;
  darkColor: string | null;
}

/**
 * Creates a dark variation of colors from the middle shade (index 5)
 */
export function createDarkVariation(colors: Color[]): Color[] {
  return generatePalette(colors[5]).dark;
}

/**
 * Generates a color palette with light and dark variations
 */
export function generatePalette(
  refColor: Color,
  desaturateNearGray: boolean = true,
  stepCount: number = DEFAULT_STEP_COUNT,
  gamma: number = DEFAULT_GAMMA,
): ColorPalette {
  const oklchValues = refColor.oklch();
  const c = oklchValues[1];

  if (desaturateNearGray && c < GRAY_DESATURATION_THRESHOLD) {
    return generateGrayPalette(refColor, stepCount);
  }

  return generateColorfulPalette(refColor, stepCount, gamma);
}

/**
 * Generates a palette for near-gray colors
 */
function generateGrayPalette(refColor: Color, stepCount: number): ColorPalette {
  const spectrum = chroma
    .scale([chroma("black"), refColor, chroma("white")])
    .mode(MODE)
    .gamma(1)
    .colors(SPECTRUM_COLORS, null);

  const darkestDark = findContrast(spectrum, BLACK, CONTRAST_THRESHOLDS.MIN_CONTRAST) ?? spectrum[0];
  const lightestDark = findContrast(spectrum, darkestDark, CONTRAST_THRESHOLDS.MAX_CONTRAST) ?? spectrum[0];
  const middleValue = findContrast(spectrum, darkestDark, CONTRAST_THRESHOLDS.MID_CONTRAST) ?? spectrum[50];

  return {
    dark: [
      ...chroma
        .scale([darkestDark, middleValue])
        .mode(MODE)
        .gamma(DARK_GAMMA)
        .colors(6, null)
        .slice(0, -1),
      middleValue,
      ...chroma
        .scale([middleValue, lightestDark])
        .mode(MODE)
        .gamma(LIGHT_GAMMA)
        .colors(6, null)
        .slice(1),
    ],
    light: chroma
      .scale([chroma("white"), chroma("black")])
      .mode(MODE)
      .gamma(1)
      .colors(stepCount + 2, null)
      .slice(1, -1),
  };
}

/**
 * Generates a palette for colorful (non-gray) colors
 */
function generateColorfulPalette(refColor: Color, stepCount: number, gamma: number): ColorPalette {
  const oklchValues = refColor.oklch();
  const l = oklchValues[0];
  const c = oklchValues[1];
  const h = oklchValues[2];
  
  const darkRef = chroma.oklch(clamp(0.6, l, 1), c, h);
  const lightRef = chroma.oklch(clamp(0, l, 0.82), c, h);

  const darkSpectrum = chroma
    .scale([chroma("black"), darkRef, chroma("white")])
    .mode(MODE)
    .gamma(1)
    .colors(SPECTRUM_COLORS, null);

  const lightSpectrum = chroma
    .scale([chroma("black"), lightRef, chroma("white")])
    .mode(MODE)
    .gamma(1)
    .colors(SPECTRUM_COLORS, null);

  const reversedDarkSpectrum = [...darkSpectrum].reverse();
  const reversedLightSpectrum = [...lightSpectrum].reverse();

  const darkestDark = findLuminance(darkSpectrum, DARK_LUMINANCE_THRESHOLD) ?? darkSpectrum[0];
  const find700 = findContrast(darkSpectrum, darkestDark, CONTRAST_THRESHOLDS.DARK_700_CONTRAST, 0.088, 0.073) ?? darkSpectrum[0];
  const find300 = findContrast(darkSpectrum, darkestDark, CONTRAST_THRESHOLDS.DARK_300_CONTRAST) ?? darkSpectrum[0];
  const lightestDark = findLuminance(darkSpectrum, LIGHT_LUMINANCE_THRESHOLD) ?? reversedDarkSpectrum[0];
  const lightestLight = findContrast(reversedLightSpectrum, WHITE, CONTRAST_THRESHOLDS.LIGHT_CONTRAST) ?? reversedLightSpectrum[0];
  const darkestLight = findContrast(lightSpectrum, BLACK, CONTRAST_THRESHOLDS.DARK_LIGHT_CONTRAST) ?? lightSpectrum[0];

  return {
    dark: chroma
      .scale([darkestDark, find300, darkRef, find700, lightestDark])
      .mode(MODE)
      .gamma(gamma)
      .colors(stepCount, null),
    light: chroma
      .scale([lightestLight, lightRef, darkestLight])
      .mode(MODE)
      .gamma(gamma)
      .colors(stepCount, null),
  };
}

/**
 * Sets CSS custom properties for color variables
 */
export function setVariables(
  root: HTMLElement,
  type: ColorType,
  mode: ColorMode,
  colors?: Color[],
): void {
  if (!colors) {
    TailwindColorSpacing.forEach((_, i) => {
      root.style.removeProperty(
        `--color-${type}-${mode}-${TailwindColorSpacing[i]}`,
      );
    });
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
 */
export function updateThemeVariables(theme: V1ThemeSpec | undefined): void {
  const root = document.documentElement;
  const { darkMode } = featureFlags;
  const allowNewPalette = get(darkMode);

  // Handle new CSS property first (takes precedence over legacy colors)
  if (theme?.css) {
    injectCustomCSS(theme.css);
    return; // CSS themes override legacy color themes
  }

  // Handle legacy color properties
  updateLegacyColors(theme, root, allowNewPalette);
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
    const chromaColor = convertThemeColorToChroma(theme.primaryColor);
    const originalLightPalette = generateColorPalette(chromaColor);
    const { light, dark } = generatePalette(chromaColor, false);

    setVariables(root, "theme", "light", allowNewPalette ? light : originalLightPalette);
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
    const chromaColor = convertThemeColorToChroma(theme.secondaryColor);
    const originalLightPalette = generateColorPalette(chromaColor);
    const { light, dark } = generatePalette(chromaColor, false);

    setVariables(root, "theme-secondary", "light", allowNewPalette ? light : originalLightPalette);
    setVariables(root, "theme-secondary", "dark", dark);
  } else {
    setVariables(root, "theme-secondary", "light");
    setVariables(root, "theme-secondary", "dark");
  }
}

/**
 * Converts theme color specification to chroma color
 */
function convertThemeColorToChroma(colorSpec: { red?: number; green?: number; blue?: number }): Color {
  return chroma.rgb(
    (colorSpec.red ?? 1) * COLOR_SCALE_FACTOR,
    (colorSpec.green ?? 1) * COLOR_SCALE_FACTOR,
    (colorSpec.blue ?? 1) * COLOR_SCALE_FACTOR,
  );
}

/**
 * Finds a color in the spectrum with the specified luminance
 */
function findLuminance(spectrum: Color[], luminance: number): Color | undefined {
  return spectrum.find((color) => {
    const [l] = color.oklch();
    return l >= luminance;
  });
}

/**
 * Finds a color in the spectrum with the specified contrast ratio
 */
function findContrast(
  spectrum: Color[],
  comparedTo: Color,
  contrast: number,
  maxSaturation?: number,
  minSaturation?: number,
): Color | undefined {
  const color = spectrum.find((color) => {
    return chroma.contrast(comparedTo, color) >= contrast;
  });

  if (!color) return undefined;

  const saturation = color.oklch()[1];

  if (maxSaturation && saturation > maxSaturation) {
    return color.set("oklch.c", maxSaturation);
  } else if (minSaturation && saturation < minSaturation) {
    return color.set("oklch.c", minSaturation);
  }

  return color;
}

/**
 * Injects custom CSS from theme definition
 */
function injectCustomCSS(css: string): void {
  removeExistingCustomCSS();
  createAndInjectStyle(css);
  applySimpleColorVariables(css);
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
function applySimpleColorVariables(css: string): void {
  const root = document.documentElement;
  const { darkMode } = featureFlags;
  const allowNewPalette = get(darkMode);
  
  const colorMatches = extractColorVariables(css);
  
  applyPrimaryColorVariables(colorMatches, root, allowNewPalette);
  applySecondaryColorVariables(colorMatches, root, allowNewPalette);
}

/**
 * Extracts color variables from CSS
 */
function extractColorVariables(css: string): {
  primary: ColorMatch;
  secondary: ColorMatch;
} {
  const lightPrimaryMatch = css.match(/:root\s*\{[^}]*--primary:\s*([^;]+);/);
  const darkPrimaryMatch = css.match(/:root\.dark\s*\{[^}]*--primary:\s*([^;]+);/);
  const lightSecondaryMatch = css.match(/:root\s*\{[^}]*--secondary:\s*([^;]+);/);
  const darkSecondaryMatch = css.match(/:root\.dark\s*\{[^}]*--secondary:\s*([^;]+);/);
  
  return {
    primary: {
      lightColor: lightPrimaryMatch ? lightPrimaryMatch[1].trim() : null,
      darkColor: darkPrimaryMatch ? darkPrimaryMatch[1].trim() : null,
    },
    secondary: {
      lightColor: lightSecondaryMatch ? lightSecondaryMatch[1].trim() : null,
      darkColor: darkSecondaryMatch ? darkSecondaryMatch[1].trim() : null,
    },
  };
}

/**
 * Applies primary color variables to the theme system
 */
function applyPrimaryColorVariables(
  colorMatches: { primary: ColorMatch; secondary: ColorMatch },
  root: HTMLElement,
  allowNewPalette: boolean,
): void {
  const { lightColor, darkColor } = colorMatches.primary;
  
  if (!lightColor && !darkColor) return;
  
  try {
    const palettes = generatePalettesFromColors(lightColor, darkColor);
    
    if (palettes) {
      applyPaletteToVariables(root, "theme", palettes, allowNewPalette);
      applyPaletteToVariables(root, "primary", palettes, allowNewPalette);
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
  allowNewPalette: boolean,
): void {
  const { lightColor, darkColor } = colorMatches.secondary;
  
  if (!lightColor && !darkColor) return;
  
  try {
    const palettes = generatePalettesFromColors(lightColor, darkColor);
    
    if (palettes) {
      applyPaletteToVariables(root, "theme-secondary", palettes, allowNewPalette);
      applyPaletteToVariables(root, "secondary", palettes, allowNewPalette);
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
  allowNewPalette: boolean,
): void {
  setVariables(root, type, "light", allowNewPalette ? palettes.light : palettes.light);
  setVariables(root, type, "dark", palettes.dark);
}