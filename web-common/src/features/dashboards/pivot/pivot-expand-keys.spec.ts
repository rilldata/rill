import { describe, expect, it } from "vitest";
import {
  buildExpandKey,
  childExpandKey,
  expandKeyDepth,
  expandKeySegments,
  parentExpandKey,
} from "./pivot-expand-keys";

const dims = ["app", "src", "month"];

describe("pivot-expand-keys", () => {
  it("has one segment per level", () => {
    const key = buildExpandKey(dims, [
      "Galatea-Stories",
      "facebook",
      "Jul 2026",
    ]);
    expect(expandKeySegments(key)).toHaveLength(3);
    expect(expandKeyDepth(key)).toBe(3);
  });

  it("keeps values that contain dots, spaces or the separator intact", () => {
    const key = buildExpandKey(dims, ["a.b.c", "x y z", "p\0q"]);
    expect(expandKeyDepth(key)).toBe(3);
    expect(parentExpandKey(key)).toBe(buildExpandKey(dims, ["a.b.c", "x y z"]));
  });

  it("encodes null distinctly from a genuinely absent level", () => {
    const withNull = buildExpandKey(dims, ["app", null]);
    const shallow = buildExpandKey(dims, ["app"]);
    expect(withNull).not.toBe(shallow);
    expect(expandKeyDepth(withNull)).toBe(2);
    expect(expandKeyDepth(shallow)).toBe(1);
  });

  it("encodes null distinctly from strings that look like null", () => {
    const keys = [null, "<NULL>", "null"].map((v) => buildExpandKey(dims, [v]));
    expect(new Set(keys).size).toBe(3);
  });

  it("keys a top-level empty-string value as depth 1, not as the root", () => {
    const key = buildExpandKey(dims, [""]);
    expect(key).not.toBe("");
    expect(expandKeyDepth(key)).toBe(1);
    expect(parentExpandKey(key)).toBe("");
    expect(expandKeyDepth(buildExpandKey(dims, ["", ""]))).toBe(2);
  });

  it("keys the same value under different dimensions differently", () => {
    expect(buildExpandKey(["player_size"], ["Not Available"])).not.toBe(
      buildExpandKey(["device_os"], ["Not Available"]),
    );
  });

  it("computes the parent by stripping the deepest segment", () => {
    const key = buildExpandKey(dims, ["a", "b", "c"]);
    expect(parentExpandKey(key)).toBe(buildExpandKey(dims, ["a", "b"]));
    expect(parentExpandKey(buildExpandKey(dims, ["a"]))).toBe("");
    expect(parentExpandKey("")).toBe("");
  });

  it("childExpandKey extends a parent and is inverse of parentExpandKey", () => {
    const parent = buildExpandKey(dims, ["a", "b"]);
    const child = childExpandKey(parent, "month", "c");
    expect(child).toBe(buildExpandKey(dims, ["a", "b", "c"]));
    expect(parentExpandKey(child)).toBe(parent);
    expect(childExpandKey("", "app", "a")).toBe(buildExpandKey(dims, ["a"]));
  });

  it("treats the empty root key as depth 0", () => {
    expect(expandKeyDepth("")).toBe(0);
    expect(expandKeySegments("")).toEqual([]);
  });
});
