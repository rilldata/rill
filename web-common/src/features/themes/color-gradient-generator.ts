import type { TailwindColors } from "@rilldata/web-common/features/themes/color-config";
import type { ThemeColor } from "@rilldata/web-common/features/themes/color-utils";
import convert from "color-convert";

// Based off of https://www.learnui.design/blog/color-in-ui-design-a-practical-framework.html

const CMY_HUES = [180, 300, 60];
const RGB_HUES = [360, 240, 120, 0];

function hueShift(hues: Array<number>, hue: number, intensity: number) {
  const closestHue = hues.sort(
      (a, b) => Math.abs(a - hue) - Math.abs(b - hue)
    )[0],
    hueShift = closestHue - hue;
  return Math.round(intensity * hueShift * 0.5);
}

function lighten([h, s, v]: ThemeColor, intensity: number): ThemeColor {
  const hue = h + hueShift(CMY_HUES, h, intensity);
  const saturation = s - Math.round(s * intensity);
  const value = v + Math.round((100 - v) * intensity);

  return convert.hsv.hsl([hue, saturation, value]);
}

function darken([h, s, v]: ThemeColor, intensity: number): ThemeColor {
  const inverseIntensity = 1 - intensity;
  const hue = h + hueShift(RGB_HUES, h, inverseIntensity);
  const saturation = s + Math.round((100 - s) * inverseIntensity);
  const value = v - Math.round(v * inverseIntensity);

  return convert.hsv.hsl([hue, saturation, value]);
}

const intensityMap: Record<keyof TailwindColors, number> = {
  50: 0.95,
  100: 0.9,
  200: 0.75,
  300: 0.6,
  400: 0.3,
  500: 0,
  600: 0.9,
  700: 0.75,
  800: 0.6,
  900: 0.45,
  950: 0.29,
};
// TODO: instead of assuming the input as 500 try to fit it
export function generateColorGradients(hsl: ThemeColor): TailwindColors {
  const hsv = convert.hsl.hsv(hsl);
  const colors = {} as TailwindColors;
  [50, 100, 200, 300, 400].forEach((level) => {
    colors[level] = lighten(hsv, intensityMap[level]);
  });

  [600, 700, 800, 900, 950].forEach((level) => {
    colors[level] = darken(hsv, intensityMap[level]);
  });

  return colors;
}
