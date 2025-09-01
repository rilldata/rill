import { TailwindColorSpacing } from "./color-config.ts";
import type { V1ThemeSpec } from "../../../../web-common/src/runtime-client/index.ts";
import chroma, { type Color } from "chroma-js";
import { clamp } from "../../../../web-common/src/lib/clamp.ts";
import { generateColorPalette } from "./palette-generator.ts";
import { featureFlags } from "../feature-flags.ts";
import { get } from "svelte/store";

export const BLACK = chroma("black");
export const WHITE = chroma("white");
export const MODE = "oklab";

export function createDarkVariation(colors: Color[]) {
  return generatePalette(colors[5]).dark;
}

function generatePalette(
  refColor: Color,
  desaturateNearGray: boolean = true,
  stepCount: number = 11,
  gamma: number = 1.12,
) {
  const [l, c, h] = refColor.oklch();

  if (desaturateNearGray && c < 0.05) {
    const spectrum = chroma
      .scale([chroma("black"), refColor, chroma("white")])
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
        .scale([chroma("white"), chroma("black")])
        .mode(MODE)
        .gamma(1)
        .colors(stepCount + 2, null)
        .slice(1, -1),
    };
  }

  const darkRef = chroma.oklch(clamp(0.6, l, 1), c, h);

  const lightRef = chroma.oklch(clamp(0, l, 0.82), c, h);

  const darkSpectrum = chroma
    .scale([chroma("black"), darkRef, chroma("white")])
    .mode(MODE)
    .gamma(1)
    .colors(102, null);

  const lightSpectrum = chroma
    .scale([chroma("black"), lightRef, chroma("white")])
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

export function setVariables(
  root: HTMLElement,
  type: string,
  mode: "dark" | "light",

  colors?: Color[],
) {
  if (!colors) {
    TailwindColorSpacing.forEach((_, i) => {
      root.style.removeProperty(
        `--color-${type}-${mode}-${TailwindColorSpacing[i]}`,
      );
    });
  } else {
    colors.forEach((color, i) => {
      root.style.setProperty(
        `--color-${type}-${mode}-${TailwindColorSpacing[i]}`,
        color.css("oklch"),
      );
    });
  }
}

export function updateThemeVariables(theme: V1ThemeSpec | undefined) {
  const root = document.documentElement;
  const { darkMode } = featureFlags;
  const allowNewPalette = get(darkMode);

  if (theme?.primaryColor) {
    const chromaColor = chroma.rgb(
      (theme.primaryColor.red ?? 1) * 256,
      (theme.primaryColor.green ?? 1) * 256,
      (theme.primaryColor.blue ?? 1) * 256,
    );

    const originalLightPalette = generateColorPalette(chromaColor);
    const { light, dark } = generatePalette(chromaColor, false);

    setVariables(
      root,
      "theme",
      "light",
      allowNewPalette ? light : originalLightPalette,
    );

    setVariables(root, "theme", "dark", dark);
  } else {
    setVariables(root, "theme", "light");
    setVariables(root, "theme", "dark");
  }

  if (theme?.secondaryColor) {
    const chromaColor = chroma.rgb(
      (theme.secondaryColor.red ?? 1) * 256,
      (theme.secondaryColor.green ?? 1) * 256,
      (theme.secondaryColor.blue ?? 1) * 256,
    );

    const originalLightPalette = generateColorPalette(chromaColor);
    const { light, dark } = generatePalette(chromaColor, false);

    setVariables(
      root,
      "theme-secondary",
      "light",
      allowNewPalette ? light : originalLightPalette,
    );
    setVariables(root, "theme-secondary", "dark", dark);
  } else {
    setVariables(root, "theme-secondary", "light");
    setVariables(root, "theme-secondary", "dark");
  }
}

function findLuminance(spectrum: Color[], lumn: number) {
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
) {
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
