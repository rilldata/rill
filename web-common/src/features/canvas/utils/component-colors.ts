import type {
  V1CanvasItem,
  V1ThemeColors,
  V1ThemeSpec,
} from "@rilldata/web-common/runtime-client";
import { themeManager } from "@rilldata/web-common/features/themes/theme-manager";

/**
 * Extracts theme overrides from a canvas item.
 * Returns ThemeColors for light and dark modes.
 */
export function getComponentThemeOverrides(
  item: V1CanvasItem | undefined,
): { light?: V1ThemeColors; dark?: V1ThemeColors } {
  if (!item) {
    return {};
  }

  return {
    light: item.lightThemeOverride || undefined,
    dark: item.darkThemeOverride || undefined,
  };
}

/**
 * Merges component theme overrides with global theme.
 * Component overrides take precedence.
 * Returns a Record<string, string> suitable for chart consumption.
 */
export function mergeComponentTheme(
  globalTheme: V1ThemeSpec | undefined,
  componentOverrides: { light?: V1ThemeColors; dark?: V1ThemeColors },
  isDarkMode: boolean,
): Record<string, string> | undefined {
  // Start with global theme for current mode
  const globalThemeObject = themeManager.resolveThemeObject(
    globalTheme,
    isDarkMode,
  );

  // Get component override for current mode
  const componentOverride = isDarkMode
    ? componentOverrides.dark
    : componentOverrides.light;

  if (!componentOverride && !globalThemeObject) {
    return undefined;
  }

  // Merge: start with global, overlay component (component takes precedence)
  const merged: Record<string, string> = { ...globalThemeObject };

  if (componentOverride) {
    // Add primary and secondary if present
    if (componentOverride.primary) {
      merged.primary = componentOverride.primary;
    }
    if (componentOverride.secondary) {
      merged.secondary = componentOverride.secondary;
    }

    // Add variables (component variables override global)
    if (componentOverride.variables) {
      Object.assign(merged, componentOverride.variables);
    }
  }

  return Object.keys(merged).length > 0 ? merged : undefined;
}

/**
 * Generates scoped CSS variables for component theme overrides.
 * Returns CSS string to inject in <style> tag.
 * 
 * Maps theme variables to CSS custom properties.
 * Primary and secondary are handled via theme object for charts, not CSS variables.
 * Other variables map directly to CSS variables (background, surface, card, foreground, border, etc.)
 */
export function generateComponentThemeCSS(
  componentId: string,
  overrides: { light?: V1ThemeColors; dark?: V1ThemeColors },
): string {
  if (!overrides.light && !overrides.dark) {
    return "";
  }

  const cssRules: string[] = [];

  // Generate light mode CSS variables
  if (overrides.light) {
    const lightVars: string[] = [];

    // Add primary and secondary if present
    if (overrides.light.primary) {
      lightVars.push(`    --color-theme-500: ${overrides.light.primary};`);
      lightVars.push(`    --color-theme-600: ${overrides.light.primary};`);
    }
    if (overrides.light.secondary) {
      lightVars.push(`    --color-theme-secondary: ${overrides.light.secondary};`);
    }

    // Add variables (direct mapping to CSS variables)
    if (overrides.light.variables) {
      for (const [key, value] of Object.entries(overrides.light.variables)) {
        // Map theme variable names to CSS variable names
        const cssVarName = mapThemeVariableToCSS(key);
        lightVars.push(`    ${cssVarName}: ${value};`);
      }
    }

    if (lightVars.length > 0) {
      cssRules.push(`#${componentId} {\n${lightVars.join("\n")}\n  }`);
    }
  }

  // Generate dark mode CSS variables
  if (overrides.dark) {
    const darkVars: string[] = [];

    // Add primary and secondary if present
    if (overrides.dark.primary) {
      darkVars.push(`    --color-theme-500: ${overrides.dark.primary};`);
      darkVars.push(`    --color-theme-600: ${overrides.dark.primary};`);
    }
    if (overrides.dark.secondary) {
      darkVars.push(`    --color-theme-secondary: ${overrides.dark.secondary};`);
    }

    // Add variables
    if (overrides.dark.variables) {
      for (const [key, value] of Object.entries(overrides.dark.variables)) {
        const cssVarName = mapThemeVariableToCSS(key);
        darkVars.push(`    ${cssVarName}: ${value};`);
      }
    }

    if (darkVars.length > 0) {
      cssRules.push(`:root.dark #${componentId} {\n${darkVars.join("\n")}\n  }`);
    }
  }

  return cssRules.join("\n\n");
}

/**
 * Maps theme variable names to CSS custom property names.
 * Most variables map directly, but some have special mappings.
 */
function mapThemeVariableToCSS(themeVar: string): string {
  // Direct mappings for common theme variables
  const directMappings: Record<string, string> = {
    background: "--background",
    surface: "--surface",
    card: "--card",
    "card-foreground": "--card-foreground",
    foreground: "--foreground",
    border: "--border",
    input: "--input",
    ring: "--ring",
    popover: "--popover",
    "popover-foreground": "--popover-foreground",
    primary: "--primary",
    "primary-foreground": "--primary-foreground",
    secondary: "--secondary",
    "secondary-foreground": "--secondary-foreground",
    muted: "--muted",
    "muted-foreground": "--muted-foreground",
    accent: "--accent",
    "accent-foreground": "--accent-foreground",
    destructive: "--destructive",
    "destructive-foreground": "--destructive-foreground",
  };

  // Return direct mapping if available, otherwise use theme variable name as-is
  return directMappings[themeVar] || `--${themeVar}`;
}
