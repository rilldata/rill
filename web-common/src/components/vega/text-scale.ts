import type { Config } from "vega-lite";

// `fontSize` on marks, and `labelFontSize`, `titleFontSize` and
// `subtitleFontSize` on axes, legends, headers and titles.
const FONT_SIZE_KEY = /^(fontSize|[a-z]+FontSize)$/;

// Subtrees that hold rows rather than chart definitions: a `fontSize` key in
// there is a column of the data.
const DATA_KEYS = new Set(["data", "datasets"]);

// The sizes Vega and Vega-Lite fall back to for the roles Rill's chart config
// leaves unset. A scaled config states them, so that these roles scale too.
const VEGA_DEFAULT_TEXT_MARK_FONT_SIZE = 11;
const VEGA_DEFAULT_TITLE_FONT_SIZE = 13;
const VEGA_DEFAULT_SUBTITLE_FONT_SIZE = 12;

/**
 * Multiplies every font size in a Vega-Lite spec or config by `scale`, and
 * returns the input itself when the scale is 1.
 *
 * It multiplies whatever is there (Rill's defaults, a theme's sizes, a chart's
 * own config or a hand-written spec), so a text size preset composes with
 * them instead of replacing them.
 *
 * A size given as an expression is left alone: it is computed from the
 * chart's dimensions to fit a space, such as the hole of a donut, and
 * multiplying it would overflow that space. The spec that writes such a size
 * decides how it follows the scale.
 */
export function scaleVegaText<T>(value: T, scale: number): T {
  if (scale === 1 || value === null || typeof value !== "object") return value;
  if (Array.isArray(value)) {
    return value.map((item: unknown) => scaleVegaText(item, scale)) as T;
  }
  return Object.fromEntries(
    Object.entries(value).map(([key, child]: [string, unknown]) => {
      if (DATA_KEYS.has(key)) return [key, child];
      if (FONT_SIZE_KEY.test(key)) return [key, scaleFontSize(child, scale)];
      return [key, scaleVegaText(child, scale)];
    }),
  ) as T;
}

/**
 * A chart config with its font sizes multiplied by `scale`, including the
 * ones it leaves to Vega's defaults.
 */
export function scaleConfigText(config: Config, scale: number): Config {
  if (scale === 1) return config;
  return scaleVegaText(
    {
      ...config,
      text: { fontSize: VEGA_DEFAULT_TEXT_MARK_FONT_SIZE, ...config.text },
      title: {
        fontSize: VEGA_DEFAULT_TITLE_FONT_SIZE,
        subtitleFontSize: VEGA_DEFAULT_SUBTITLE_FONT_SIZE,
        ...config.title,
      },
    },
    scale,
  );
}

function scaleFontSize(size: unknown, scale: number): unknown {
  if (typeof size === "number") return roundSize(size * scale);
  if (size === null || typeof size !== "object") return size;
  if (Array.isArray(size)) {
    return size.map((item: unknown) => scaleFontSize(item, scale));
  }

  // A conditional size: `{ condition: { test, value }, value }`.
  const scaled: Record<string, unknown> = { ...size };
  if ("value" in scaled) scaled.value = scaleFontSize(scaled.value, scale);
  if ("condition" in scaled) {
    scaled.condition = scaleFontSize(scaled.condition, scale);
  }
  return scaled;
}

// Keeps float noise such as 13.750000000000002 out of the spec.
function roundSize(size: number): number {
  return Math.round(size * 100) / 100;
}
