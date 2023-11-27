import chroma, { Color } from "chroma-js";

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

function lighten([h, s, v]: [number, number, number], intensity: number) {
  const hue = h + hueShift(CMY_HUES, h, intensity);
  const saturation = s - Math.round(s * intensity);
  const value = v + Math.round((100 - v) * intensity);

  return chroma.hsv(hue, saturation, value);
}

function darken([h, s, v]: [number, number, number], intensity: number) {
  const inverseIntensity = 1 - intensity;
  const hue = h + hueShift(RGB_HUES, h, inverseIntensity);
  const saturation = s + Math.round((100 - s) * inverseIntensity);
  const value = v - Math.round(v * inverseIntensity);

  return chroma.hsv(hue, saturation, value);
}

const intensityMap = {
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
export function generateColorGradients(input: Color) {
  const hsv = input.hsv();
  const colors = new Array<Color>();

  [50, 100, 200, 300, 400].forEach((level) => {
    colors.push(lighten(hsv, intensityMap[level]));
  });
  colors.push(input);
  [600, 700, 800, 900, 950].forEach((level) => {
    colors.push(darken(hsv, intensityMap[level]));
  });

  return colors;
}
