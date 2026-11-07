import { hasContext } from "svelte";

/**
 * Context key under which ThemeProvider exposes its theme store.
 */
export const THEME_STORE_CONTEXT_KEY = Symbol("theme-store");

/**
 * Returns the theme boundary class when called from a component rendered
 * inside a ThemeProvider, and undefined otherwise.
 *
 * Dashboard theme CSS variables are scoped to the `dashboard-theme-boundary`
 * class (see Theme.generateCSS). Floating content such as dropdowns and
 * popovers is portalled to document.body, outside the ThemeProvider's DOM
 * subtree, so it must carry the class itself to pick up the dashboard theme.
 * Svelte context follows the component tree rather than the DOM, so it
 * identifies dashboard content even after portalling.
 *
 * Must be called during component initialization.
 */
export function getThemeBoundaryClass(): string | undefined {
  return hasContext(THEME_STORE_CONTEXT_KEY)
    ? "dashboard-theme-boundary"
    : undefined;
}
