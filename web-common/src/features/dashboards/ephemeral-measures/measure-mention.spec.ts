import { describe, expect, it } from "vitest";
import {
  applyMeasureMention,
  filterMeasureMentions,
  findMeasureMention,
} from "./measure-mention";

describe("findMeasureMention", () => {
  it("finds the mention the cursor is in", () => {
    expect(findMeasureMention("@tot", 4)).toEqual({ start: 0, query: "tot" });
    expect(findMeasureMention("cost - @net rev", 15)).toEqual({
      start: 7,
      query: "net rev",
    });
    expect(findMeasureMention("(@", 2)).toEqual({ start: 1, query: "" });
  });

  it("uses the closest @ before the cursor", () => {
    expect(findMeasureMention("@a + @b", 7)).toEqual({ start: 5, query: "b" });
    // The cursor is before the second "@", so it is still in the first mention.
    expect(findMeasureMention("@a + @b", 4)).toEqual({
      start: 0,
      query: "a +",
    });
  });

  it("returns null without an @ before the cursor", () => {
    expect(findMeasureMention("total - cost", 5)).toBeNull();
    expect(findMeasureMention("", 0)).toBeNull();
    expect(findMeasureMention("total @x", 3)).toBeNull();
  });

  it("ignores an @ that directly follows an identifier or quoted name", () => {
    expect(findMeasureMention("total@x", 7)).toBeNull();
    expect(findMeasureMention('"a b"@x', 7)).toBeNull();
  });
});

describe("filterMeasureMentions", () => {
  const measures = [
    { name: "total_revenue", displayName: "Total Revenue" },
    { name: "total_cost", displayName: "Total Cost" },
    { name: "net_rev", displayName: "Net Revenue" },
    { name: "margin" },
  ];

  it("returns everything for an empty query", () => {
    expect(filterMeasureMentions(measures, "")).toEqual(measures);
    expect(filterMeasureMentions(measures, "  ")).toEqual(measures);
  });

  it("matches display names and names case-insensitively", () => {
    expect(
      filterMeasureMentions(measures, "REV").map((mes) => mes.name),
    ).toEqual(["total_revenue", "net_rev"]);
    expect(
      filterMeasureMentions(measures, "net rev ").map((mes) => mes.name),
    ).toEqual(["net_rev"]);
    expect(
      filterMeasureMentions(measures, "marg").map((mes) => mes.name),
    ).toEqual(["margin"]);
    expect(filterMeasureMentions(measures, "profit")).toEqual([]);
  });
});

describe("applyMeasureMention", () => {
  it("replaces the mention with the measure reference and a space", () => {
    expect(
      applyMeasureMention("@tot", { start: 0, query: "tot" }, "total_cost"),
    ).toEqual({ value: "total_cost ", cursor: 11 });
    expect(
      applyMeasureMention(
        "cost - @net rev",
        { start: 7, query: "net rev" },
        "net_rev",
      ),
    ).toEqual({ value: "cost - net_rev ", cursor: 15 });
  });

  it("does not add a space when one already follows", () => {
    expect(
      applyMeasureMention("@t * 2", { start: 0, query: "t" }, "total"),
    ).toEqual({ value: "total * 2", cursor: 5 });
  });

  it("quotes names that need it", () => {
    expect(
      applyMeasureMention("@sel", { start: 0, query: "sel" }, "select"),
    ).toEqual({ value: '"select" ', cursor: 9 });
  });
});
