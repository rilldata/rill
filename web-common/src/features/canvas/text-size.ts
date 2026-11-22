/**
 * Text size presets for canvas components, set with `text_size` on a component
 * or on the canvas, which is then the default for its components.
 * The runtime validates the same list; see `canvasTextSizes` in runtime/parser/parse_canvas.go.
 */
export const TEXT_SIZES = ["small", "medium", "large", "xlarge"] as const;

export type TextSize = (typeof TEXT_SIZES)[number];

export const DEFAULT_TEXT_SIZE: TextSize = "medium";

// A preset multiplies the font sizes a component would otherwise draw with,
// whether those are Rill's defaults or the theme's. It sets no sizes of its own.
const TEXT_SCALES: Record<TextSize, number> = {
  small: 0.875,
  medium: 1,
  large: 1.25,
  xlarge: 1.5,
};

export function isTextSize(value: unknown): value is TextSize {
  return TEXT_SIZES.includes(value as TextSize);
}

/**
 * The font size multiplier of a component: its own preset, else the canvas's,
 * else medium. Values that are not a preset are skipped, such as a draft the
 * runtime has not validated.
 */
export function resolveTextScale(
  componentTextSize: unknown,
  canvasTextSize: unknown,
): number {
  if (isTextSize(componentTextSize)) return TEXT_SCALES[componentTextSize];
  if (isTextSize(canvasTextSize)) return TEXT_SCALES[canvasTextSize];
  return TEXT_SCALES[DEFAULT_TEXT_SIZE];
}

/**
 * Value for the `--text-scale` custom property that the components' styles
 * multiply their font sizes by. Undefined at the default, so that the
 * property is not set and the styles use their fallback of 1.
 */
export function textScaleStyle(scale: number): string | undefined {
  return scale === 1 ? undefined : String(scale);
}
