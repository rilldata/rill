import { generateColorPalette } from "@rilldata/web-common/features/themes/palette-generator";
import { TailwindColorSpacing } from "./color-config";
import type { V1Color, V1ThemeSpec } from "@rilldata/web-common/runtime-client";
import chroma from "chroma-js";

const ThemeBoundarySelector = ".dashboard-theme-boundary";

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
