export type ThemeColor = [number, number, number];

export function RGBToHSL(rgb: ThemeColor): ThemeColor {
  let rdif: number;
  let gdif: number;
  let bdif: number;
  let h = 0;
  let s: number;

  const r = rgb[0] / 255;
  const g = rgb[1] / 255;
  const b = rgb[2] / 255;
  const v = Math.max(r, g, b);
  const diff = v - Math.min(r, g, b);
  const diffc = function (c: number) {
    return (v - c) / 6 / diff + 1 / 2;
  };

  if (diff === 0) {
    h = 0;
    s = 0;
  } else {
    s = diff / v;
    rdif = diffc(r);
    gdif = diffc(g);
    bdif = diffc(b);

    if (r === v) {
      h = bdif - gdif;
    } else if (g === v) {
      h = 1 / 3 + rdif - bdif;
    } else if (b === v) {
      h = 2 / 3 + gdif - rdif;
    }

    if (h < 0) {
      h += 1;
    } else if (h > 1) {
      h -= 1;
    }
  }

  return [Math.round(h * 360), Math.round(s * 100), Math.round(v * 100)];
}

const HexRegex = /[a-f0-9]{6}|[a-f0-9]{3}/i;
export function HexToRGB(hex: string): ThemeColor {
  const match = (hex as any).toString(16).match(HexRegex);
  if (!match) {
    return [0, 0, 0];
  }

  let colorString = match[0];

  if (match[0].length === 3) {
    colorString = colorString
      .split("")
      .map((char) => {
        return char + char;
      })
      .join("");
  }

  const integer = parseInt(colorString, 16);
  const r = (integer >> 16) & 0xff;
  const g = (integer >> 8) & 0xff;
  const b = integer & 0xff;

  return [r, g, b];
}

export function HexToHSL(hex: string): ThemeColor {
  const hsl = RGBToHSL(HexToRGB(hex));
  console.log("hex", hsl);
  return hsl;
}
