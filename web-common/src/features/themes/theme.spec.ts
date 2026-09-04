import { describe, expect, it } from "vitest";
import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
import { Theme } from "./theme";

const PRIMARY = "#49A9DE";
const SECONDARY = "#17407D";

/**
 * Splits the generated CSS into its light and dark rules, each as a
 * variable name (without the leading `--`) to value map.
 */
function parseCSS(css: string): {
  light: Record<string, string>;
  dark: Record<string, string>;
} {
  const blocks = [...css.matchAll(/([^{}]+)\{([^{}]*)\}/g)];

  const parseBlock = (body: string) =>
    Object.fromEntries(
      [...body.matchAll(/--([\w-]+):\s*([^;]+);/g)].map(([, name, value]) => [
        name,
        value.trim(),
      ]),
    );

  const find = (dark: boolean) =>
    parseBlock(
      blocks.find(
        ([, selector]) => selector.includes(".dark ") === dark,
      )?.[2] ?? "",
    );

  return { light: find(false), dark: find(true) };
}

describe("Theme", () => {
  describe("palette-derived semantic tokens", () => {
    // app.css declares these on :root against Rill's own palette. Custom properties are
    // substituted where they are declared, so a theme scoped to .dashboard-theme-boundary
    // cannot reach them; the theme has to re-declare them on the boundary itself.
    const colors = { primary: PRIMARY, secondary: SECONDARY };
    const { light, dark } = parseCSS(
      new Theme({ light: colors, dark: colors }).css,
    );

    it("derives the hover and active surfaces from the theme palette", () => {
      expect(light["surface-hover"]).toBe(light["color-theme-50"]);
      expect(light["surface-active"]).toBe(light["color-theme-100"]);
    });

    it("points popover-accent at surface-hover, as app.css does", () => {
      expect(light["popover-accent"]).toBe(light["color-theme-50"]);
      expect(light["popover-accent"]).toBe(light["surface-hover"]);
    });

    it("derives accents, focus rings, and dimension and measure chips", () => {
      expect(light["accent-primary"]).toBe(light["color-theme-500"]);
      expect(light["icon-accent"]).toBe(light["color-theme-500"]);
      expect(light["ring-focus"]).toBe(light["color-theme-300"]);
      expect(light["fg-accent"]).toBe(light["color-theme-900"]);
      expect(light["dimension"]).toBe(light["color-theme-100"]);
      expect(light["dimension-foreground"]).toBe(light["color-theme-800"]);
      expect(light["measure"]).toBe(light["color-theme-secondary-100"]);
      expect(light["measure-foreground"]).toBe(
        light["color-theme-secondary-800"],
      );
    });

    it("keeps the neutral gray surfaces in dark mode", () => {
      // Emitted as `unset` so the boundary inherits :root.dark rather than tinting.
      expect(dark["surface-hover"]).toBe("unset");
      expect(dark["surface-active"]).toBe("unset");
      expect(dark["popover-accent"]).toBe("unset");
    });

    it("uses light-palette shades for dark-mode accents, matching :root.dark", () => {
      // :root.dark reaches for --color-primary-light-* rather than the dark palette,
      // because those shades stay legible against dark surfaces. The light block's
      // color-theme-* are those same shades.
      expect(dark["accent-primary"]).toBe(light["color-theme-600"]);
      expect(dark["icon-accent"]).toBe(light["color-theme-400"]);
      expect(dark["dimension"]).toBe(light["color-theme-950"]);
    });

    it("falls back to Rill's defaults in dark mode when the theme is light-only", () => {
      const { dark } = parseCSS(new Theme({ light: colors }).css);

      expect(dark["accent-primary"]).toBe("unset");
      expect(dark["dimension"]).toBe("unset");
    });

    it("leaves secondary-derived tokens alone when the theme has no secondary", () => {
      const { light } = parseCSS(
        new Theme({ light: { primary: PRIMARY } }).css,
      );

      expect(light["dimension"]).toBe(light["color-theme-100"]);
      expect(light).not.toHaveProperty("measure");
      expect(light).not.toHaveProperty("accent-secondary-action");
    });
  });

  describe("explicit theme variables", () => {
    it("win over the derived values", () => {
      const { light } = parseCSS(
        new Theme({
          light: {
            primary: PRIMARY,
            variables: { "surface-hover": "#ff0000" },
          },
        }).css,
      );

      expect(light["surface-hover"]).toBe("hsl(0deg 100% 50%)");
      expect(light["popover-accent"]).toBe(light["surface-hover"]);
    });

    it("let fg-primary keep ownership of fg-accent", () => {
      const { light } = parseCSS(
        new Theme({
          light: {
            primary: PRIMARY,
            variables: { "fg-primary": "#111111" },
          },
        }).css,
      );

      expect(light["fg-accent"]).toBe(light["fg-primary"]);
    });

    it("let fg-primary keep ownership of fg-accent in dark mode too", () => {
      // The light block derives fg-accent from the palette. Resetting it to `unset` in the
      // dark block would both discard dark's own fg-primary and fall back to Rill's indigo.
      const { dark } = parseCSS(
        new Theme({
          light: { primary: PRIMARY },
          dark: {
            primary: PRIMARY,
            variables: { "fg-primary": "#eeeeee" },
          },
        }).css,
      );

      expect(dark["fg-accent"]).toBe(dark["fg-primary"]);
    });
  });

  describe("palette aliasing", () => {
    it("shadows Rill's primary and secondary palettes inside the boundary", () => {
      // So that the `bg-primary-500`-style utilities used throughout dashboard
      // components follow the theme without each one having to be migrated.
      const { light } = parseCSS(
        new Theme({ light: { primary: PRIMARY, secondary: SECONDARY } }).css,
      );

      expect(light["color-primary-500"]).toBe(light["color-theme-500"]);
      expect(light["color-secondary-500"]).toBe(
        light["color-theme-secondary-500"],
      );
    });
  });

  describe("legacy colors format", () => {
    it("derives the same tokens from primaryColorRaw and secondaryColorRaw", () => {
      const legacy: V1ThemeSpec = {
        primaryColorRaw: PRIMARY,
        secondaryColorRaw: SECONDARY,
      };
      const { light } = parseCSS(new Theme(legacy).css);

      expect(light["surface-hover"]).toBe(light["color-theme-50"]);
      expect(light["popover-accent"]).toBe(light["color-theme-50"]);
      expect(light["measure"]).toBe(light["color-theme-secondary-100"]);
    });
  });
});
