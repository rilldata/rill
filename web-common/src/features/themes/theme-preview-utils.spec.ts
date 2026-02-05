import { describe, it, expect } from "vitest";
import { parseDocument } from "yaml";
import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
import {
  parseThemeFromYaml,
  extractThemeColors,
  updateThemeColor,
  SEQUENTIAL_COLOR_COUNT,
  DIVERGING_COLOR_COUNT,
  QUALITATIVE_COLOR_COUNT,
  BACKGROUND_FALLBACK_LIGHT,
  BACKGROUND_FALLBACK_DARK,
  CARD_FALLBACK_LIGHT,
  CARD_FALLBACK_DARK,
  FG_PRIMARY_FALLBACK_LIGHT,
  FG_PRIMARY_FALLBACK_DARK,
} from "./theme-preview-utils";

describe("parseThemeFromYaml", () => {
  it("parses valid YAML content into theme and themeData", () => {
    const yaml = `
light:
  primary: "#ff0000"
  surface-background: "#ffffff"
dark:
  primary: "#00ff00"
  surface-background: "#000000"
`;
    const { theme, themeData } = parseThemeFromYaml(yaml);

    expect(themeData).toEqual({
      light: {
        primary: "#ff0000",
        "surface-background": "#ffffff",
      },
      dark: {
        primary: "#00ff00",
        "surface-background": "#000000",
      },
    });
    expect(theme).toBeDefined();
  });

  it("handles null content", () => {
    const { theme, themeData } = parseThemeFromYaml(null);

    expect(themeData).toEqual({});
    expect(theme).toBeDefined();
  });

  it("handles empty string content", () => {
    const { theme, themeData } = parseThemeFromYaml("");

    expect(themeData).toEqual({});
    expect(theme).toBeDefined();
  });

  it("handles YAML with only light mode", () => {
    const yaml = `
light:
  primary: "#ff0000"
`;
    const { themeData } = parseThemeFromYaml(yaml);

    expect(themeData.light).toEqual({ primary: "#ff0000" });
    expect(themeData.dark).toBeUndefined();
  });
});

describe("extractThemeColors", () => {
  it("extracts colors for light mode", () => {
    const themeData = {
      light: {
        primary: "#ff0000",
        "surface-background": "#ffffff",
        "surface-card": "#f0f0f0",
        "fg-primary": "#111111",
        "surface-subtle": "#e0e0e0",
        "color-sequential-1": "#aaa",
        "color-sequential-2": "#bbb",
      },
      dark: {
        primary: "#00ff00",
      },
    };

    const colors = extractThemeColors(themeData, "light");

    expect(colors.primaryColor).toBe("#ff0000");
    expect(colors.backgroundColor).toBe("#ffffff");
    expect(colors.cardColor).toBe("#f0f0f0");
    expect(colors.fgPrimary).toBe("#111111");
    expect(colors.surfaceHeader).toBe("#e0e0e0");
    expect(colors.sequentialColors[0]).toBe("#aaa");
    expect(colors.sequentialColors[1]).toBe("#bbb");
  });

  it("extracts colors for dark mode", () => {
    const themeData = {
      light: {
        primary: "#ff0000",
      },
      dark: {
        primary: "#00ff00",
        "surface-background": "#000000",
        "surface-card": "#1a1a1a",
        "fg-primary": "#eeeeee",
        "surface-subtle": "#2a2a2a",
      },
    };

    const colors = extractThemeColors(themeData, "dark");

    expect(colors.primaryColor).toBe("#00ff00");
    expect(colors.backgroundColor).toBe("#000000");
    expect(colors.cardColor).toBe("#1a1a1a");
    expect(colors.fgPrimary).toBe("#eeeeee");
    expect(colors.surfaceHeader).toBe("#2a2a2a");
  });

  it("uses fallback values when colors are not defined", () => {
    const themeData = {};

    const lightColors = extractThemeColors(themeData, "light");
    expect(lightColors.primaryColor).toBe("var(--color-theme-500)");
    expect(lightColors.backgroundColor).toBe(BACKGROUND_FALLBACK_LIGHT);
    expect(lightColors.cardColor).toBe(CARD_FALLBACK_LIGHT);
    expect(lightColors.fgPrimary).toBe(FG_PRIMARY_FALLBACK_LIGHT);
    expect(lightColors.surfaceHeader).toBe("var(--surface-subtle)");

    const darkColors = extractThemeColors(themeData, "dark");
    expect(darkColors.backgroundColor).toBe(BACKGROUND_FALLBACK_DARK);
    expect(darkColors.cardColor).toBe(CARD_FALLBACK_DARK);
    expect(darkColors.fgPrimary).toBe(FG_PRIMARY_FALLBACK_DARK);
  });

  it("supports legacy color names (background, card)", () => {
    const themeData = {
      light: {
        background: "#fff000",
        card: "#000fff",
      },
    } as V1ThemeSpec;

    const colors = extractThemeColors(themeData, "light");
    expect(colors.backgroundColor).toBe("#fff000");
    expect(colors.cardColor).toBe("#000fff");
  });

  it("prefers new semantic names over legacy names", () => {
    const themeData = {
      light: {
        "surface-background": "#new-bg",
        background: "#old-bg",
        "surface-card": "#new-card",
        card: "#old-card",
      },
    } as V1ThemeSpec;

    const colors = extractThemeColors(themeData, "light");
    expect(colors.backgroundColor).toBe("#new-bg");
    expect(colors.cardColor).toBe("#new-card");
  });

  it("generates correct number of palette colors with fallbacks", () => {
    const themeData = {};

    const colors = extractThemeColors(themeData, "light");

    expect(colors.sequentialColors).toHaveLength(SEQUENTIAL_COLOR_COUNT);
    expect(colors.divergingColors).toHaveLength(DIVERGING_COLOR_COUNT);
    expect(colors.qualitativeColors).toHaveLength(QUALITATIVE_COLOR_COUNT);

    // Check fallback format
    expect(colors.sequentialColors[0]).toBe("var(--color-sequential-1)");
    expect(colors.divergingColors[5]).toBe("var(--color-diverging-6)");
    expect(colors.qualitativeColors[10]).toBe("var(--color-qualitative-11)");
  });

  it("extracts palette colors from theme data", () => {
    const themeData = {
      light: {
        "color-sequential-1": "#seq1",
        "color-sequential-5": "#seq5",
        "color-diverging-1": "#div1",
        "color-qualitative-1": "#qual1",
      },
    } as V1ThemeSpec;

    const colors = extractThemeColors(themeData, "light");

    expect(colors.sequentialColors[0]).toBe("#seq1");
    expect(colors.sequentialColors[4]).toBe("#seq5");
    expect(colors.divergingColors[0]).toBe("#div1");
    expect(colors.qualitativeColors[0]).toBe("#qual1");
  });
});

