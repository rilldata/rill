import {
  defaultPrimaryColors,
  defaultSecondaryColors,
} from "@rilldata/web-common/features/themes/color-config";
import chroma, { type Color } from "chroma-js";
import type { V1ThemeSpec as RuntimeV1ThemeSpec } from "@rilldata/web-common/runtime-client";
import type { ThemeModeColors, V1ThemeSpec } from "./theme-types";

export function resolveThemeColors(
  themeSpec: RuntimeV1ThemeSpec | V1ThemeSpec | undefined,
  isThemeModeDark: boolean,
): { primary: Color; secondary: Color } {
  const spec = themeSpec as V1ThemeSpec | undefined;
  const modeTheme = isThemeModeDark
    ? (spec?.dark as ThemeModeColors | undefined)
    : (spec?.light as ThemeModeColors | undefined);

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
  themeSpec: RuntimeV1ThemeSpec | V1ThemeSpec | undefined,
  isThemeModeDark: boolean,
): Record<string, string> | undefined {
  const spec = themeSpec as V1ThemeSpec | undefined;
  const modeTheme = isThemeModeDark
    ? (spec?.dark as ThemeModeColors | undefined)
    : (spec?.light as ThemeModeColors | undefined);

  if (modeTheme) {
    const merged: Record<string, string> = { ...modeTheme.variables };
    if (modeTheme.primary) merged.primary = modeTheme.primary;
    if (modeTheme.secondary) merged.secondary = modeTheme.secondary;
    return merged;
  }

  return undefined;
}

