import { generateColorPalette } from "@rilldata/web-common/features/themes/palette-generator";
import { TailwindColorSpacing } from "./color-config";
import type { V1Color, V1ThemeSpec } from "@rilldata/web-common/runtime-client";
import chroma, { type Color } from "chroma-js";
import { get } from "svelte/store";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { allColors } from "./colors";

const root = document.documentElement;

const ThemeBoundarySelector = ".dashboard-theme-boundary";

export const viewModeStore = localStorageStore("theme", "light");

const darkQuery = window.matchMedia("(prefers-color-scheme: dark)");

export const contrasts = Object.values(allColors.purple).map((color) => {
  return chroma.contrast(chroma("white"), color);
});

darkQuery.addEventListener("change", ({ matches }) => {
  if (get(viewModeStore) !== "system") return;

  // get root html element

  if (matches) {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
});

viewModeStore.subscribe((viewMode) => {
  if (viewMode === "system") {
    if (darkQuery.matches) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  } else {
    if (viewMode === "dark") {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
    // document.documentElement.setAttribute("data-theme", viewMode);
  }
});

const MIN_CONTRAST = 1.45;
const MIN_DARK_CONTRAST = 1.2;

// const tailwind = [
//   { name: "red", value: chroma.oklch(0.637, 0.237, 25.331) },
//   {
//     name: "orange",
//     value: chroma.oklch(0.705, 0.213, 47.604),
//   },
//   { name: "amber", value: chroma.oklch(0.769, 0.188, 70.08) },
//   { name: "yellow", value: chroma.oklch(0.795, 0.184, 86.047) },
//   {
//     name: "lime",
//     value: chroma.oklch(0.768, 0.233, 130.85),
//   },
//   {
//     name: "green",
//     value: chroma.oklch(0.723, 0.219, 149.579),
//   },
//   { name: "emerald", value: chroma.oklch(0.696, 0.17, 162.48) },
//   { name: "teal", value: chroma.oklch(0.704, 0.14, 182.503) },
//   { name: "cyan", value: chroma.oklch(0.715, 0.143, 215.221) },
//   { name: "sky", value: chroma.oklch(0.685, 0.169, 237.323) },
//   {
//     name: "blue",
//     value: chroma.oklch(0.623, 0.214, 259.815),
//   },
//   { name: "indigo", value: chroma.oklch(0.585, 0.233, 277.117) },
//   { name: "violet", value: chroma.oklch(0.606, 0.25, 292.717) },
//   {
//     name: "purple",
//     value: chroma.oklch(0.627, 0.265, 303.9),
//   },
//   {
//     name: "fuchsia",
//     value: chroma.oklch(0.667, 0.295, 322.15),
//   },
//   { name: "pink", value: chroma.oklch(0.656, 0.241, 354.308) },
//   { name: "rose", value: chroma.oklch(0.586, 0.253, 17.585) },
//   { name: "gray", value: chroma.oklch(0.446, 0.03, 256.802) },
// ];

// const root = document.documentElement;

Object.entries(allColors).forEach(([colorName, colors]) => {
  // const colors = getColors(tailwindColor.value, true);

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
