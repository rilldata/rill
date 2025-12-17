/**
 * Default Theme Values
 *
 * Provides default theme color values for light and dark modes.
 * Values are extracted from the theme manager and CSS variables at runtime,
 * avoiding duplication of default color definitions.
 */

import { primary, secondary } from "./colors";
import chroma from "chroma-js";

/**
 * Get the CSS string representation of a chroma color
 */
function getColorCSS(color: chroma.Color): string {
  return color.css("hsl");
}

/**
 * Map theme variable names to their corresponding CSS variable names
 */
function getCSSVariableName(variableName: string): string {
  // Core UI colors map directly to CSS variables
  if (variableName === "background") return "--background";
  if (variableName === "surface") return "--surface";
  if (variableName === "card") return "--card";
  if (variableName === "foreground") return "--foreground";
  if (variableName === "ring") return "--ring";
  if (variableName === "border") return "--border";
  if (variableName === "input") return "--input";
  if (variableName === "muted") return "--muted";
  if (variableName === "accent") return "--accent";
  if (variableName === "destructive") return "--destructive";
  if (variableName === "primary") return "--primary";
  if (variableName === "secondary") return "--secondary";

  // Palette colors map directly
  if (variableName.startsWith("color-")) {
    return `--${variableName}`;
  }

  // Other variables with dashes
  return `--${variableName}`;
}

/**
 * Get default value for primary or secondary from colors.ts
 */
function getPrimarySecondaryDefault(
  variableName: "primary" | "secondary",
): string {
  const color = variableName === "primary" ? primary[500] : secondary[500];
  return getColorCSS(color);
}

/**
 * Get the default theme value for a given variable and mode
 * @param variableName - The theme variable name (e.g., "primary", "color-sequential-1")
 * @param mode - The theme mode ("light" or "dark")
 * @returns The default CSS color value, or empty string if not found
 */
export function getDefaultThemeValue(
  variableName: string,
  _mode: "light" | "dark",
): string {
  // For primary and secondary, use the default from colors.ts
  if (variableName === "primary" || variableName === "secondary") {
    return getPrimarySecondaryDefault(variableName);
  }

  // For all other variables, read from CSS variables at runtime
  // This avoids duplicating default values - they come from app.css
  if (typeof window !== "undefined") {
    const cssVarName = getCSSVariableName(variableName);
    
    // Get the computed value from the document root
    // The CSS variables are defined in app.css with appropriate light/dark values
    const computed = getComputedStyle(document.documentElement)
      .getPropertyValue(cssVarName)
      .trim();
    
    if (computed) {
      // Convert oklch to hsl if needed (browser may compute CSS vars as oklch)
      // This ensures the color picker can handle the value
      if (computed.startsWith("oklch(")) {
        try {
          const hslColor = chroma(computed).css("hsl");
          return hslColor;
        } catch {
          // If conversion fails, return the original
          return computed;
        }
      }
      return computed;
    }
  }

  // SSR fallback: return empty string
  return "";
}

/**
 * Check if a value matches the default for a given variable and mode
 * @param variableName - The theme variable name
 * @param value - The value to check
 * @param mode - The theme mode
 * @returns True if the value matches the default
 */
export function isDefaultValue(
  variableName: string,
  value: string,
  mode: "light" | "dark",
): boolean {
  const defaultValue = getDefaultThemeValue(variableName, mode);
  if (!defaultValue || !value) return false;

  // Normalize both values for comparison (handle different formats)
  try {
    // Use chroma to normalize colors for comparison
    const defaultColor = chroma(defaultValue);
    const valueColor = chroma(value);
    return defaultColor.hex() === valueColor.hex();
  } catch {
    // If chroma can't parse, do string comparison
    return defaultValue.trim() === value.trim();
  }
}

