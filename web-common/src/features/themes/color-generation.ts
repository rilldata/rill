/**
 * Color Palette Generation
 *
 * Functions for generating color palettes for light and dark modes
 * using the OKLab color space for perceptually uniform colors.
 */

import chroma, { type Color } from "chroma-js";
import { clamp } from "../../../../web-common/src/lib/clamp.ts";

export const BLACK = chroma("black");
export const WHITE = chroma("white");
export const MODE = "oklab";

export const DEFAULT_STEP_COUNT = 11;
export const DEFAULT_GAMMA = 1.12;

/**
 * Generates light and dark color palettes from a reference color
 */
export function generatePalette(
  refColor: Color,
  desaturateNearGray: boolean = true,
  stepCount: number = 11,
  gamma: number = 1.12,
): { light: Color[]; dark: Color[] } {
  const [l, c, h] = refColor.oklch();

  if (desaturateNearGray && c < 0.05) {
    const spectrum = chroma
      .scale([BLACK, refColor, WHITE])
      .mode(MODE)
      .gamma(1)
      .colors(102, null);

    const darkestDark = findContrast(spectrum, BLACK, 1.06) ?? spectrum[0];
    const lightestDark =
      findContrast(spectrum, darkestDark, 18.5) ?? spectrum[0];

    const middleValue = findContrast(spectrum, darkestDark, 5) ?? spectrum[50];

    return {
      dark: [
        ...chroma
          .scale([darkestDark, middleValue])
          .mode(MODE)
          .gamma(1.15)
          .colors(6, null)
          .slice(0, -1),
        middleValue,
        ...chroma
          .scale([middleValue, lightestDark])
          .mode(MODE)
          .gamma(0.6)
          .colors(6, null)
          .slice(1),
      ],
      light: chroma
        .scale([WHITE, BLACK])
        .mode(MODE)
        .gamma(1)
        .colors(stepCount + 2, null)
        .slice(1, -1),
    };
  }

  const darkRef = chroma.oklch(clamp(0.6, l, 1), c, h);

  const lightRef = chroma.oklch(clamp(0, l, 0.82), c, h);

  const darkSpectrum = chroma
    .scale([BLACK, darkRef, WHITE])
    .mode(MODE)
    .gamma(1)
    .colors(102, null);

  const lightSpectrum = chroma
    .scale([BLACK, lightRef, WHITE])
    .mode(MODE)
    .gamma(1)
    .colors(102, null);

  const reversedDarkSpectrum = [...darkSpectrum].reverse();
  const reversedLightSpectrum = [...lightSpectrum].reverse();

  const darkestDark = findLuminance(darkSpectrum, 0.25) ?? darkSpectrum[0];
  const find700 =
    findContrast(darkSpectrum, darkestDark, 10.8, 0.088, 0.073) ??
    darkSpectrum[0];

  const find300 = findContrast(darkSpectrum, darkestDark, 2) ?? darkSpectrum[0];

  const lightestDark =
    findLuminance(darkSpectrum, 0.965) ?? reversedDarkSpectrum[0];

  const lightestLight =
    findContrast(reversedLightSpectrum, WHITE, 1.065) ??
    reversedLightSpectrum[0];
  const darkestLight =
    findContrast(lightSpectrum, BLACK, 1.2) ?? lightSpectrum[0];

  return {
    dark: chroma
      .scale([darkestDark, find300, darkRef, find700, lightestDark])
      .mode(MODE)
      .gamma(gamma)
      .colors(stepCount, null),
    light: chroma
      .scale([lightestLight, lightRef, darkestLight])
      .mode(MODE)
      .gamma(gamma)
      .colors(stepCount, null),
  };
}

/**
 * Creates dark mode variation of a color palette
 */
export function createDarkVariation(colors: Color[]): Color[] {
  return generatePalette(colors[5]).dark;
}

function findLuminance(spectrum: Color[], lumn: number): Color | undefined {
  return spectrum.find((color) => {
    const [l] = color.oklch();
    return l >= lumn;
  });
}

function findContrast(
  spectrum: Color[],
  comparedTo: Color,
  contrast: number,
  maxSaturation?: number,
  minSaturation?: number,
): Color | undefined {
  const color = spectrum.find((color) => {
    return chroma.contrast(comparedTo, color) >= contrast;
  });

  const saturation = color?.oklch()[1] ?? 0;

  if (maxSaturation && saturation > maxSaturation) {
    return color?.set("oklch.c", maxSaturation);
  } else if (minSaturation && saturation < minSaturation) {
    return color?.set("oklch.c", minSaturation);
  }

  return color;
}
