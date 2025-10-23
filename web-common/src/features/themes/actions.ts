/**
 * Theme Actions
 *
 * Main functions for updating theme variables from the new theme structure
 * and legacy color-based themes.
 */

import { TailwindColorSpacing } from "./color-config.ts";
import type {
  V1ThemeSpec,
  V1ThemeColors,
} from "@rilldata/web-common/runtime-client";
import chroma, { type Color } from "chroma-js";
import { generateColorPalette } from "./palette-generator.ts";
import { featureFlags } from "../feature-flags.ts";
import { get } from "svelte/store";
import {
  generatePalette,
  DEFAULT_STEP_COUNT,
  DEFAULT_GAMMA,
} from "./color-generation.ts";
import { sanitizeThemeVariables } from "./css-sanitizer.ts";
import { clearCSSVariableCache } from "@rilldata/web-common/components/vega/vega-config";

const CUSTOM_THEME_STYLE_ID = "rill-custom-theme";

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
function setIntermediateVariables(root: HTMLElement, type: string): void {
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
  const root = scopeElement || document.documentElement;
  const { darkMode } = featureFlags;
  const allowNewPalette = get(darkMode);
  const isDarkMode = document.documentElement.classList.contains("dark");

  if (!theme) {
    clearThemeVariables(root);
    return;
  }

  const currentModeTheme = isDarkMode ? theme.dark : theme.light;
  const hasCurrentModeTheme = Boolean(
    currentModeTheme?.variables ||
      currentModeTheme?.primary ||
      currentModeTheme?.secondary,
  );

  clearThemeVariables(root);

  if (hasCurrentModeTheme && theme && currentModeTheme) {
    injectCurrentModeThemeVariables(currentModeTheme, root);
    handleCurrentModePrimarySecondary(currentModeTheme, root);
    return;
  }

  updatePrimaryColor(theme, root, allowNewPalette);
  updateSecondaryColor(theme, root, allowNewPalette);
}

function clearThemeVariables(root: HTMLElement): void {
  removeExistingCustomCSS();
  clearCSSVariableCache();

  if (root !== document.documentElement) {
    setVariables(root, "theme", "light");
    setVariables(root, "theme", "dark");
    setVariables(root, "primary", "light");
    setVariables(root, "primary", "dark");
    setVariables(root, "theme-secondary", "light");
    setVariables(root, "theme-secondary", "dark");
    setVariables(root, "secondary", "light");
    setVariables(root, "secondary", "dark");
  }
}

function injectCurrentModeThemeVariables(
  currentModeTheme: V1ThemeColors,
  scopeElement: HTMLElement,
): void {
  const vars = currentModeTheme.variables;
  if (!vars || typeof vars !== "object") return;

  const variables = sanitizeThemeVariables(vars);
  if (Object.keys(variables).length === 0) return;

  const scopeSelector =
    scopeElement === document.documentElement
      ? undefined
      : ".dashboard-theme-boundary";

  let css = "";
  const selector = scopeSelector || ":root";
  css += `${selector} {\n`;
  for (const [name, value] of Object.entries(variables)) {
    css += `  ${name}: ${value};\n`;
  }
  css += "}\n";

  createAndInjectStyle(css);
}

function handleCurrentModePrimarySecondary(
  currentModeTheme: V1ThemeColors,
  root: HTMLElement,
): void {
  const isDarkMode = document.documentElement.classList.contains("dark");
  const mode = isDarkMode ? "dark" : "light";

  const primaryColor = currentModeTheme.primary;
  if (primaryColor && typeof primaryColor === "string") {
    try {
      const palette = generatePalette(
        chroma(primaryColor),
        false,
        DEFAULT_STEP_COUNT,
        DEFAULT_GAMMA,
      );
      const colors = isDarkMode ? palette.dark : palette.light;

      setVariables(root, "theme", mode, colors);
      setVariables(root, "primary", mode, colors);
      setIntermediateVariables(root, "theme");
      setIntermediateVariables(root, "primary");
    } catch (error) {
      console.error("Failed to generate palette from primary color:", error);
    }
  }

  const secondaryColor = currentModeTheme.secondary;
  if (secondaryColor && typeof secondaryColor === "string") {
    try {
      const palette = generatePalette(
        chroma(secondaryColor),
        false,
        DEFAULT_STEP_COUNT,
        DEFAULT_GAMMA,
      );
      const colors = isDarkMode ? palette.dark : palette.light;

      setVariables(root, "theme-secondary", mode, colors);
      setVariables(root, "secondary", mode, colors);
      setIntermediateVariables(root, "theme-secondary");
      setIntermediateVariables(root, "secondary");
    } catch (error) {
      console.error("Failed to generate palette from secondary color:", error);
    }
  }
}

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
  const style = document.createElement("style");
  style.id = CUSTOM_THEME_STYLE_ID;
  style.textContent = css;
  document.head.appendChild(style);
}
