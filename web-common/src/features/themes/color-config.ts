import chroma, { Color } from "chroma-js";

export const TailwindColorSpacing = [
  50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950,
];

// should match web-common/src/app.css
// storing the HSL values to easily swap just the Hue from the theme config
export const DefaultPrimaryColors: Array<Color> = [
  chroma("eff6ff"), // 50
  chroma("dbeafe"), // 100
  chroma("bfdbfe"), // 200
  chroma("93c5fd"), // 300
  chroma("60a5fa"), // 400
  chroma("3b82f6"), // 500
  chroma("2563eb"), // 600
  chroma("1d4ed8"), // 700
  chroma("1e40af"), // 800
  chroma("1e3a8a"), // 900
  chroma("172554"), // 950
];
