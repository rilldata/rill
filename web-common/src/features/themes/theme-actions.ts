import {
  DefaultPrimaryColors,
  TailwindColorSpacing,
} from "@rilldata/web-common/features/themes/color-config";
import { generateColorGradients } from "@rilldata/web-common/features/themes/color-gradient-generator";
import { generateColorPaletteUsingScale } from "@rilldata/web-common/features/themes/color-palette";
import chroma, { Color } from "chroma-js";

const PrimaryCSSVariablePrefix = "--color-primary-";
const SecondaryCSSVariablePrefix = "--color-secondary-";

export function setTheme({
  primary,
  secondary,
  mode,
}: {
  primary: string;
  secondary: string | null;
  mode: string | null;
}) {
  setPrimaryColor(primary, mode);

  if (secondary) {
    setSecondaryColor(secondary, 80);
  }
}

function setPrimaryColor(primary: string, mode: string | null) {
  const primaryColor = chroma(primary);
  let colors: Array<Color>;
  switch (mode) {
    case "v1":
      colors = generateColorGradients(primaryColor);
      break;

    case "v2":
      colors = generateColorPaletteUsingScale(primaryColor);
      break;

    default:
      colors = copySaturationAndLightness(primaryColor);
      break;
  }

  // console.log(
  //   primary,
  //   mode,
  //   Object.values(colors).map((c) => `${c.hex()}`)
  // );
  const root = document.querySelector(":root") as HTMLElement;

  for (let i = 0; i < TailwindColorSpacing.length; i++) {
    console.log(TailwindColorSpacing[i], colors[i]?.hex());
    root.style.setProperty(
      `${PrimaryCSSVariablePrefix}${TailwindColorSpacing[i]}`,
      `hsl(${themeColorToHSLString(colors[i])})`
    );
  }

  root.style.setProperty(
    `${PrimaryCSSVariablePrefix}graph-line`,
    themeColorToHSLString(colors[TailwindColorSpacing.indexOf(600)])
  );
  root.style.setProperty(
    `${PrimaryCSSVariablePrefix}area-area`,
    themeColorToHSLString(colors[TailwindColorSpacing.indexOf(800)])
  );
}

function setSecondaryColor(secondary: string, variance: number) {
  const secondaryColor = chroma(secondary);
  const [hue] = secondaryColor.hsl();
  const root = document.querySelector(":root") as HTMLElement;

  root.style.setProperty(
    `${SecondaryCSSVariablePrefix}gradient-max`,
    ((hue + variance) % 360) + ""
  );
  root.style.setProperty(
    `${SecondaryCSSVariablePrefix}gradient-min`,
    ((360 + hue - variance) % 360) + ""
  );
}

function themeColorToHSLString(color: Color) {
  const [h, s, l] = color.hsl();
  return `${Number.isNaN(h) ? 0 : h * 360}, ${Math.round(
    s * 100
  )}%, ${Math.round(l * 100)}%`;
}

export function copySaturationAndLightness(input: Color) {
  const colors = new Array<Color>(TailwindColorSpacing.length);
  for (let i = 0; i < DefaultPrimaryColors.length; i++) {
    colors[i] = chroma.hsl(
      input.hsl()[0],
      DefaultPrimaryColors[i].hsl()[1],
      DefaultPrimaryColors[i].hsl()[2]
    );
  }
  return colors;
}
