import { describe, it, expect } from "vitest";
import { stripUndefined } from "./strip-undefined";

describe("stripUndefined", () => {
  it("removes top-level undefined values", () => {
    expect(stripUndefined({ a: 1, b: undefined, c: "x" })).toEqual({
      a: 1,
      c: "x",
    });
  });

  it("preserves null (null !== undefined)", () => {
    expect(stripUndefined({ a: null })).toEqual({ a: null });
  });

  it("preserves falsy non-undefined values", () => {
    expect(stripUndefined({ a: 0, b: false, c: "" })).toEqual({
      a: 0,
      b: false,
      c: "",
    });
  });

  it("recursively strips nested objects", () => {
    expect(stripUndefined({ a: { b: 1, c: undefined } })).toEqual({
      a: { b: 1 },
    });
  });

  it("strips undefined inside arrays of objects", () => {
    expect(stripUndefined({ items: [{ a: 1, b: undefined }] })).toEqual({
      items: [{ a: 1 }],
    });
  });

  it("leaves primitive arrays untouched", () => {
    expect(stripUndefined({ tags: ["a", "b"] })).toEqual({
      tags: ["a", "b"],
    });
  });

  it("does not recurse into Date instances", () => {
    const d = new Date("2024-01-01");
    const result = stripUndefined({ created: d });
    expect(result.created).toBe(d);
  });

  it("handles deeply nested structures", () => {
    expect(stripUndefined({ a: { b: { c: { d: undefined, e: 1 } } } })).toEqual(
      { a: { b: { c: { e: 1 } } } },
    );
  });

  it("handles empty objects", () => {
    expect(stripUndefined({})).toEqual({});
  });

  it("handles arrays containing nested arrays", () => {
    expect(
      stripUndefined({
        matrix: [
          [1, 2],
          [3, 4],
        ],
      }),
    ).toEqual({
      matrix: [
        [1, 2],
        [3, 4],
      ],
    });
  });

  it("strips all keys when every value is undefined", () => {
    expect(stripUndefined({ a: undefined, b: undefined })).toEqual({});
  });
});
