/**
 * Theme Utilities
 *
 * Re-exports theme resolution functions from the centralized theme manager.
 * Maintained for backwards compatibility.
 */

import { themeManager } from "./theme-manager";
import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
import chroma, { type Color } from "chroma-js";

export const resolveThemeColors = (
  themeSpec: V1ThemeSpec | undefined,
  isThemeModeDark: boolean,
): { primary: Color; secondary: Color } =>
  themeManager.resolveThemeColors(themeSpec, isThemeModeDark);

export const resolveThemeObject = (
  themeSpec: V1ThemeSpec | undefined,
  isThemeModeDark: boolean,
): Record<string, string> | undefined =>
  themeManager.resolveThemeObject(themeSpec, isThemeModeDark);

// Default fallback color - using primary blue instead of red for better UX
const FALLBACK_COLOR = "#6366f1"; // indigo-500

export function getChroma(colorString: string | undefined): Color {
  if (!colorString) {
    return chroma(FALLBACK_COLOR);
  }

  try {
    const trimmedString = colorString?.trim();
    return chroma(trimmedString);
  } catch (e) {
    console.error("Invalid color string:", colorString, e);
    return chroma(FALLBACK_COLOR);
  }
}
