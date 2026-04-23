import type {
  V1ThemeSpec,
  V1ThemeColors,
} from "@rilldata/web-common/runtime-client";
import { stringify } from "yaml";

/**
 * Serializes a V1ThemeSpec to a YAML string matching the format
 * expected by runtime/parser/parse_theme.go.
 */
export function buildThemeYaml(spec: V1ThemeSpec): string {
  const doc: Record<string, unknown> = { type: "theme" };

  const light = flattenThemeColors(spec.light);
  if (light && Object.keys(light).length > 0) {
    doc.light = light;
  }

  const dark = flattenThemeColors(spec.dark);
  if (dark && Object.keys(dark).length > 0) {
    doc.dark = dark;
  }

  return stringify(doc, { lineWidth: 0 });
}

/**
 * Flattens V1ThemeColors into a plain object with primary/secondary at top
 * level and variables merged in (the format the YAML parser expects).
 */
function flattenThemeColors(
  colors: V1ThemeColors | undefined,
): Record<string, string> | undefined {
  if (!colors) return undefined;

  const result: Record<string, string> = {};
  if (colors.primary) result.primary = colors.primary;
  if (colors.secondary) result.secondary = colors.secondary;
  if (colors.variables) {
    Object.assign(result, colors.variables);
  }
  return Object.keys(result).length > 0 ? result : undefined;
}

/**
 * Converts a V1ThemeSpec into a plain object suitable for inline embedding
 * in an explore/canvas YAML file (e.g. `parsedDocument.set("theme", obj)`).
 */
export function buildInlineThemeObject(
  spec: V1ThemeSpec,
): Record<string, unknown> {
  const result: Record<string, unknown> = {};

  const light = flattenThemeColors(spec.light);
  if (light && Object.keys(light).length > 0) result.light = light;

  const dark = flattenThemeColors(spec.dark);
  if (dark && Object.keys(dark).length > 0) result.dark = dark;

  return result;
}
