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

    it("leave fg-inverse to app.css when only fg-primary is set", () => {
      // fg-inverse is the text on tooltips and primary buttons, whose backgrounds stay dark
      // or brand-colored. Deriving it from a dark fg-primary made that text unreadable.
      const { light, dark } = parseCSS(
        new Theme({
          light: { primary: PRIMARY, variables: { "fg-primary": "#0f172a" } },
          dark: { primary: PRIMARY, variables: { "fg-primary": "#f1f5f9" } },
        }).css,
      );

      expect(light).not.toHaveProperty("fg-inverse");
      expect(dark).not.toHaveProperty("fg-inverse");
      // The rest of the hierarchy still follows fg-primary.
      expect(light["fg-muted"]).toContain(light["fg-primary"]);
    });

    it("keep an explicit fg-inverse, and reset a light-only one in dark mode", () => {
      const { light, dark } = parseCSS(
        new Theme({
          light: {
            variables: { "fg-primary": "#0f172a", "fg-inverse": "#ffffff" },
          },
          dark: { variables: { "fg-primary": "#f1f5f9" } },
        }).css,
      );

      expect(light["fg-inverse"]).toBe("hsl(0deg 0% 100%)");
      expect(dark["fg-inverse"]).toBe("unset");
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

  describe("length variables", () => {
    it("emits radius as a length instead of parsing it as a color", () => {
      const { light, dark } = parseCSS(
        new Theme({
          light: { variables: { radius: "1.5rem", border: "#ff0000" } },
          dark: { variables: { radius: "0px" } },
        }).css,
      );

      expect(light["radius"]).toBe("1.5rem");
      expect(light["border"]).toBe("hsl(0deg 100% 50%)");
      expect(dark["radius"]).toBe("0px");
    });

    it("resets a light-only radius in dark mode, like a light-only color", () => {
      const { light, dark } = parseCSS(
        new Theme({ light: { variables: { radius: "12px" } } }).css,
      );

      expect(light["radius"]).toBe("12px");
      expect(dark["radius"]).toBe("unset");
    });

    it("drops a value that is not a plain length", () => {
      // The value is written into a stylesheet as-is, so it must not be able to close the rule.
      const css = new Theme({
        light: { variables: { radius: "0; } body { display: none" } },
      }).css;

      expect(parseCSS(css).light).not.toHaveProperty("radius");
      expect(css).not.toContain("display");
    });
  });

  describe("font and chrome tokens", () => {
    it("applies the font to the boundary and declares the chrome tokens once, for both modes", () => {
      const theme = new Theme({
        font: { family: "Brand Sans" },
        chrome: { borderWidthPx: 0, fontSizePx: 13, fontWeight: 600 },
        dark: { primary: PRIMARY },
      });
      const { light, dark } = parseCSS(theme.css);

      expect(theme.font?.cssFamily).toBe(`"Brand Sans", "Inter", sans-serif`);
      expect(light["theme-font-family"]).toBe(theme.font?.cssFamily);
      expect(theme.css).toContain("font-family: var(--theme-font-family);");
      // A width of zero removes the borders, so it must not be mistaken for unset.
      expect(light["chrome-border-width"]).toBe("0px");
      expect(light["chrome-font-size"]).toBe("13px");
      expect(light["chrome-font-weight"]).toBe("600");
      expect(dark["theme-font-family"]).toBeUndefined();
      expect(dark["chrome-border-width"]).toBeUndefined();
    });

    it("leaves Rill's font and chrome alone when the theme sets neither", () => {
      const theme = new Theme({ light: { primary: PRIMARY } });

      expect(theme.font).toBeUndefined();
      expect(theme.css).not.toContain("font-family");
      expect(theme.css).not.toContain("--chrome-");
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
