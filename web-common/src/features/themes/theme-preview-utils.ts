import { parseDocument } from "yaml";
import { writable } from "svelte/store";
import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
import { Theme } from "./theme";

/**
 * Shared store for theme preview mode across editor and inspector
 */
export const themePreviewMode = writable<PreviewMode>("light");

/**
 * Number of colors in each palette
 */
export const SEQUENTIAL_COLOR_COUNT = 9;
export const DIVERGING_COLOR_COUNT = 11;
export const QUALITATIVE_COLOR_COUNT = 24;

/**
 * CSS variable fallback for theme background colors
 */
export const BACKGROUND_FALLBACK_LIGHT = "var(--color-gray-50, #f9fafb)";
export const BACKGROUND_FALLBACK_DARK = "var(--color-gray-900, #111827)";
export const CARD_FALLBACK_LIGHT = "var(--color-white, #ffffff)";
export const CARD_FALLBACK_DARK = "var(--color-gray-700, #374151)";

export type PreviewMode = "light" | "dark";

export interface ThemeColors {
  sequentialColors: string[];
  divergingColors: string[];
  qualitativeColors: string[];
  primaryColor: string;
  backgroundColor: string;
  cardColor: string;
}

/**
 * Parses YAML content and returns a Theme instance
 */
export function parseThemeFromYaml(yamlContent: string | null): {
  theme: Theme;
  themeData: V1ThemeSpec;
} {
  const parsedDocument = yamlContent ? parseDocument(yamlContent) : null;
  const themeData = (parsedDocument?.toJSON() || {}) as V1ThemeSpec;
  const theme = themeData ? new Theme(themeData) : new Theme(undefined);
  return { theme, themeData };
}

/**
 * Extracts color palette arrays from theme YAML data
 */
export function extractThemeColors(
  themeData: V1ThemeSpec,
  previewMode: PreviewMode,
): ThemeColors {
  const lightModeYaml = (themeData?.light || {}) as Record<string, string>;
  const darkModeYaml = (themeData?.dark || {}) as Record<string, string>;
  const currentModeYaml =
    previewMode === "light" ? lightModeYaml : darkModeYaml;

  // Extract palette colors into arrays
  const sequentialColors = Array.from(
    { length: SEQUENTIAL_COLOR_COUNT },
    (_, i) => {
      const key = `color-sequential-${i + 1}`;
      return currentModeYaml[key] || `var(--color-sequential-${i + 1})`;
    },
  );

  const divergingColors = Array.from(
    { length: DIVERGING_COLOR_COUNT },
    (_, i) => {
      const key = `color-diverging-${i + 1}`;
      return currentModeYaml[key] || `var(--color-diverging-${i + 1})`;
    },
  );

  const qualitativeColors = Array.from(
    { length: QUALITATIVE_COLOR_COUNT },
    (_, i) => {
      const key = `color-qualitative-${i + 1}`;
      return currentModeYaml[key] || `var(--color-qualitative-${i + 1})`;
    },
  );

  // Extract theme colors with CSS variable fallbacks
  const primaryColor = currentModeYaml["primary"] || "var(--color-theme-500)";
  const backgroundColor =
    currentModeYaml["background"] ||
    (previewMode === "light"
      ? BACKGROUND_FALLBACK_LIGHT
      : BACKGROUND_FALLBACK_DARK);
  const cardColor =
    currentModeYaml["card"] ||
    (previewMode === "light" ? CARD_FALLBACK_LIGHT : CARD_FALLBACK_DARK);

  return {
    sequentialColors,
    divergingColors,
    qualitativeColors,
    primaryColor,
    backgroundColor,
    cardColor,
  };
}
