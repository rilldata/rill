import { parseDocument, type Document, YAMLMap } from "yaml";
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
 * CSS variable fallbacks for theme colors
 */
export const BACKGROUND_FALLBACK_LIGHT = "var(--color-gray-50, #f9fafb)";
export const BACKGROUND_FALLBACK_DARK = "var(--color-gray-900, #111827)";
export const CARD_FALLBACK_LIGHT = "var(--color-white, #ffffff)";
export const CARD_FALLBACK_DARK = "var(--color-gray-700, #374151)";
export const FG_PRIMARY_FALLBACK_LIGHT = "var(--fg-primary, #111827)";
export const FG_PRIMARY_FALLBACK_DARK = "var(--fg-primary, #f9fafb)";

export type PreviewMode = "light" | "dark";

export interface ThemeColors {
  sequentialColors: string[];
  divergingColors: string[];
  qualitativeColors: string[];
  primaryColor: string;
  backgroundColor: string;
  cardColor: string;
  fgPrimary: string;
  surfaceHeader: string;
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
  // Support both new semantic names and legacy names for backwards compatibility
  const primaryColor = currentModeYaml["primary"] || "var(--color-theme-500)";
  const backgroundColor =
    currentModeYaml["surface-background"] ||
    currentModeYaml["background"] ||
    (previewMode === "light"
      ? BACKGROUND_FALLBACK_LIGHT
      : BACKGROUND_FALLBACK_DARK);
  const cardColor =
    currentModeYaml["surface-card"] ||
    currentModeYaml["card"] ||
    (previewMode === "light" ? CARD_FALLBACK_LIGHT : CARD_FALLBACK_DARK);
  const fgPrimary =
    currentModeYaml["fg-primary"] ||
    (previewMode === "light"
      ? FG_PRIMARY_FALLBACK_LIGHT
      : FG_PRIMARY_FALLBACK_DARK);
  const surfaceHeader =
    currentModeYaml["surface-subtle"] || "var(--surface-subtle)";

  return {
    sequentialColors,
    divergingColors,
    qualitativeColors,
    primaryColor,
    backgroundColor,
    cardColor,
    fgPrimary,
    surfaceHeader,
  };
}

/**
 * Updates a color value in a parsed YAML theme document
 * @param parsedDocument - The parsed YAML document
 * @param mode - The theme mode ("light" or "dark")
 * @param colorKey - The color key to update (e.g., "primary", "surface-background")
 * @param value - The new color value
 * @param removeComments - Whether to remove inline comments from the edited line
 * @returns The updated YAML string
 */
export function updateThemeColor(
  parsedDocument: Document,
  mode: PreviewMode,
  colorKey: string,
  value: string,
  removeComments: boolean = true,
): string {
  // Get or create the mode section (light/dark)
  let modeSection: YAMLMap = parsedDocument.get(mode, true) as YAMLMap;

  // If the mode section doesn't exist or isn't a YAMLMap, create it
  if (!(modeSection instanceof YAMLMap)) {
    const newMap = new YAMLMap();
    parsedDocument.set(mode, newMap);
    modeSection = newMap;
  }

  // Set the color value
  modeSection.set(colorKey, value);

  // Optionally remove inline comment from the edited line
  if (removeComments) {
    const valueNode = modeSection.get(colorKey, true);
    if (
      valueNode &&
      typeof valueNode === "object" &&
      valueNode !== null &&
      "comment" in valueNode
    ) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (valueNode as any).comment = undefined;
    }
  }

  return parsedDocument.toString();
}
