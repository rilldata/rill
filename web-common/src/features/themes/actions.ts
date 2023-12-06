import {
  DefaultPrimaryColors,
  TailwindColorSpacing,
} from "@rilldata/web-common/features/themes/color-config";
import {
  convertColor,
  RGBToHSL,
} from "@rilldata/web-common/features/themes/color-utils";
import type { HSLColor } from "@rilldata/web-common/features/themes/color-utils";
import type { V1Color, V1Theme } from "@rilldata/web-common/runtime-client";

const PrimaryCSSVariablePrefix = "--color-primary-";
const SecondaryCSSVariablePrefix = "--color-secondary-";
const ThemeBoundarySelector = ".dashboard-theme-boundary";

export function setTheme(theme: V1Theme) {
  if (theme.spec?.primaryColor) setPrimaryColor(theme.spec?.primaryColor);

  if (theme.spec?.secondaryColor)
    setSecondaryColor(theme.spec?.secondaryColor, 80);
}

function setPrimaryColor(primary: V1Color) {
  const colors = generateColorPalette(primary);

  const root = document.querySelector(ThemeBoundarySelector) as HTMLElement;
  if (!root) return;

  for (let i = 0; i < TailwindColorSpacing.length; i++) {
    root.style.setProperty(
      `${PrimaryCSSVariablePrefix}${TailwindColorSpacing[i]}`,
      `hsl(${themeColorToHSLString(colors[i])})`
    );
  }

  const [hue] = RGBToHSL(convertColor(primary));
  const hueVal = Math.round(hue) + "";
  [
    "--color-primary-graph-line-hue",
    "--color-primary-graph-area-hue",
    "--color-primary-graph-scrubbing-line-hue",
    "--color-primary-graph-scrubbing-area-hue",
    "--color-primary-scrub-box-hue",
    "--color-primary-scrub-area-0-hue",
    "--color-primary-scrub-area-1-hue",
    "--color-primary-scrub-area-2-hue",
  ].forEach((cssVar) => root.style.setProperty(cssVar, hueVal));
}

function setSecondaryColor(secondary: V1Color, variance: number) {
  const [hue] = RGBToHSL(convertColor(secondary));
  const root = document.querySelector(ThemeBoundarySelector) as HTMLElement;
  if (!root) return;

  root.style.setProperty(
    `${SecondaryCSSVariablePrefix}gradient-max-hue`,
    ((hue + variance) % 360) + ""
  );
  root.style.setProperty(
    `${SecondaryCSSVariablePrefix}gradient-min-hue`,
    ((360 + hue - variance) % 360) + ""
  );
}

function themeColorToHSLString([h, s, l]: HSLColor) {
  return `${Number.isNaN(h) ? 0 : h}, ${Math.round(s)}%, ${Math.round(l)}%`;
}

/**
 * Right now copies over saturation and lightness from the default primary color of blue, keeping the hue from input
 */
export function generateColorPalette(input: V1Color) {
  const [hue] = RGBToHSL(convertColor(input));
  const colors = new Array<HSLColor>(TailwindColorSpacing.length);
  for (let i = 0; i < DefaultPrimaryColors.length; i++) {
    colors[i] = [hue, DefaultPrimaryColors[i][1], DefaultPrimaryColors[i][2]];
  }
  return colors;
}
