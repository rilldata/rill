import {
  defaultPrimaryColors,
  defaultSecondaryColors,
} from "@rilldata/web-common/features/themes/color-config";
import chroma, { type Color } from "chroma-js";
import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";

export function resolveThemeColors(
  themeSpec: V1ThemeSpec | undefined,
  isThemeModeDark: boolean,
): { primary: Color; secondary: Color } {
  if (!themeSpec) {
    return {
      primary: chroma(`hsl(${defaultPrimaryColors[500]})`),
      secondary: chroma(`hsl(${defaultSecondaryColors[500]})`),
    };
  }

  const modeTheme = isThemeModeDark ? themeSpec.dark : themeSpec.light;
  const primaryColor = modeTheme?.primary;
  const secondaryColor = modeTheme?.secondary;

  return {
    primary: primaryColor
      ? chroma(primaryColor)
      : chroma(`hsl(${defaultPrimaryColors[500]})`),
    secondary: secondaryColor
      ? chroma(secondaryColor)
      : chroma(`hsl(${defaultSecondaryColors[500]})`),
  };
}

export function resolveThemeObject(
  themeSpec: V1ThemeSpec | undefined,
  isThemeModeDark: boolean,
): Record<string, string> | undefined {
  if (!themeSpec) return undefined;

  const modeTheme = isThemeModeDark ? themeSpec.dark : themeSpec.light;
  if (!modeTheme) return undefined;

  const merged: Record<string, string> = {};

  if (modeTheme.variables) {
    Object.assign(merged, modeTheme.variables);
  }

  if (modeTheme.primary) {
    merged.primary = modeTheme.primary;
  }

  if (modeTheme.secondary) {
    merged.secondary = modeTheme.secondary;
  }

  return Object.keys(merged).length > 0 ? merged : undefined;
}
