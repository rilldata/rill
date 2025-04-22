import { describe, it, expect } from "vitest";
import { splitDimensionSearchText } from "./split-dimension-search-text";

describe("splitDimensionSearchText", () => {
  it("should split by comma and return trimmed parts", () => {
    const result = splitDimensionSearchText("facebook, google   ,   rill,");
    expect(result).toEqual(["facebook", "google", "rill"]);
  });

  it("should split by newline and return trimmed parts", () => {
    const result = splitDimensionSearchText(`facebook
    google
    rill
    `);
    expect(result).toEqual(["facebook", "google", "rill"]);
  });

  it("should split by newline when comma is present and return trimmed parts", () => {
    const result = splitDimensionSearchText(`facebook  ,  google
    rill
    `);
    expect(result).toEqual(["facebook  ,  google", "rill"]);
  });
});
