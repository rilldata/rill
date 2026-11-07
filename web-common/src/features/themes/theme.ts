import type {
  V1ThemeColors,
  V1ThemeSpec,
} from "@rilldata/web-common/runtime-client";
import { type Color } from "chroma-js";
import { generateColorPalette } from "./palette-generator";
import { TailwindColorSpacing } from "./color-config";
import { primary } from "./colors";
import { getChroma, resolveThemeObject } from "./theme-utils";
import { createDarkVariation } from "./color-generation";

type Shade = (typeof TailwindColorSpacing)[number];

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

  /**
   * Semantic tokens that app.css derives from Rill's own primary/secondary palettes.
   *
   * Those declarations live on :root, and a custom property's var() references are
   * substituted at the element that declares it. So they always resolve against the
   * default Rill palette, and a dashboard theme -- which is scoped to the
   * .dashboard-theme-boundary element further down the tree -- can never influence them.
   * The result is Rill's indigo leaking into hover, selection, focus rings, and dimension
   * and measure chips on a rethemed dashboard.
   *
   * Re-declaring them on the boundary from the theme's own palettes fixes that. The shades
   * below mirror app.css exactly, so only the hue source changes, not the design.
   *
   * `light` is the shade :root uses, `dark` the shade :root.dark uses. Both read from the
   * un-darkened palette of the mode's own primary, because :root.dark also picks
   * light-palette shades (--color-primary-light-*) -- they stay legible against dark
   * surfaces. A token with no `dark` shade is not palette-derived in dark mode (dark
   * surfaces are neutral grays) and is left to app.css.
   */
  private static PALETTE_DERIVED_TOKENS: {
    token: keyof Colors;
    source: "primary" | "secondary";
    light: Shade;
    dark?: Shade;
  }[] = [
    { token: "surface-hover", source: "primary", light: 50 },
    { token: "surface-active", source: "primary", light: 100 },
    { token: "fg-accent", source: "primary", light: 900, dark: 100 },
    { token: "accent-primary", source: "primary", light: 500, dark: 600 },
    {
      token: "accent-primary-action",
      source: "primary",
      light: 500,
      dark: 400,
    },
    { token: "accent-secondary", source: "primary", light: 500, dark: 700 },
    {
      token: "accent-secondary-action",
      source: "secondary",
      light: 500,
      dark: 400,
    },
    { token: "icon-accent", source: "primary", light: 500, dark: 400 },
    { token: "ring-focus", source: "primary", light: 300, dark: 500 },
    { token: "dimension", source: "primary", light: 100, dark: 950 },
    { token: "dimension-foreground", source: "primary", light: 800, dark: 100 },
    { token: "dimension-border", source: "primary", light: 800, dark: 800 },
    { token: "measure", source: "secondary", light: 100, dark: 950 },
    { token: "measure-foreground", source: "secondary", light: 800, dark: 100 },
    { token: "measure-border", source: "secondary", light: 800, dark: 800 },
  ];

  // Opacity percentages for fg-* hierarchy when auto-generating from fg-primary
  private static FG_OPACITY_HIERARCHY: Record<string, number> = {
    "fg-secondary": 85,
    "fg-tertiary": 70,
    "fg-muted": 55,
    "fg-disabled": 40,
    "fg-inverse": 100,
    "fg-accent": 100,
  };

  private stringifyVars(vars: Record<string, Color | undefined>): string {
    const lines: string[] = [];

    for (const [k, v] of Object.entries(vars)) {
      lines.push(`--${k}: ${v?.css("hsl") ?? "unset"};`);
    }

    // Generate fg-* hierarchy from fg-primary using color-mix
    const fgPrimary = vars["fg-primary"];
    if (fgPrimary) {
      const fgPrimaryHsl = fgPrimary.css("hsl");
      for (const [fgVar, opacity] of Object.entries(
        Theme.FG_OPACITY_HIERARCHY,
      )) {
        // Only generate if not explicitly set
        if (!(fgVar in vars)) {
          if (opacity === 100) {
            lines.push(`--${fgVar}: ${fgPrimaryHsl};`);
          } else {
            lines.push(
              `--${fgVar}: color-mix(in oklab, ${fgPrimaryHsl} ${opacity}%, transparent);`,
            );
          }
        }
      }
    }

    return lines.join("\n  ");
  }

  private generateCSS(): string {
    const darkColors: Record<string, Color | undefined> = this.colors.dark;
    const lightColors = this.colors.light;

    for (const [k] of Object.entries(lightColors)) {
      if (k in darkColors) continue;

      // A light-only token is reset to `unset` in dark mode, so the boundary inherits
      // :root.dark instead of keeping the light value. The fg-* hierarchy is the exception:
      // stringifyVars only fills a slot that isn't already a key, so an `unset` placeholder
      // would suppress the value dark's own fg-primary should produce.
      if (darkColors["fg-primary"] && k in Theme.FG_OPACITY_HIERARCHY) continue;

      darkColors[k] = undefined;
    }

    const css = `
 .dashboard-theme-boundary {
  ${this.stringifyVars(lightColors)}
}
.dark .dashboard-theme-boundary {
  ${this.stringifyVars(darkColors)}
}`.trim();

    return css;
  }

  private processTheme(spec: V1ThemeSpec) {
    // Handle legacy theme format (colors: primary/secondary)
    // If neither light nor dark is defined, but we have legacy color fields,
    // generate both light and dark palettes from them
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
      const darkColors = this.processColors(legacyColors, true);
      return { dark: darkColors, light: lightColors };
    }

    const darkColors = this.processColors(spec.dark ?? {}, true);
    const lightColors = this.processColors(spec.light ?? {});

    return { dark: darkColors, light: lightColors };
  }

  private processColors(colors: V1ThemeColors, dark: boolean = false) {
    const finalColors: Colors = {};
    const { primary, secondary, variables } = colors;

    let primaryPalette: Color[] | undefined;
    let secondaryPalette: Color[] | undefined;

    if (primary) {
      const primaryReference = getChroma(primary);
      primaryPalette = generateColorPalette(primaryReference);
      const finalColorPalette = dark
        ? createDarkVariation(primaryPalette)
        : primaryPalette;
      for (const [i, color] of finalColorPalette.entries()) {
        finalColors[`color-theme-${TailwindColorSpacing[i]}`] = color;
        // Alias Rill's own palette to the theme's inside the boundary, so the ~300
        // `bg-primary-500`-style utilities scattered through dashboard components follow
        // the theme too. Only shadowed within the boundary: the rest of the app -- nav,
        // file explorer, admin chrome -- keeps Rill's brand colors.
        finalColors[`color-primary-${TailwindColorSpacing[i]}`] = color;
      }
      finalColors.primary = primaryReference;
    }

    if (secondary) {
      const secondaryReference = getChroma(secondary);
      secondaryPalette = generateColorPalette(secondaryReference);
      const finalColorPalette = dark
        ? createDarkVariation(secondaryPalette)
        : secondaryPalette;
      for (const [i, color] of finalColorPalette.entries()) {
        finalColors[`color-theme-secondary-${TailwindColorSpacing[i]}`] = color;
        finalColors[`color-secondary-${TailwindColorSpacing[i]}`] = color;
      }
      finalColors.secondary = secondaryReference;
    }

    for (const [k, v] of Object.entries(variables ?? {})) {
      if (!v) continue;
      finalColors[k] = getChroma(v);
    }

    this.applyPaletteDerivedTokens(
      finalColors,
      { primary: primaryPalette, secondary: secondaryPalette },
      dark,
    );

    return finalColors;
  }

  /**
   * Fills in the semantic tokens listed in PALETTE_DERIVED_TOKENS from the theme's
   * palettes. Tokens the theme sets explicitly are left untouched, as are tokens whose
   * source palette the theme doesn't define: a theme with only a primary color keeps
   * Rill's defaults for the secondary-derived tokens.
   */
  private applyPaletteDerivedTokens(
    finalColors: Colors,
    palettes: { primary?: Color[]; secondary?: Color[] },
    dark: boolean,
  ): void {
    for (const {
      token,
      source,
      light,
      dark: darkShade,
    } of Theme.PALETTE_DERIVED_TOKENS) {
      if (token in finalColors) continue;

      // fg-accent is part of the fg-* hierarchy generated from fg-primary in stringifyVars.
      // A theme that sets its own fg-primary owns fg-accent too, so leave it alone.
      if (token === "fg-accent" && "fg-primary" in finalColors) continue;

      const shade = dark ? darkShade : light;
      if (shade === undefined) continue;

      const color = palettes[source]?.[TailwindColorSpacing.indexOf(shade)];
      if (color) finalColors[token] = color;
    }

    // popover-accent follows surface-hover, mirroring app.css.
    if (!("popover-accent" in finalColors) && finalColors["surface-hover"]) {
      finalColors["popover-accent"] = finalColors["surface-hover"];
    }
  }
}

