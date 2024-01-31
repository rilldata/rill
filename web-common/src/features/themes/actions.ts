import {
  defaultPrimaryColors,
  defaultSecondaryColors,
  LightnessMap,
} from "./color-config";
import type { V1Color, V1Theme } from "@rilldata/web-common/runtime-client";
import Color from "colorjs.io";

const ThemeBoundarySelector = ".dashboard-theme-boundary";

export function setTheme(theme: V1Theme) {
  if (theme.spec?.primaryColor)
    updateColorVars("primary", theme.spec?.primaryColor, defaultPrimaryColors);

  if (theme.spec?.secondaryColor)
    updateColorVars(
      "secondary",
      theme.spec?.secondaryColor,
      defaultSecondaryColors,
    );
}

function updateColorVars(
  colorVarKind: "primary" | "secondary" | "muted",
  userThemeColor: V1Color,
  defaultColorMap: LightnessMap,
) {
  const root = document.querySelector(ThemeBoundarySelector) as HTMLElement;
  if (!root) return;

  // get the from the theme primary color
  const huePrimary = new Color(
    "srgb",
    [
      userThemeColor.red ?? 0,
      userThemeColor.green ?? 0,
      userThemeColor.blue ?? 0,
    ],
    userThemeColor.alpha ?? 1,
  ).lch.h;

  // Update CSS variables
  Object.entries(defaultColorMap).forEach(([lightnessNum, colorCssString]) => {
    const color = new Color(colorCssString);
    // update the default color with the hue from the theme color
    color.lch.h = huePrimary;
    root.style.setProperty(
      `--color-${colorVarKind}-${lightnessNum}`,
      color.toString({ format: "lch" }),
    );
  });
}
