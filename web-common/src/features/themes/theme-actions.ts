import {
  DefaultPrimaryColors,
  TailwindColors,
} from "@rilldata/web-common/features/themes/color-config";
import { generateColorGradients } from "@rilldata/web-common/features/themes/color-gradient-generator";
import type { ThemeColor } from "@rilldata/web-common/features/themes/color-utils";
import convert from "color-convert";

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

function copySaturationAndLightness(hsl: ThemeColor) {
  const colors = {} as TailwindColors;
  for (const shade in DefaultPrimaryColors) {
    colors[shade] = [
      hsl[0],
      DefaultPrimaryColors[shade][1],
      DefaultPrimaryColors[shade][2],
    ];
  }
  return colors;
}

function setPrimaryColor(primary: string, mode: string | null) {
  const primaryHsl = convert.hex.hsl(primary);
  let colors: TailwindColors;
  switch (mode) {
    case "smart":
      colors = generateColorGradients(primaryHsl);
      break;

    //TODO: alternatives,
    //      https://www.npmjs.com/package/apca-w3
    //      https://gka.github.io/chroma.js/

    default:
      colors = copySaturationAndLightness(primaryHsl);
      break;
  }

  console.log(
    primary,
    mode,
    Object.values(colors).map((c) => `${c}`)
  );
  const root = document.querySelector(":root") as HTMLElement;

  for (const shade in colors) {
    root.style.setProperty(
      `${PrimaryCSSVariablePrefix}${shade}`,
      `hsl(${themeColorToHSLString(colors[shade])})`
    );
  }

  root.style.setProperty(
    `${PrimaryCSSVariablePrefix}graph-line`,
    themeColorToHSLString(colors["600"])
  );
  root.style.setProperty(
    `${PrimaryCSSVariablePrefix}area-area`,
    themeColorToHSLString(colors["800"])
  );
}

function setSecondaryColor(secondary: string, variance: number) {
  const [hue] = convert.hex.hsl(secondary);
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

function themeColorToHSLString(color: ThemeColor) {
  return `${color[0]}, ${color[1]}%, ${color[2]}%`;
}
