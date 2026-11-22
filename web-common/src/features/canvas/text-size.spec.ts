import { describe, expect, it } from "vitest";
import { resolveTextScale } from "./text-size";

describe("resolveTextScale", () => {
  it("is 1 when neither the component nor the canvas sets a size", () => {
    expect(resolveTextScale(undefined, undefined)).toBe(1);
  });

  it("takes the canvas's size for a component without one", () => {
    expect(resolveTextScale(undefined, "large")).toBe(1.25);
  });

  it("lets a component's size replace the canvas's, rather than compound it", () => {
    expect(resolveTextScale("xlarge", "large")).toBe(1.5);
    expect(resolveTextScale("medium", "large")).toBe(1);
    expect(resolveTextScale("small", undefined)).toBe(0.875);
  });

  // The runtime rejects other values, but an unvalidated draft or a templated
  // value can still reach the frontend.
  it("skips a value that is not a preset", () => {
    expect(resolveTextScale("huge", "large")).toBe(1.25);
    expect(resolveTextScale(1.5, undefined)).toBe(1);
    expect(resolveTextScale("", "")).toBe(1);
  });
});
