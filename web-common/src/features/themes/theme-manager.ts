/**
 * Theme Manager
 *
 * Central manager for all theme-related operations including:
 * - CSS variable resolution and caching
 * - Theme color resolution
 * - CSS variable injection
 */

import type {
  V1ThemeSpec,
  V1ThemeColors,
} from "@rilldata/web-common/runtime-client";
import chroma, { type Color } from "chroma-js";
import {
  defaultPrimaryColors,
  defaultSecondaryColors,
  TailwindColorSpacing,
} from "./color-config";
import { sanitizeThemeVariables } from "./css-sanitizer";
import {
  generatePalette,
  DEFAULT_STEP_COUNT,
  DEFAULT_GAMMA,
} from "./color-generation";

const CUSTOM_THEME_STYLE_ID = "rill-custom-theme";

class ThemeManager {
  private cssVarCache = new Map<string, string>();

  /**
   * CSS Variable Resolution
   */

  public resolveCSSVariable(
    cssVar: string,
    isDarkMode: boolean,
    fallback?: string,
  ): string {
    const cacheKey = `${cssVar}-${isDarkMode}`;

    if (this.cssVarCache.has(cacheKey)) {
      return this.cssVarCache.get(cacheKey)!;
    }

    const resolvedValue = this.resolveCSSVariableUncached(
      cssVar,
      isDarkMode,
      fallback,
    );
    this.cssVarCache.set(cacheKey, resolvedValue);

    return resolvedValue;
  }

  public clearCSSVariableCache(): void {
    this.cssVarCache.clear();
  }

  private resolveCSSVariableUncached(
    cssVar: string,
    isDarkMode: boolean,
    fallback?: string,
  ): string {
    if (typeof window === "undefined") return fallback || cssVar;

    const varName = cssVar
      .replace("var(", "")
      .replace(")", "")
      .split(",")[0]
      .trim();

    const palettePattern =
      /^--color-(theme|primary|secondary|theme-secondary)-(\d+)$/;
    const match = varName.match(palettePattern);

    if (match) {
      const [, colorType, shade] = match;
      const modeVariant = isDarkMode
        ? `--color-${colorType}-dark-${shade}`
        : `--color-${colorType}-light-${shade}`;

      const themeBoundary = document.querySelector(".dashboard-theme-boundary");
      if (themeBoundary) {
        const scopedValue = getComputedStyle(
          themeBoundary as HTMLElement,
        ).getPropertyValue(modeVariant);
        if (scopedValue && scopedValue.trim()) {
          return scopedValue.trim();
        }
      }

      const computed = getComputedStyle(
        document.documentElement,
      ).getPropertyValue(modeVariant);
      if (computed && computed.trim()) {
        return computed.trim();
      }
    }

    const themeBoundary = document.querySelector(".dashboard-theme-boundary");
    if (themeBoundary) {
      const scopedValue = getComputedStyle(
        themeBoundary as HTMLElement,
      ).getPropertyValue(varName);
      if (scopedValue && scopedValue.trim()) {
        return scopedValue.trim();
      }
    }

    const computed = getComputedStyle(
      document.documentElement,
    ).getPropertyValue(varName);
    if (computed && computed.trim()) {
      return computed.trim();
    }

    return fallback || cssVar;
  }

  /**
   * Theme Color Resolution
   */

  public resolveThemeColors(
    themeSpec: V1ThemeSpec | undefined,
    isThemeModeDark: boolean,
  ): { primary: Color; secondary: Color } {
    if (!themeSpec) {
      return {
        primary: chroma(`hsl(${defaultPrimaryColors[500]})`),
        secondary: chroma(`hsl(${defaultSecondaryColors[500]})`),
      };
    }

    const modeTheme = isThemeModeDark ? themeSpec.dark : themeSpec.light;
    const primaryColor = modeTheme?.primary;
    const secondaryColor = modeTheme?.secondary;

    return {
      primary: primaryColor
        ? chroma(primaryColor)
        : chroma(`hsl(${defaultPrimaryColors[500]})`),
      secondary: secondaryColor
        ? chroma(secondaryColor)
        : chroma(`hsl(${defaultSecondaryColors[500]})`),
    };
  }

  public resolveThemeObject(
    themeSpec: V1ThemeSpec | undefined,
    isThemeModeDark: boolean,
  ): Record<string, string> | undefined {
    if (!themeSpec) return undefined;

    const modeTheme = isThemeModeDark ? themeSpec.dark : themeSpec.light;
    if (!modeTheme) return undefined;

    const merged: Record<string, string> = {};

    if (modeTheme.variables) {
      Object.assign(merged, modeTheme.variables);
    }

    if (modeTheme.primary) {
      merged.primary = modeTheme.primary;
    }

    if (modeTheme.secondary) {
      merged.secondary = modeTheme.secondary;
    }

    return Object.keys(merged).length > 0 ? merged : undefined;
  }

  /**
   * CSS Variable Injection
   */

