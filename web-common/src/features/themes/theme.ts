import type {
  V1ThemeColors,
  V1ThemeSpec,
} from "@rilldata/web-common/runtime-client";
import { type Color } from "chroma-js";
import { generateColorPalette } from "./palette-generator";
import { TailwindColorSpacing } from "./color-config";
import { primary } from "./colors";
import { getChroma, resolveThemeObject } from "./theme-utils";

export class Theme {
  colors: { light: Colors; dark: Colors };
  spec: V1ThemeSpec;
  css: string;
  resolvedThemeObject: {
    dark: Record<string, string>;
    light: Record<string, string>;
  };

  constructor(spec: V1ThemeSpec | undefined) {
    if (spec) {
      this.processThemeSpec(spec);
    }
  }

  getColor(name: string, dark: boolean, fallback: string = "500") {
    const color = this.colors?.[dark ? "dark" : "light"]?.[name]?.css("hsl");
    if (!color) {
      return primary[fallback].css("hsl");
    }
    return color;
  }

  processThemeSpec(spec: V1ThemeSpec) {
    this.spec = spec;
    this.colors = this.processTheme(spec);
    this.css = this.generateCSS();

    // Compatibility with current implementation, this can eventually be removed
    this.resolvedThemeObject = {
      dark: resolveThemeObject(spec, true) ?? {},
      light: resolveThemeObject(spec, false) ?? {},
    };
  }

  private stringifyVars(vars: Record<string, Color | undefined>) {
    return Object.entries(vars)
      .map(([k, v]) => `--${k}: ${v?.css("hsl") ?? "unset"};`)
      .join("\n  ");
  }

  private generateCSS(): string {
    const darkColors: Record<string, Color | undefined> = this.colors.dark;
    const lightColors = this.colors.light;

    for (const [k] of Object.entries(lightColors)) {
      if (!(k in darkColors)) {
        darkColors[k] = undefined;
      }
    }

    return `
 .dashboard-theme-boundary {
  ${this.stringifyVars(lightColors)}
}
.dark .dashboard-theme-boundary {
  ${this.stringifyVars(darkColors)}
}`.trim();
  }

  private processTheme(spec: V1ThemeSpec) {
    // Handle legacy theme format (colors: primary/secondary)
    // If neither light nor dark is defined, but we have legacy color fields,
    // treat them as light mode colors only for backwards compatibility
    const hasLegacyColors =
      !spec.light &&
      !spec.dark &&
      (spec.primaryColorRaw || spec.secondaryColorRaw);

    if (hasLegacyColors) {
      const legacyColors: V1ThemeColors = {
        primary: spec.primaryColorRaw,
        secondary: spec.secondaryColorRaw,
      };
      const lightColors = this.processColors(legacyColors);
      return { dark: {}, light: lightColors };
    }

    const darkColors = this.processColors(spec.dark ?? {});
    const lightColors = this.processColors(spec.light ?? {});

    return { dark: darkColors, light: lightColors };
  }

  private processColors(colors: V1ThemeColors) {
    const finalColors: Colors = {};
    const { primary, secondary, variables } = colors;

    if (primary) {
      const primaryReference = getChroma(primary);
      const primaryPalette = generateColorPalette(primaryReference);
      for (const [i, color] of primaryPalette.entries()) {
        finalColors[`color-theme-${TailwindColorSpacing[i]}`] = color;
      }
      finalColors.primary = primaryReference;
    }

    if (secondary) {
      const secondaryReference = getChroma(secondary);
      const secondaryPalette = generateColorPalette(secondaryReference);
      for (const [i, color] of secondaryPalette.entries()) {
        finalColors[`color-theme-secondary-${TailwindColorSpacing[i]}`] = color;
      }
      finalColors.secondary = secondaryReference;
    }

    for (const [k, v] of Object.entries(variables ?? {})) {
      if (!v) continue;
      finalColors[k] = getChroma(v);
    }

    return finalColors;
  }
}

// Needs refinement and better generated types from the backend
type Colors = {
  primary?: Color;
  secondary?: Color;

  ring?: Color;
  radius?: Color;
  surface?: Color;
  background?: Color;
  foreground?: Color;

  card?: Color;
  "card-foreground"?: Color;
  popover?: Color;
  "popover-foreground"?: Color;
  "primary-foreground"?: Color;
  "secondary-foreground"?: Color;
  muted?: Color;
  "muted-foreground"?: Color;
  accent?: Color;
  "accent-foreground"?: Color;
  destructive?: Color;
  "destructive-foreground"?: Color;
  border?: Color;
  input?: Color;

  "color-sequential-1"?: Color;
  "color-sequential-2"?: Color;
  "color-sequential-3"?: Color;
  "color-sequential-4"?: Color;
  "color-sequential-5"?: Color;
  "color-sequential-6"?: Color;
  "color-sequential-7"?: Color;
  "color-sequential-8"?: Color;
  "color-sequential-9"?: Color;

  "color-diverging-1"?: Color;
  "color-diverging-2"?: Color;
  "color-diverging-3"?: Color;
  "color-diverging-4"?: Color;
  "color-diverging-5"?: Color;
  "color-diverging-6"?: Color;
  "color-diverging-7"?: Color;
  "color-diverging-8"?: Color;
  "color-diverging-9"?: Color;
  "color-diverging-10"?: Color;
  "color-diverging-11"?: Color;

  "color-qualitative-1"?: Color;
  "color-qualitative-2"?: Color;
  "color-qualitative-3"?: Color;
  "color-qualitative-4"?: Color;
  "color-qualitative-5"?: Color;
  "color-qualitative-6"?: Color;
  "color-qualitative-7"?: Color;
  "color-qualitative-8"?: Color;
  "color-qualitative-9"?: Color;
  "color-qualitative-10"?: Color;
  "color-qualitative-11"?: Color;
  "color-qualitative-12"?: Color;
  "color-qualitative-13"?: Color;
  "color-qualitative-14"?: Color;
  "color-qualitative-15"?: Color;
  "color-qualitative-16"?: Color;
  "color-qualitative-17"?: Color;
  "color-qualitative-18"?: Color;
  "color-qualitative-19"?: Color;
  "color-qualitative-20"?: Color;
  "color-qualitative-21"?: Color;
  "color-qualitative-22"?: Color;
  "color-qualitative-23"?: Color;
  "color-qualitative-24"?: Color;
};
