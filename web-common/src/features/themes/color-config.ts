import type { ThemeColor } from "@rilldata/web-common/features/themes/color-utils";
import convert from "color-convert";

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
  50: convert.hex.hsl("eff6ff"),
  100: convert.hex.hsl("dbeafe"),
  200: convert.hex.hsl("bfdbfe"),
  300: convert.hex.hsl("93c5fd"),
  400: convert.hex.hsl("60a5fa"),
  500: convert.hex.hsl("3b82f6"),
  600: convert.hex.hsl("2563eb"),
  700: convert.hex.hsl("1d4ed8"),
  800: convert.hex.hsl("1e40af"),
  900: convert.hex.hsl("1e3a8a"),
  950: convert.hex.hsl("172554"),
};
