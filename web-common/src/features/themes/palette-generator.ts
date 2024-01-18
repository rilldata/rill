import { TailwindColorPresetsConverted } from "@rilldata/web-common/features/themes/color-config";
import chroma, { type Color } from "chroma-js";

export type ThemeColorShift = [hueShift: number, saturationShift: number];

export function applyShiftToPalette(
  palette: Color[],
  [hueShift, saturationShift]: ThemeColorShift,
): Color[] {
  return palette.map((c) => {
    const [h, s, l] = c.hsl();
    return chroma.hsl(
      Math.min(360, h + hueShift),
      Math.min(1, s + saturationShift),
      l,
    );
  });
}

/**
 * Go through all the tailwind palette and select the one with a shade with the smallest distance with input color.
 * Return the palette and the shift needed for the palette to match the input color.
 */
export function getPaletteAndShift(
  inputColor: Color,
): [palette: Color[], shift: ThemeColorShift] {
  let minPalette = TailwindColorPresetsConverted[0];
  let [paletteMatchIdx, minDistance] = getClosestShade(minPalette, inputColor);

  for (let i = 1; i < TailwindColorPresetsConverted.length; i++) {
    const [matchIdx, distance] = getClosestShade(
      TailwindColorPresetsConverted[i],
      inputColor,
    );
    if (distance >= minDistance) continue;
    minDistance = distance;
    paletteMatchIdx = matchIdx;
    minPalette = TailwindColorPresetsConverted[i];
  }

  return [minPalette, getShift(minPalette[paletteMatchIdx], inputColor)];
}

/**
 * Get the index of the shade in a palette that has the least distance with the input color.
 */
function getClosestShade(
  palette: Color[],
  inputColor: Color,
): [index: number, distance: number] {
  let minIdx = 0;
  let minDistance = chroma.distance(palette[0], inputColor);

  for (let i = 1; i < palette.length; i++) {
    const distance = chroma.distance(palette[i], inputColor);
    if (distance >= minDistance) continue;
    minDistance = distance;
    minIdx = i;
  }

  return [minIdx, minDistance];
}

/**
 * Get the hue and saturation shift for a source color towards a target color.
 */
function getShift(src: Color, tar: Color): ThemeColorShift {
  const [sh, ss] = src.hsl();
  const [th, ts] = tar.hsl();
  return [th - sh, ts - ss];
}
