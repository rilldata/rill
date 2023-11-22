import {
  DefaultPrimaryColors,
  TailwindColors,
} from "@rilldata/web-common/features/themes/color-config";
import {
  HexToRGB,
  RGBToHSL,
} from "@rilldata/web-common/features/themes/color-utils";

const PrimaryCSSVariablePrefix = "--color-primary-";

// Temporary function for testing
export function resetToDefault() {
  setToColors(DefaultPrimaryColors);
}

export function setTheme({ primary }: { primary: string }) {
  const primaryHsl = RGBToHSL(HexToRGB(primary));
  const primaryColors = {} as TailwindColors;
  for (const shade in DefaultPrimaryColors) {
    primaryColors[shade] = [
      primaryHsl[0],
      DefaultPrimaryColors[shade][1],
      DefaultPrimaryColors[shade][2],
    ];
  }
  setToColors(primaryColors);
}

function setToColors(primaryColors: TailwindColors) {
  const root = document.querySelector(":root") as HTMLElement;

  for (const shade in primaryColors) {
    const h = Math.round(primaryColors[shade][0]);
    const s = Math.round(primaryColors[shade][1]);
    const l = Math.round(primaryColors[shade][2]);
    root.style.setProperty(
      `${PrimaryCSSVariablePrefix}${shade}`,
      `hsl(${h}, ${s}%, ${l}%)`
    );
  }
}