describe("updateThemeColor", () => {
  it("updates an existing color value", () => {
    const yaml = `light:
  primary: "#ff0000"
`;
    const doc = parseDocument(yaml);

    const result = updateThemeColor(doc, "light", "primary", "#00ff00");

    expect(result).toContain('primary: "#00ff00"');
  });

  it("adds a new color to existing mode section", () => {
    const yaml = `light:
  primary: "#ff0000"
`;
    const doc = parseDocument(yaml);

    const result = updateThemeColor(doc, "light", "secondary", "#0000ff");

    expect(result).toContain('primary: "#ff0000"');
    expect(result).toContain('secondary: "#0000ff"');
  });

  it("creates mode section if it does not exist", () => {
    const yaml = `light:
  primary: "#ff0000"
`;
    const doc = parseDocument(yaml);

    const result = updateThemeColor(doc, "dark", "primary", "#00ff00");

    expect(result).toContain("light:");
    expect(result).toContain("dark:");
    expect(result).toContain('primary: "#00ff00"');
  });

  it("works with empty document", () => {
    const yaml = "";
    const doc = parseDocument(yaml);

    const result = updateThemeColor(doc, "light", "primary", "#ff0000");

    expect(result).toContain("light:");
    expect(result).toContain('primary: "#ff0000"');
  });

  it("removes inline comments by default", () => {
    const yaml = `light:
  primary: "#ff0000" # this is a comment
`;
    const doc = parseDocument(yaml);

    const result = updateThemeColor(doc, "light", "primary", "#00ff00");

    expect(result).not.toContain("# this is a comment");
    expect(result).toContain('primary: "#00ff00"');
  });

  it("preserves inline comments when removeComments is false", () => {
    const yaml = `light:
  primary: "#ff0000" # this is a comment
`;
    const doc = parseDocument(yaml);

    const result = updateThemeColor(doc, "light", "primary", "#00ff00", false);

    expect(result).toContain("# this is a comment");
    expect(result).toContain('primary: "#00ff00"');
  });

  it("handles hyphenated color keys", () => {
    const yaml = `light:
  surface-background: "#ffffff"
`;
    const doc = parseDocument(yaml);

    const result = updateThemeColor(
      doc,
      "light",
      "surface-background",
      "#000000",
    );

    expect(result).toContain('surface-background: "#000000"');
  });

  it("handles palette color keys", () => {
    const yaml = `light:
  color-sequential-1: "#aaaaaa"
`;
    const doc = parseDocument(yaml);

    const result = updateThemeColor(
      doc,
      "light",
      "color-sequential-1",
      "#bbbbbb",
    );

    expect(result).toContain('color-sequential-1: "#bbbbbb"');
  });
});
