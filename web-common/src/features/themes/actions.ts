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
import { generatePalette } from "./color-generation.ts";
import { themeManager } from "./theme-manager";

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
    themeManager.clearTheme(root);
    return;
  }

  const currentModeTheme: V1ThemeColors | undefined = isDarkMode
    ? theme.dark
    : theme.light;

  themeManager.clearTheme(root);
  themeManager.applyTheme(theme, currentModeTheme, root);

  const hasCurrentModeTheme = Boolean(
    currentModeTheme?.variables ||
      currentModeTheme?.primary ||
      currentModeTheme?.secondary,
  );

  if (!hasCurrentModeTheme) {
    updatePrimaryColor(theme, root, allowNewPalette);
    updateSecondaryColor(theme, root, allowNewPalette);
    themeManager.clearCSSVariableCache();
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
