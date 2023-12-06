import {
  HexToHSL,
  HSLColor,
} from "@rilldata/web-common/features/themes/color-utils";

export const TailwindColorSpacing = [
  50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950,
];

// should match web-common/src/app.css
// storing the HSL values to easily swap just the Hue from the theme config
export const DefaultPrimaryColors: Array<HSLColor> = [
  HexToHSL("eff6ff"), // 50
  HexToHSL("dbeafe"), // 100
  HexToHSL("bfdbfe"), // 200
  HexToHSL("93c5fd"), // 300
  HexToHSL("60a5fa"), // 400
  HexToHSL("3b82f6"), // 500
  HexToHSL("2563eb"), // 600
  HexToHSL("1d4ed8"), // 700
  HexToHSL("1e40af"), // 800
  HexToHSL("1e3a8a"), // 900
  HexToHSL("172554"), // 950
];