// Theme color variables - includes both new semantic names and deprecated names for backwards compatibility
type Colors = {
  primary?: Color;
  secondary?: Color;

  // Surface semantic variables
  "surface-base"?: Color;
  "surface-background"?: Color;
  "surface-hover"?: Color;
  "surface-active"?: Color;
  "surface-overlay"?: Color;
  "surface-subtle"?: Color;
  "surface-muted"?: Color;
  "surface-card"?: Color;

  // Foreground semantic variables
  "fg-primary"?: Color;
  "fg-secondary"?: Color;
  "fg-tertiary"?: Color;
  "fg-inverse"?: Color;
  "fg-muted"?: Color;
  "fg-disabled"?: Color;
  "fg-accent"?: Color;

  // Accent semantic variables
  "accent-primary"?: Color;
  "accent-primary-action"?: Color;
  "accent-secondary"?: Color;
  "accent-secondary-action"?: Color;

  // Icon semantic variables
  "icon-default"?: Color;
  "icon-muted"?: Color;
  "icon-disabled"?: Color;
  "icon-accent"?: Color;

  // Border, input, and radius
  border?: Color;
  input?: Color;
  radius?: Color;

  // Ring (focus states)
  "ring-focus"?: Color;
  "ring-offset"?: Color;

  // Dimension styling
  dimension?: Color;
  "dimension-foreground"?: Color;
  "dimension-border"?: Color;

  // Measure styling
  measure?: Color;
  "measure-foreground"?: Color;
  "measure-border"?: Color;

  // Tooltip
  tooltip?: Color;

  // Destructive actions
  destructive?: Color;
  "destructive-foreground"?: Color;

  // KPI delta coloring
  "kpi-positive"?: Color;
  "kpi-negative"?: Color;

  // Popover
  popover?: Color;
  "popover-accent"?: Color;
  "popover-foreground"?: Color;
  "popover-footer"?: Color;

  // Sequential palette (9 colors)
  "color-sequential-1"?: Color;
  "color-sequential-2"?: Color;
  "color-sequential-3"?: Color;
  "color-sequential-4"?: Color;
  "color-sequential-5"?: Color;
  "color-sequential-6"?: Color;
  "color-sequential-7"?: Color;
  "color-sequential-8"?: Color;
  "color-sequential-9"?: Color;

  // Diverging palette (11 colors)
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

  // Qualitative palette (24 colors)
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
