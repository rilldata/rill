import type { V1Color } from "@rilldata/web-common/runtime-client";

export type HSLColor = [number, number, number];
export type RGBColor = [number, number, number];

export function convertColor(color: V1Color): RGBColor {
  return [color.red ?? 0, color.green ?? 0, color.blue ?? 0].map(
    (c) => c * 256,
  ) as RGBColor;
}

// These methods are copied over from color-convert.
// Just for these 2 methods, it didnt make sense to add a dependency.
export function RGBToHSL(rgb: RGBColor): HSLColor {
  const r = rgb[0] / 255;
  const g = rgb[1] / 255;
  const b = rgb[2] / 255;
  const min = Math.min(r, g, b);
  const max = Math.max(r, g, b);
  const delta = max - min;
  let h = 0;
  let s: number;

  if (max === min) {
    h = 0;
  } else if (r === max) {
    h = (g - b) / delta;
  } else if (g === max) {
    h = 2 + (b - r) / delta;
  } else if (b === max) {
    h = 4 + (r - g) / delta;
  }

  h = Math.min(h * 60, 360);

  if (h < 0) {
    h += 360;
  }

  const l = (min + max) / 2;

  if (max === min) {
    s = 0;
  } else if (l <= 0.5) {
    s = delta / (max + min);
  } else {
    s = delta / (2 - max - min);
  }

  return [h, s * 100, l * 100];
}

const HexRegex = /[a-f0-9]{6}|[a-f0-9]{3}/i;
export function HexToRGB(hex: string): RGBColor {
  const match = (hex as any).toString(16).match(HexRegex);
  if (!match) {
    return [0, 0, 0];
  }

  let colorString = match[0];

  if (match[0].length === 3) {
    colorString = colorString
      .split("")
      .map((char) => {
        return char + char;
      })
      .join("");
  }

  const integer = parseInt(colorString, 16);
  const r = (integer >> 16) & 0xff;
  const g = (integer >> 8) & 0xff;
  const b = integer & 0xff;

  return [r, g, b];
}

export function HexToHSL(hex: string): HSLColor {
  return RGBToHSL(HexToRGB(hex));
}