  public applyTheme(
    theme: V1ThemeSpec | undefined,
    currentModeTheme: V1ThemeColors | undefined,
    root: HTMLElement,
  ): void {
    const hasCurrentModeTheme = Boolean(
      currentModeTheme?.variables ||
        currentModeTheme?.primary ||
        currentModeTheme?.secondary,
    );

    if (hasCurrentModeTheme && theme && currentModeTheme) {
      this.injectCurrentModeThemeVariables(currentModeTheme, root);
      this.handleCurrentModePrimarySecondary(currentModeTheme, root);
      this.clearCSSVariableCache();
    }
  }

  public clearTheme(root: HTMLElement): void {
    this.removeExistingCustomCSS();

    if (root !== document.documentElement) {
      TailwindColorSpacing.forEach((_, i) => {
        const spacing = TailwindColorSpacing[i];
        ["theme", "primary", "theme-secondary", "secondary"].forEach((type) => {
          ["light", "dark"].forEach((mode) => {
            root.style.removeProperty(`--color-${type}-${mode}-${spacing}`);
          });
        });
      });
    }
  }

  private injectCurrentModeThemeVariables(
    currentModeTheme: V1ThemeColors,
    scopeElement: HTMLElement,
  ): void {
    const vars = currentModeTheme.variables;
    if (!vars || typeof vars !== "object") return;

    const variables = sanitizeThemeVariables(vars);
    if (Object.keys(variables).length === 0) return;

    const scopeSelector =
      scopeElement === document.documentElement
        ? undefined
        : ".dashboard-theme-boundary";

    let css = "";
    const selector = scopeSelector || ":root";
    css += `${selector} {\n`;
    for (const [name, value] of Object.entries(variables)) {
      css += `  ${name}: ${value};\n`;
    }
    css += "}\n";

    this.createAndInjectStyle(css);
  }

  private handleCurrentModePrimarySecondary(
    currentModeTheme: V1ThemeColors,
    root: HTMLElement,
  ): void {
    const primaryColor = currentModeTheme.primary;
    if (primaryColor && typeof primaryColor === "string") {
      try {
        const palette = generatePalette(
          chroma(primaryColor),
          false,
          DEFAULT_STEP_COUNT,
          DEFAULT_GAMMA,
        );

        this.setColorVariables(root, "theme", "light", palette.light);
        this.setColorVariables(root, "theme", "dark", palette.dark);
        this.setColorVariables(root, "primary", "light", palette.light);
        this.setColorVariables(root, "primary", "dark", palette.dark);
        this.setIntermediateVariables(root, "theme");
        this.setIntermediateVariables(root, "primary");
      } catch (error) {
        console.error("Failed to generate palette from primary color:", error);
      }
    }

    const secondaryColor = currentModeTheme.secondary;
    if (secondaryColor && typeof secondaryColor === "string") {
      try {
        const palette = generatePalette(
          chroma(secondaryColor),
          false,
          DEFAULT_STEP_COUNT,
          DEFAULT_GAMMA,
        );

        this.setColorVariables(root, "theme-secondary", "light", palette.light);
        this.setColorVariables(root, "theme-secondary", "dark", palette.dark);
        this.setColorVariables(root, "secondary", "light", palette.light);
        this.setColorVariables(root, "secondary", "dark", palette.dark);
        this.setIntermediateVariables(root, "theme-secondary");
        this.setIntermediateVariables(root, "secondary");
      } catch (error) {
        console.error(
          "Failed to generate palette from secondary color:",
          error,
        );
      }
    }
  }

  private setColorVariables(
    root: HTMLElement,
    type: string,
    mode: "dark" | "light",
    colors?: Color[],
  ): void {
    if (!colors) {
      if (root !== document.documentElement) {
        TailwindColorSpacing.forEach((_, i) => {
          root.style.removeProperty(
            `--color-${type}-${mode}-${TailwindColorSpacing[i]}`,
          );
        });
      }
    } else {
      colors.forEach((color, i) => {
        root.style.setProperty(
          `--color-${type}-${mode}-${TailwindColorSpacing[i]}`,
          color.css("hsl"),
        );
      });
    }
  }

  private setIntermediateVariables(root: HTMLElement, type: string): void {
    TailwindColorSpacing.forEach((spacing) => {
      root.style.setProperty(
        `--color-${type}-${spacing}`,
        `light-dark(var(--color-${type}-light-${spacing}), var(--color-${type}-dark-${spacing}))`,
      );
    });
  }

  private removeExistingCustomCSS(): void {
    const existingStyle = document.getElementById(CUSTOM_THEME_STYLE_ID);
    if (existingStyle) {
      existingStyle.remove();
    }
  }

  private createAndInjectStyle(css: string): void {
    const style = document.createElement("style");
    style.id = CUSTOM_THEME_STYLE_ID;
    style.textContent = css;
    document.head.appendChild(style);
  }
}

export const themeManager = new ThemeManager();

if (typeof window !== "undefined") {
  (window as any).__clearRillCSSCache = () =>
    themeManager.clearCSSVariableCache();
}
