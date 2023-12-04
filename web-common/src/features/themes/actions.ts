import {
  DefaultPrimaryColors,
  TailwindColorSpacing,
} from "@rilldata/web-common/features/themes/color-config";
import {
  convertColor,
  RGBToHSL,
} from "@rilldata/web-common/features/themes/color-utils";
import type { ThemeColor } from "@rilldata/web-common/features/themes/color-utils";
import type { V1Color, V1Theme } from "@rilldata/web-common/runtime-client";

const PrimaryCSSVariablePrefix = "--color-primary-";
const SecondaryCSSVariablePrefix = "--color-secondary-";
const ThemeBoundrySelector = ".dashboard-theme-boundary";

export function setTheme(theme: V1Theme) {
  if (theme.spec?.primaryColor) setPrimaryColor(theme.spec?.primaryColor);

  if (theme.spec?.secondaryColor)
    setSecondaryColor(theme.spec?.secondaryColor, 80);
}

function setPrimaryColor(primary: V1Color) {
  const colors = copySaturationAndLightness(primary);

  const root = document.querySelector(ThemeBoundrySelector) as HTMLElement;

  for (let i = 0; i < TailwindColorSpacing.length; i++) {
    root.style.setProperty(
      `${PrimaryCSSVariablePrefix}${TailwindColorSpacing[i]}`,
      `hsl(${themeColorToHSLString(colors[i])})`
    );
  }

  const [hue] = RGBToHSL(convertColor(primary));
  const hueVal = Math.round(hue) + "";
  [
    "--color-primary-graph-line",
    "--color-primary-graph-area",
    "--color-primary-graph-scrubbing-line",
    "--color-primary-graph-scrubbing-area",
    "--color-primary-scrub-box",
    "--color-primary-scrub-area-0",
    "--color-primary-scrub-area-1",
    "--color-primary-scrub-area-2",
  ].forEach((cssVar) => root.style.setProperty(cssVar, hueVal));
}

function setSecondaryColor(secondary: V1Color, variance: number) {
  const [hue] = RGBToHSL(convertColor(secondary));
  const root = document.querySelector(ThemeBoundrySelector) as HTMLElement;

  root.style.setProperty(
    `${SecondaryCSSVariablePrefix}gradient-max`,
    ((hue + variance) % 360) + ""
  );
  root.style.setProperty(
    `${SecondaryCSSVariablePrefix}gradient-min`,
    ((360 + hue - variance) % 360) + ""
  );
}

function themeColorToHSLString([h, s, l]: ThemeColor) {
  return `${Number.isNaN(h) ? 0 : h}, ${Math.round(s)}%, ${Math.round(l)}%`;
}

export function copySaturationAndLightness(input: V1Color) {
  const [hue] = RGBToHSL(convertColor(input));
  const colors = new Array<ThemeColor>(TailwindColorSpacing.length);
  for (let i = 0; i < DefaultPrimaryColors.length; i++) {
    colors[i] = [hue, DefaultPrimaryColors[i][1], DefaultPrimaryColors[i][2]];
  }
  return colors;
}
