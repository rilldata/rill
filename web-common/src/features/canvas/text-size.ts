import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { ComponentInputParam } from "./inspector/types";

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

/**
 * The presets as select options, smallest first.
 * Built per call, so that the labels follow the active locale.
 */
export function textSizeOptions(): { value: TextSize; label: string }[] {
  return [
    { value: "small", label: m.canvas_text_size_small() },
    { value: "medium", label: m.canvas_text_size_medium() },
    { value: "large", label: m.canvas_text_size_large() },
    { value: "xlarge", label: m.canvas_text_size_xlarge() },
  ];
}

/**
 * The "Text size" select of a component's inspector. Its first choice has an
 * empty value, which removes the key, so the component follows the canvas.
 */
export function getTextSizeOption(): ComponentInputParam {
  return {
    type: "select",
    optional: true,
    label: m.canvas_text_size_label(),
    meta: {
      default: "",
      placeholder: m.canvas_text_size_canvas_default(),
      options: [
        { value: "", label: m.canvas_text_size_canvas_default() },
        ...textSizeOptions(),
      ],
    },
  };
}
