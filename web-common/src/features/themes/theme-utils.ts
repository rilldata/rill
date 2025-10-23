/**
 * Theme Utilities
 *
 * Re-exports theme resolution functions from the centralized theme manager.
 * Maintained for backwards compatibility.
 */

import { themeManager } from "./theme-manager";

export const resolveThemeColors =
  themeManager.resolveThemeColors.bind(themeManager);
export const resolveThemeObject =
  themeManager.resolveThemeObject.bind(themeManager);
