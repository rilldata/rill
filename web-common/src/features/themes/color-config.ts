import {
  HexToRGB,
  RGBToHSL,
} from "@rilldata/web-common/features/themes/color-utils";
import type { ThemeColor } from "@rilldata/web-common/features/themes/color-utils";

export type TailwindColors = {
  50: ThemeColor;
  100: ThemeColor;
  200: ThemeColor;
  300: ThemeColor;
  400: ThemeColor;
  500: ThemeColor;
  600: ThemeColor;
  700: ThemeColor;
  800: ThemeColor;
  900: ThemeColor;
  950: ThemeColor;
};

// should match web-common/src/app.css
// storing the HSL values to easily swap just the Hue from the theme config
export const DefaultPrimaryColors: TailwindColors = {
  50: RGBToHSL(HexToRGB("eff6ff")),
  100: RGBToHSL(HexToRGB("dbeafe")),
  200: RGBToHSL(HexToRGB("bfdbfe")),
  300: RGBToHSL(HexToRGB("93c5fd")),
  400: RGBToHSL(HexToRGB("60a5fa")),
  500: RGBToHSL(HexToRGB("3b82f6")),
  600: RGBToHSL(HexToRGB("2563eb")),
  700: RGBToHSL(HexToRGB("1d4ed8")),
  800: RGBToHSL(HexToRGB("1e40af")),
  900: RGBToHSL(HexToRGB("1e3a8a")),
  950: RGBToHSL(HexToRGB("172554")),
};
