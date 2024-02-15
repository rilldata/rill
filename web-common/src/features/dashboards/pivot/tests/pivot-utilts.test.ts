import { sortAcessors } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import { describe, expect, it } from "vitest";

describe("sortAcessors function", () => {
  it("should correctly sort accessors with basic sorting", () => {
    const input = ["c1v2m3", "c0v0m0", "c2v3m1"];
    const expected = ["c0v0m0", "c1v2m3", "c2v3m1"];
    expect(sortAcessors(input)).toEqual(expected);
  });

  it("should sort accessors with varying numbers of c<num>v<num> sequences", () => {
    const input = ["c0v1_c1v2m0", "c0v0_c1v1m2", "c0v0_c1v0m1"];
    const expected = ["c0v0_c1v0m1", "c0v0_c1v1m2", "c0v1_c1v2m0"];
    expect(sortAcessors(input)).toEqual(expected);
  });

  it("should sort accessors with the same c-v values but different m values", () => {
    const input = ["c0v1m3", "c0v1m1", "c0v1m2"];
    const expected = ["c0v1m1", "c0v1m2", "c0v1m3"];
    expect(sortAcessors(input)).toEqual(expected);
  });

  it("should sort accessors with multiple c-v-m sequences, including different-number lengths", () => {
    const input = ["c1v10_c2v20m30", "c1v2_c2v3m4", "c1v10_c2v3m4"];
    const expected = ["c1v2_c2v3m4", "c1v10_c2v3m4", "c1v10_c2v20m30"];
    expect(sortAcessors(input)).toEqual(expected);
  });
});
