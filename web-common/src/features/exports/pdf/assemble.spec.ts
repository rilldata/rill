import { describe, expect, it } from "vitest";
import { parseColor } from "./assemble";

describe("parseColor", () => {
  it("parses rgb and rgba colors", () => {
    expect(parseColor("rgb(31, 35, 41)")).toEqual({ r: 31, g: 35, b: 41 });
    expect(parseColor("rgba(31, 35, 41, 1)")).toEqual({
      r: 31,
      g: 35,
      b: 41,
    });
  });

  it("treats transparent backgrounds as white", () => {
    expect(parseColor("rgba(0, 0, 0, 0)")).toEqual({
      r: 255,
      g: 255,
      b: 255,
    });
  });

  it("parses modern CSS colors used by theme tokens", () => {
    expect(parseColor("oklch(24.3% 0 0)")).toEqual({
      r: 32,
      g: 32,
      b: 32,
    });
    expect(parseColor("color(srgb 0.1 0.2 0.3)")).toEqual({
      r: 26,
      g: 51,
      b: 77,
    });
  });
});
