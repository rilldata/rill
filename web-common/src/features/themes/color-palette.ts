import { TailwindColorSpacing } from "@rilldata/web-common/features/themes/color-config";
import chroma, { Color } from "chroma-js";

export function generateColorPaletteUsingScale(refColor: Color) {
  const scale = chroma
    .scale(["white", refColor, "black"])
    .domain([0, chromaDomainByLightnessSnapped(refColor), 1000]);

  return TailwindColorSpacing.map((cd) => scale(cd));
}

export function generateColorPaletteUsingScaleDifferentStartAndEnd(
  refColor: Color
) {
  const beg = chroma.scale(["white", refColor]).domain([0, 1]).colors(50)[10];
  const end = chroma.scale(["black", refColor]).domain([0, 1]).colors(50)[10];

  const scale = chroma
    .scale([beg, refColor, end])
    .domain([50, chromaDomainByLightnessSnapped(refColor), 950]);

  return TailwindColorSpacing.map((cd) => scale(cd));
}

export function generateColorPaletteUsingDarken(refColor: Color) {
  const colors = new Array<Color>(TailwindColorSpacing.length);
  const domainIndex = TailwindColorSpacing.indexOf(
    chromaDomainByLightnessSnapped(refColor)
  );
  for (
    let i = domainIndex - 1, c = refColor.brighten();
    i >= 0;
    i--, c = c.brighten()
  ) {
    colors[i] = c;
  }
  colors[domainIndex] = refColor;
  for (
    let i = domainIndex + 1, c = refColor.darken();
    i < TailwindColorSpacing.length;
    i++, c = c.darken()
  ) {
    colors[i] = c;
  }
  return colors;
}

function chromaDomainByLightnessSnapped(refColor: Color) {
  const refHsl = refColor.hsl();
  const bnwPalette = chroma
    .scale(["white", "black"])
    .domain([0, ...TailwindColorSpacing, 1000]);
  let closest = TailwindColorSpacing[0];
  let closestDist = Math.abs(
    bnwPalette(TailwindColorSpacing[0]).hsl()[2] - refHsl[2]
  );

  for (let i = 1; i < TailwindColorSpacing.length; i++) {
    const dist = Math.abs(
      bnwPalette(TailwindColorSpacing[i]).hsl()[2] - refHsl[2]
    );
    if (dist < closestDist) {
      closest = TailwindColorSpacing[i];
      closestDist = dist;
    }
  }

  return closest;
}
