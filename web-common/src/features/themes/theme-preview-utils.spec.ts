import { describe, it, expect } from "vitest";
import { parseDocument } from "yaml";
import { updateThemeColor } from "./theme-preview-utils";

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
