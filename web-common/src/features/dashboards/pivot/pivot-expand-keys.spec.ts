import { describe, expect, it } from "vitest";
import {
  buildExpandKey,
  childExpandKey,
  encodeExpandKeyValue,
  expandKeyDepth,
  expandKeySegments,
  parentExpandKey,
} from "./pivot-expand-keys";

describe("pivot-expand-keys", () => {
  it("round-trips a value path through build/segments", () => {
    const values = ["Galatea-Stories", "facebook", "Jul 2026"];
    const key = buildExpandKey(values);
    expect(expandKeySegments(key)).toEqual(values.map(encodeExpandKeyValue));
    expect(expandKeyDepth(key)).toBe(3);
  });

  it("keeps values that contain dots or spaces intact", () => {
    const values = ["a.b.c", "x y z"];
    const key = buildExpandKey(values);
    expect(expandKeySegments(key)).toEqual(values.map(encodeExpandKeyValue));
    expect(expandKeyDepth(key)).toBe(2);
  });

  it("encodes null distinctly from a genuinely absent level", () => {
    const withNull = buildExpandKey(["app", null]);
    const shallow = buildExpandKey(["app"]);
    expect(withNull).not.toBe(shallow);
    expect(expandKeyDepth(withNull)).toBe(2);
    expect(expandKeyDepth(shallow)).toBe(1);
  });

  it("encodes null distinctly from strings that look like null", () => {
    const keys = [null, "<NULL>", "null"].map((v) => buildExpandKey([v]));
    expect(new Set(keys).size).toBe(3);
  });

  it("keys a top-level empty-string value as depth 1, not as the root", () => {
    const key = buildExpandKey([""]);
    expect(key).not.toBe("");
    expect(expandKeyDepth(key)).toBe(1);
    expect(parentExpandKey(key)).toBe("");
    expect(expandKeyDepth(buildExpandKey(["", ""]))).toBe(2);
  });

  it("computes the parent by stripping the deepest segment", () => {
    const key = buildExpandKey(["app", "src", "month"]);
    const parent = parentExpandKey(key);
    expect(parent).toBe(buildExpandKey(["app", "src"]));
    expect(parentExpandKey(buildExpandKey(["app"]))).toBe("");
    expect(parentExpandKey("")).toBe("");
  });

  it("childExpandKey extends a parent and is inverse of parentExpandKey", () => {
    const parent = buildExpandKey(["app", "src"]);
    const child = childExpandKey(parent, "month");
    expect(child).toBe(buildExpandKey(["app", "src", "month"]));
    expect(parentExpandKey(child)).toBe(parent);
    expect(childExpandKey("", "app")).toBe(buildExpandKey(["app"]));
  });

  it("treats the empty root key as depth 0", () => {
    expect(expandKeyDepth("")).toBe(0);
    expect(expandKeySegments("")).toEqual([]);
  });
});
