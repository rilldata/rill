import { TailwindColorSpacing } from "@rilldata/web-common/features/themes/color-config";
import {
  convertColor,
  HexToHSL,
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
  const root = document.querySelector(ThemeBoundarySelector) as HTMLElement;
  if (!root) return;

  generateColorPalette(primary, getDefaultPrimaryColors()).forEach((c, i) => {
    root.style.setProperty(
      `${PrimaryCSSVariablePrefix}${TailwindColorSpacing[i]}`,
      `hsl(${themeColorToHSLString(c)})`,
    );
  });

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

  console.log("before theme update, getPrimaryColors", getPrimaryColors());
  getPrimaryColors().forEach(([cssVar, color]) => {
    root.style.setProperty(
      cssVar,
      `hsl(${themeColorToHSLString(replaceHue(hue, color))})`,
    );
  });
  console.log("after theme update, getPrimaryColors", getPrimaryColors());
}

function setSecondaryColor(secondary: V1Color, variance: number) {
  const [hue] = RGBToHSL(convertColor(secondary));
  const root = document.querySelector(ThemeBoundarySelector) as HTMLElement;
  if (!root) return;

  root.style.setProperty(
    `${SecondaryCSSVariablePrefix}gradient-max-hue`,
    ((hue + variance) % 360) + "",
  );
  root.style.setProperty(
    `${SecondaryCSSVariablePrefix}gradient-min-hue`,
    ((360 + hue - variance) % 360) + "",
  );
}

function themeColorToHSLString([h, s, l]: HSLColor) {
  return `${Number.isNaN(h) ? 0 : h}, ${Math.round(s)}%, ${Math.round(l)}%`;
}

// Get the default primary color defined in app.css
function getDefaultPrimaryColors() {
  const style = getComputedStyle(document.documentElement);
  return TailwindColorSpacing.map((c) => {
    const hex = style.getPropertyValue(`${PrimaryCSSVariablePrefix}${c}`);
    return HexToHSL(hex.substring(1));
  });
}

const getPrimaryColors = () => getPaletteVarsByPrefix("--color-primary-");

/**
 * Gets all the CSS variables that start with the provided prefix.
 * This will enable us to extend the list of colors in the CSS and
 * have them automatically picked up here. This would have been
 * useful in the past when we needed e.g. blue-75, but hard-coded it
 * in components.
 */
function getPaletteVarsByPrefix(prefix: string): [string, HSLColor][] {
  const style = getComputedStyle(document.documentElement);
  // Create a dynamic regex based on the provided prefix.
  // Matches the input prefix followed by a 1 to 3 digit number/
  const regex = new RegExp(`^${prefix}\\d{1,3}$`);
  return Object.values(style)
    .filter((v) => regex.test(v))
    .map((varName) => [varName, HexToHSL(style.getPropertyValue(varName))]);
}

/**
 * Replaces hue in HSL color
 */
function replaceHue(hue: number, hsl: HSLColor): HSLColor {
  return [hue, hsl[1], hsl[2]];
}

/**
 * Right now copies over saturation and lightness from the default primary color, keeping the hue from input
 */
export function generateColorPalette(
  input: V1Color,
  inputColors: HSLColor[],
): HSLColor[] {
  const [hue] = RGBToHSL(convertColor(input));
  return inputColors.map((c) => replaceHue(hue, c));
}
