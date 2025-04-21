import { TailwindColorSpacing } from "./color-config";
import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
import chroma, { type Color } from "chroma-js";
import { allColors } from "./colors";
import { clamp } from "@rilldata/web-common/lib/clamp";

const root = document.documentElement;
const grays = ["neutral", "stone", "slate", "zinc", "gray"];

export const contrasts = Object.values(allColors.purple).map((color) => {
  return chroma.contrast(chroma("white"), color);
});

export function initColors() {
  Object.entries(allColors).forEach(([colorName, colors]) => {
    const scale = createDarkVariation(colorName, Object.values(colors));

    scale.forEach((chromaColor, i) => {
      root.style.setProperty(
        `--color-${colorName}-dark-${TailwindColorSpacing[i]}`,
        chromaColor.css("oklch"),
      );
    });
  });
}

function createDarkVariation(name: string, colors: Color[]) {
  if (!grays.includes(name)) {
    return genPal(colors[5]).dark;
  } else {
    const deSat = colors.reverse().map((color) => color.desaturate(20));

    const shadows = deSat.slice(0, 4);
    const highs = deSat.slice(5, 10);
    const left = chroma.scale(shadows).mode("oklch").gamma(1.6).colors(4, null);
    const right = chroma.scale(highs).mode("oklch").gamma(0.8).colors(7, null);

    return [...left, ...right];
  }
}

export function findValues(
  colors: Color[],
  contrastCurve: number[],
  from: Color,
) {
  let i = 0;

  return contrastCurve.map((c) => {
    const color = colors[i];
    let contrast = chroma.contrast(from, color);

    while (contrast < c && i < colors.length - 1) {
      i++;
      contrast = chroma.contrast(from, colors[i]);
    }
    return colors[i];
  });
}

export function findContrast(
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

const BLACK = chroma("black");
const WHITE = chroma("white");
const MODE = "oklab";

export function genPal(
  refColor: Color,
  stepCount: number = 11,
  gamma: number = 1.12,
) {
  const [l, c, h] = refColor.oklch();

  const isGray = c < 0.1;

  if (isGray) {
    const spectrum = chroma
      .scale([chroma("black"), refColor.desaturate(0.2), chroma("white")])
      .mode(MODE)
      .gamma(1)
      .colors(102, null);

    const darkestDark = findLuminance(spectrum, 0.194) ?? spectrum[0];
    const lightestDark = findLuminance(spectrum, 0.92) ?? spectrum[0];

    return {
      dark: chroma
        .scale([darkestDark, lightestDark])
        .mode(MODE)
        .gamma(gamma)
        .colors(stepCount, null),
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
  if (theme?.primaryColor) {
    const chromaColor = chroma.rgb(
      (theme.primaryColor.red ?? 1) * 256,
      (theme.primaryColor.green ?? 1) * 256,
      (theme.primaryColor.blue ?? 1) * 256,
    );

    const { light, dark } = genPal(chromaColor);

    setVariables(root, "theme", "light", light);
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

    const { light, dark } = genPal(chromaColor);

    setVariables(root, "theme-secondary", "light", light);
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

export function getColors(baseColor: Color, dark: boolean): Color[] {
  const [, c, h] = baseColor.oklch();
  const steps = TailwindColorSpacing.length;

  // We define a range of lightness values
  const minL = 0.02;
  const maxL = 1;

  const range = Array.from({ length: steps }, (_, i) => {
    // Invert the index if dark mode
    const idx = dark ? steps - 1 - i : i;
    const lightness = maxL - (idx / (steps - 1)) * (maxL - minL);

    return chroma.oklch(lightness, c, h);
  });

  return range;
}
