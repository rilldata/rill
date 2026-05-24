import { describe, expect, it } from "vitest";
import {
  applyHideAllInTag,
  applyOnlyShowTag,
  applyShowAllInTag,
  computeTagVisibility,
  deriveTags,
  itemHasTag,
  namesInTag,
  type Taggable,
} from "./tag-utils";

const items: Taggable[] = [
  { name: "country", tags: ["Geography"] },
  { name: "city", tags: ["Geography", "Customer"] },
  { name: "state", tags: ["Geography"] },
  { name: "customer_id", tags: ["Customer"] },
  { name: "untagged" },
  { name: "spaced", tags: ["  Product  "] },
];

describe("tag-utils", () => {
  describe("deriveTags", () => {
    it("preserves first-appearance order and counts per tag", () => {
      expect(deriveTags(items)).toEqual([
        { name: "Geography", displayName: "Geography", totalCount: 3 },
        { name: "Customer", displayName: "Customer", totalCount: 2 },
        { name: "Product", displayName: "Product", totalCount: 1 },
      ]);
    });

    it("ignores empty and whitespace-only tag strings", () => {
      expect(
        deriveTags([
          { name: "a", tags: ["", "   ", "X"] },
          { name: "b", tags: ["X"] },
        ]),
      ).toEqual([{ name: "X", displayName: "X", totalCount: 2 }]);
    });

    it("treats casing as distinct", () => {
      expect(
        deriveTags([
          { name: "a", tags: ["X"] },
          { name: "b", tags: ["x"] },
        ]),
      ).toEqual([
        { name: "X", displayName: "X", totalCount: 1 },
        { name: "x", displayName: "x", totalCount: 1 },
      ]);
    });
  });

  describe("itemHasTag", () => {
    it("matches trimmed tag strings", () => {
      expect(itemHasTag({ tags: ["  Product  "] }, "Product")).toBe(true);
      expect(itemHasTag({ tags: ["Product"] }, "Other")).toBe(false);
      expect(itemHasTag({}, "Product")).toBe(false);
    });
  });

  describe("namesInTag", () => {
    it("returns items in spec order", () => {
      expect(namesInTag(items, "Geography")).toEqual([
        "country",
        "city",
        "state",
      ]);
      expect(namesInTag(items, "Product")).toEqual(["spaced"]);
      expect(namesInTag(items, "Missing")).toEqual([]);
    });
  });

  describe("computeTagVisibility", () => {
    it("returns 'none' when no items in tag are visible", () => {
      expect(computeTagVisibility(items, ["untagged"], "Geography")).toEqual({
        tagName: "Geography",
        visibleCount: 0,
        totalCount: 3,
        state: "none",
      });
    });

    it("returns 'partial' when some items in tag are visible", () => {
      expect(
        computeTagVisibility(items, ["country", "city"], "Geography"),
      ).toEqual({
        tagName: "Geography",
        visibleCount: 2,
        totalCount: 3,
        state: "partial",
      });
    });

    it("returns 'all' when every item in tag is visible", () => {
      expect(
        computeTagVisibility(items, ["country", "city", "state"], "Geography"),
      ).toEqual({
        tagName: "Geography",
        visibleCount: 3,
        totalCount: 3,
        state: "all",
      });
    });
  });

  describe("applyShowAllInTag", () => {
    it("unions current visible with all items in the tag, ordered by spec", () => {
      expect(applyShowAllInTag(["customer_id"], items, "Geography")).toEqual([
        "country",
        "city",
        "state",
        "customer_id",
      ]);
    });

    it("is idempotent when all tag items are already visible", () => {
      expect(
        applyShowAllInTag(["country", "city", "state"], items, "Geography"),
      ).toEqual(["country", "city", "state"]);
    });
  });

  describe("applyHideAllInTag", () => {
    it("removes items in the tag while preserving others", () => {
      expect(
        applyHideAllInTag(
          ["country", "city", "customer_id"],
          items,
          "Geography",
        ),
      ).toEqual(["customer_id"]);
    });

    it("does not touch dimensions outside the tag", () => {
      expect(applyHideAllInTag(["untagged"], items, "Geography")).toEqual([
        "untagged",
      ]);
    });

    it("falls back to the first visible item to keep at least one visible", () => {
      expect(
        applyHideAllInTag(["country", "city"], items, "Geography"),
      ).toEqual(["country"]);
    });
  });

  describe("applyOnlyShowTag", () => {
    it("limits visible names to items in the tag", () => {
      expect(
        applyOnlyShowTag(["country", "customer_id"], items, "Geography"),
      ).toEqual(["country", "city", "state"]);
    });

    it("falls back to keep one item visible if the tag is empty", () => {
      expect(applyOnlyShowTag(["country"], items, "Missing")).toEqual([
        "country",
      ]);
    });
  });
});
