import type {
  V1ThemeColors,
  V1ThemeSpec,
} from "@rilldata/web-common/runtime-client";
import chroma, { type Color } from "chroma-js";
import { generateColorPalette } from "./palette-generator";
import { TailwindColorSpacing } from "./color-config";
import { writable, type Writable } from "svelte/store";
import { primary } from "./colors";
import { resolveThemeObject } from "./theme-utils";

type Colors = Record<string, Color>;

export class Theme {
  colors: { light: Colors; dark: Colors };
  css = writable("");
  resolvedThemeObject: Writable<
    { dark: Record<string, string>; light: Record<string, string> } | undefined
  > = writable(undefined);

  constructor(
    private scope: string,
    spec?: V1ThemeSpec | undefined,
  ) {
    if (spec) {
      this.colors = this.processTheme(spec);
      this.css.set(this.generateCSS());
    }
  }

  getColor(name: string, dark: boolean, fallback: string = "500") {
    const color = this.colors?.[dark ? "dark" : "light"]?.[name]?.css("hsl");
    if (!color) {
      return primary[fallback].css("hsl");
    }
    return color;
  }

  updateThemeSpec(spec: V1ThemeSpec) {
    this.colors = this.processTheme(spec);
    this.css.set(this.generateCSS());

    // Compatibility with current implementation, this can eventually be removed
    this.resolvedThemeObject.set({
      dark: resolveThemeObject(spec, true) ?? {},
      light: resolveThemeObject(spec, false) ?? {},
    });
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
 :where([data-theme-scope="${this.scope}"]) {
  ${this.stringifyVars(lightColors)}
}
:where(.dark) :where([data-theme-scope="${this.scope}"]) {
  ${this.stringifyVars(darkColors)}
}`.trim();
  }

  private processTheme(spec: V1ThemeSpec) {
    const darkColors = this.processColors(spec.dark ?? {});
    const lightColors = this.processColors(spec.light ?? {});

    return { dark: darkColors, light: lightColors };
  }

  private processColors(colors: V1ThemeColors) {
    const finalColors: Colors = {};
    const { primary, secondary, variables } = colors;

    if (primary) {
      const primaryPalette = generateColorPalette(chroma(primary));
      for (const [i, color] of primaryPalette.entries()) {
        finalColors[`color-theme-${TailwindColorSpacing[i]}`] = color;
      }
    }

    if (secondary) {
      const secondaryPalette = generateColorPalette(chroma(secondary));
      for (const [i, color] of secondaryPalette.entries()) {
        finalColors[`color-theme-secondary-${TailwindColorSpacing[i]}`] = color;
      }
    }

    for (const [k, v] of Object.entries(variables ?? {})) {
      if (!v) continue;
      finalColors[k] = chroma(v);
    }

    return finalColors;
  }
}
