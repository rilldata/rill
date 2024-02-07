import { generateColorPalette } from "@rilldata/web-common/features/themes/palette-generator";
import { TailwindColorSpacing } from "./color-config";
import type { V1Color, V1Theme } from "@rilldata/web-common/runtime-client";
import chroma from "chroma-js";

const ThemeBoundarySelector = ".dashboard-theme-boundary";

export function setTheme(theme: V1Theme) {
  if (theme.spec?.primaryColor)
    updateColorVars("primary", theme.spec?.primaryColor);

  if (theme.spec?.secondaryColor)
    updateColorVars("secondary", theme.spec?.secondaryColor);
}

function updateColorVars(
  colorVarKind: "primary" | "secondary" | "muted",
  userThemeColor: V1Color,
) {
  const root = document.querySelector(ThemeBoundarySelector) as HTMLElement;
  if (!root) return;

  console.log(userThemeColor);
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
    root.style.setProperty(
      `--color-${colorVarKind}-${TailwindColorSpacing[i]}`,
      c.css(),
    );
  });
}
