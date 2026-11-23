import {
  applyDocumentTheme,
  applyHostFonts,
  applyHostStyleVariables,
  type McpUiHostContext,
} from "@modelcontextprotocol/ext-apps";
import { updateAllPaletteColors } from "@rilldata/web-common/features/themes/palette-store";
import { themeManager } from "@rilldata/web-common/features/themes/theme-manager";

/** Fallback when the host sends no `--font-sans`; matches ChatGPT's "system fonts only" rule. */
export const SYSTEM_FONT_STACK =
  '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif';

export interface AppliedTheme {
  isDark: boolean;
  /** Resolved font stack for Vega, which draws on a canvas and cannot read CSS variables. */
  fontFamily: string;
}

/** What Rill's cached colors were last derived from; they are re-read only when this changes. */
let colorSource: { isDark: boolean; variables: unknown } | undefined;

export function resolveIsDark(
  hostContext: McpUiHostContext | undefined,
): boolean {
  if (hostContext?.theme === "dark") return true;
  if (hostContext?.theme === "light") return false;
  return (
    typeof window !== "undefined" &&
    typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-color-scheme: dark)").matches
  );
}

export function resolveFontFamily(
  root: HTMLElement = document.documentElement,
): string {
  const value = getComputedStyle(root).getPropertyValue("--font-sans").trim();
  return value || SYSTEM_FONT_STACK;
}

/**
 * Applies the host's theme to the document: `data-theme` + `color-scheme` (ext-apps),
 * Rill's `.dark` class (what app.css and the Vega theme key on), the host's CSS variables
 * and fonts. It runs on every host-context change, including resizes, so the DOM writes are
 * idempotent and the colors Rill caches from the computed CSS variables are re-read only
 * when the mode or the host variables changed.
 */
export function applyHostTheme(
  hostContext: McpUiHostContext | undefined,
): AppliedTheme {
  const isDark = resolveIsDark(hostContext);
  const root = document.documentElement;

  applyDocumentTheme(isDark ? "dark" : "light");
  root.classList.toggle("dark", isDark);

  const variables = hostContext?.styles?.variables;
  if (variables) applyHostStyleVariables(variables, root);

  // ext-apps injects the font faces once and ignores later calls.
  const fonts = hostContext?.styles?.css?.fonts;
  if (fonts) applyHostFonts(fonts);

  if (isDark !== colorSource?.isDark || variables !== colorSource?.variables) {
    colorSource = { isDark, variables };
    themeManager.clearCSSVariableCache();
    updateAllPaletteColors();
  }

  return { isDark, fontFamily: resolveFontFamily(root) };
}
