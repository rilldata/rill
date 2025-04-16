import { generateColorPalette } from "@rilldata/web-common/features/themes/palette-generator";
import { TailwindColorSpacing } from "./color-config";
import type { V1Color, V1ThemeSpec } from "@rilldata/web-common/runtime-client";
import chroma, { type Color } from "chroma-js";
import { allColors } from "./colors";

const root = document.documentElement;

const ThemeBoundarySelector = ".dashboard-theme-boundary";

export const contrasts = Object.values(allColors.purple).map((color) => {
  return chroma.contrast(chroma("white"), color);
});

const MIN_CONTRAST = 1.45;
const MIN_DARK_CONTRAST = 1.2;

Object.entries(allColors).forEach(([colorName, colors]) => {
  Object.entries(colors).forEach(([_, chroma], i) => {
    root.style.setProperty(
      `--color-${colorName}-dark-${TailwindColorSpacing[10 - i]}`,
      chroma.css("oklch"),
    );
  });
});

export function setTheme(theme: V1ThemeSpec | undefined) {
  if (!theme) return;
  if (theme.primaryColor) updateColorVars("primary", theme.primaryColor);

  if (theme.secondaryColor) updateColorVars("secondary", theme.secondaryColor);
}

function updateColorVars(
  colorVarKind: "primary" | "secondary" | "muted",
  userThemeColor: V1Color,
) {
  const root = document.querySelector(ThemeBoundarySelector) as HTMLElement;
  if (!root) return;

  // get the color from the theme primary color
  const inputColor = chroma.rgb(
    (userThemeColor.red ?? 0) * 256,
    (userThemeColor.green ?? 0) * 256,
    (userThemeColor.blue ?? 0) * 256,
    userThemeColor.alpha ?? 1,
  );
  const palette = generateColorPalette(inputColor);
  // Update CSS variables
  palette.forEach((c, i) => {
    const hsl = c.css("hsl");

    root.style.setProperty(
      `--hsl-${colorVarKind}-${TailwindColorSpacing[i]}`,
      hsl.slice(4, -1).split(",").join(" "),
    );

    root.style.setProperty(
      `--color-${colorVarKind}-${TailwindColorSpacing[i]}`,
      hsl,
    );
  });
}

export function getColors(baseColor: Color, dark: boolean): Color[] {
  const [l, c, h] = baseColor.oklch();
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

function findBounds(colors: chroma.Color[], dark: boolean) {
  let leftBound = colors.find(
    (c) =>
      chroma.contrast(chroma("black"), c) >
      (dark ? MIN_DARK_CONTRAST : MIN_CONTRAST),
  );
  let rightBound = colors
    .reverse()
    .find(
      (c) =>
        chroma.contrast(chroma("white"), c) > (dark ? 1.09 : MIN_DARK_CONTRAST),
    );

  return {
    leftBound: dark ? leftBound : rightBound,
    rightBound: dark ? rightBound : leftBound,
  };
}
