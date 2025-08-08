export function extractHSL(color: string) {
  const [, hue, saturation, lightness] = color.match(
    /hsl\((\d+(?:\.\d+)?)(?:deg)?[,\s]+(\d+(?:\.\d+)?)%[,\s]+(\d+(?:\.\d+)?)%\)/,
  ) || [null, 0, 100, 50];

  return {
    h: +hue,
    s: +saturation,
    l: +lightness,
  };
}

export function stringColorToHsl(string: string | undefined) {
  if (!string) {
    return {
      h: 0,
      s: 100,
      l: 50,
    };
  }
  if (string.startsWith("hsl")) {
    return extractHSL(string);
  }

  const color = getComputedColor(
    hexColorWithoutPound(string) ? `#${string}` : string,
  );

  if (!color)
    return {
      h: 0,
      s: 100,
      l: 50,
    };

  if (color.startsWith("rgba")) {
    return rgbaToHSL(color);
  }

  const hsl = hexToHSL(color);

  return hsl;
}

export function rgbaToHSL(string: string) {
  const matches = string.match(
    /rgba\((\d+),\s*(\d+),\s*(\d+),\s*(\d+(\.\d+)?)\)/,
  );

  if (!matches) {
    return {
      h: 0,
      s: 0,
      l: 0,
    };
  }

  const [, red, green, blue] = matches;

  const r = +red / 255;
  const g = +green / 255;
  const b = +blue / 255;

  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);

  let h = (max + min) / 2;
  let s = h;
  const l = h;

  if (max === min) {
    return { h: 0, s: 0, l };
  }

  const d = max - min;
  s = l >= 0.5 ? d / (2 - (max + min)) : d / (max + min);
  switch (max) {
    case r:
      h = ((g - b) / d + 0) * 60;
      break;
    case g:
      h = ((b - r) / d + 2) * 60;
      break;
    case b:
      h = ((r - g) / d + 4) * 60;
      break;
  }

  return {
    h: Math.round(h),
    s: Math.round(s * 100),
    l: Math.round(l * 100),
  };
}

function hexColorWithoutPound(color: string) {
  return /^[0-9A-F]{6}$/i.test(color);
}

function getComputedColor(color: string) {
  const canvas = document.createElement("canvas").getContext("2d");
  if (!canvas) return;
  canvas.fillStyle = color;
  return canvas.fillStyle;
}

export function isValidColor(color: string | undefined): boolean {
  if (!color) return false;

  const canvas = document.createElement("canvas").getContext("2d");
  if (!canvas) return false;

  canvas.fillStyle = "#000000";
  canvas.fillStyle = color;
  const result = canvas.fillStyle;

  // If the color didn't change from our test color and input isn't black-like, it's invalid
  return (
    result !== "#000000" ||
    color.toLowerCase().includes("black") ||
    color.toLowerCase() === "#000" ||
    color.toLowerCase() === "#000000" ||
    /^hsl\(0,\s*0%?,\s*0%?\)$/i.test(color) ||
    color.toLowerCase() === "rgb(0,0,0)" ||
    color.toLowerCase() === "rgba(0,0,0,1)"
  );
}

function hexToHSL(hex: string) {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);

  let h = 0;
  let s = 100;
  let l = 50;

  if (!result)
    return {
      h,
      s,
      l,
    };

  let r = parseInt(result[1], 16);
  let g = parseInt(result[2], 16);
  let b = parseInt(result[3], 16);

  r /= 255;
  g /= 255;
  b /= 255;

  const max = Math.max(r, g, b),
    min = Math.min(r, g, b);
  h = s = l = (max + min) / 2;

  if (max == min) {
    h = s = 0;
  } else {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    switch (max) {
      case r:
        h = (g - b) / d + (g < b ? 6 : 0);
        break;
      case g:
        h = (b - r) / d + 2;
        break;
      case b:
        h = (r - g) / d + 4;
        break;
    }

    h /= 6;
  }

  h = Math.round(h * 360);
  s = Math.round(s * 100);
  l = Math.round(l * 100);

  return { h, s, l };
}
