import { describe, expect, it } from "vitest";
import {
  applyHideAllInTag,
  applyOnlyShowTag,
  applyShowAllInTag,
  buildTagIndex,
  computeTagVisibility,
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

const index = buildTagIndex(items);

describe("tag-utils", () => {
  describe("buildTagIndex", () => {
    it("preserves first-appearance order and counts per tag", () => {
      expect(index.tags).toEqual([
        { name: "Geography", totalCount: 3 },
        { name: "Customer", totalCount: 2 },
        { name: "Product", totalCount: 1 },
      ]);
    });

    it("ignores empty and whitespace-only tag strings", () => {
      const idx = buildTagIndex([
        { name: "a", tags: ["", "   ", "X"] },
        { name: "b", tags: ["X"] },
      ]);
      expect(idx.tags).toEqual([{ name: "X", totalCount: 2 }]);
    });

    it("treats casing as distinct", () => {
      const idx = buildTagIndex([
        { name: "a", tags: ["X"] },
        { name: "b", tags: ["x"] },
      ]);
      expect(idx.tags).toEqual([
        { name: "X", totalCount: 1 },
        { name: "x", totalCount: 1 },
      ]);
    });

    it("buckets items per tag in spec order", () => {
      expect(index.itemsByTag.get("Geography")?.map((i) => i.name)).toEqual([
        "country",
        "city",
        "state",
      ]);
      expect(index.itemsByTag.get("Product")?.map((i) => i.name)).toEqual([
        "spaced",
      ]);
    });
  });

  describe("namesInTag", () => {
    it("returns items in spec order", () => {
      expect(namesInTag(index, "Geography")).toEqual([
        "country",
        "city",
        "state",
      ]);
      expect(namesInTag(index, "Product")).toEqual(["spaced"]);
      expect(namesInTag(index, "Missing")).toEqual([]);
    });
  });

  describe("computeTagVisibility", () => {
    it("returns 'none' when no items in tag are visible", () => {
      expect(computeTagVisibility(index, ["untagged"], "Geography")).toEqual({
        tagName: "Geography",
        visibleCount: 0,
        totalCount: 3,
        state: "none",
      });
    });

    it("returns 'partial' when some items in tag are visible", () => {
      expect(
        computeTagVisibility(index, ["country", "city"], "Geography"),
      ).toEqual({
        tagName: "Geography",
        visibleCount: 2,
        totalCount: 3,
        state: "partial",
      });
    });

    it("returns 'all' when every item in tag is visible", () => {
      expect(
        computeTagVisibility(index, ["country", "city", "state"], "Geography"),
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
      expect(applyShowAllInTag(["customer_id"], index, "Geography")).toEqual([
        "country",
        "city",
        "state",
        "customer_id",
      ]);
    });

    it("is idempotent when all tag items are already visible", () => {
      expect(
        applyShowAllInTag(["country", "city", "state"], index, "Geography"),
      ).toEqual(["country", "city", "state"]);
    });
  });

  describe("applyHideAllInTag", () => {
    it("removes items in the tag while preserving others", () => {
      expect(
        applyHideAllInTag(
          ["country", "city", "customer_id"],
          index,
          "Geography",
        ),
      ).toEqual(["customer_id"]);
    });

    it("does not touch dimensions outside the tag", () => {
      expect(applyHideAllInTag(["untagged"], index, "Geography")).toEqual([
        "untagged",
      ]);
    });

    it("falls back to the first visible item to keep at least one visible", () => {
      expect(
        applyHideAllInTag(["country", "city"], index, "Geography"),
      ).toEqual(["country"]);
    });
  });

  describe("applyOnlyShowTag", () => {
    it("limits visible names to items in the tag", () => {
      expect(
        applyOnlyShowTag(["country", "customer_id"], index, "Geography"),
      ).toEqual(["country", "city", "state"]);
    });

    it("falls back to keep one item visible if the tag is empty", () => {
      expect(applyOnlyShowTag(["country"], index, "Missing")).toEqual([
        "country",
      ]);
    });
  });
});
