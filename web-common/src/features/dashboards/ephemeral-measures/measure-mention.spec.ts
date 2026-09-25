import { describe, expect, it } from "vitest";
import {
  filterMeasureMentions,
  serializeExpressionTokens,
  tokenizeMeasureExpression,
} from "./measure-mention";

const measures = [
  { name: "total_revenue", displayName: "Total Revenue" },
  { name: "total_cost", displayName: "Total Cost" },
  { name: "net_rev", displayName: "Net Revenue" },
  { name: "margin" },
  { name: "select", displayName: "Select" },
];

describe("filterMeasureMentions", () => {
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

describe("tokenizeMeasureExpression", () => {
  it("turns known measures into chips and keeps the rest as text", () => {
    expect(
      tokenizeMeasureExpression("total_revenue - total_cost * 2", measures),
    ).toEqual([
      { type: "measure", name: "total_revenue", displayName: "Total Revenue" },
      { type: "text", text: " - " },
      { type: "measure", name: "total_cost", displayName: "Total Cost" },
      { type: "text", text: " * 2" },
    ]);
  });

  it("keeps functions, literals and unknown names as text", () => {
    expect(
      tokenizeMeasureExpression(
        "round(margin / unknown, 2) + null + abs(-1e3)",
        measures,
      ),
    ).toEqual([
      { type: "text", text: "round(" },
      { type: "measure", name: "margin", displayName: "margin" },
      { type: "text", text: " / unknown, 2) + null + abs(-1e3)" },
    ]);
    // A function named like a measure is still a function call.
    expect(
      tokenizeMeasureExpression("round(1) + margin_2", [
        ...measures,
        { name: "round" },
        { name: "margin_2" },
      ]),
    ).toEqual([
      { type: "text", text: "round(1) + " },
      { type: "measure", name: "margin_2", displayName: "margin_2" },
    ]);
  });

  it("keeps an expression that does not parse as text", () => {
    expect(tokenizeMeasureExpression("margin +", measures)).toEqual([
      { type: "text", text: "margin +" },
    ]);
    expect(tokenizeMeasureExpression("margin @", measures)).toEqual([
      { type: "text", text: "margin @" },
    ]);
  });

  it("recognizes quoted references", () => {
    expect(tokenizeMeasureExpression('"select" + "net_rev"', measures)).toEqual(
      [
        { type: "measure", name: "select", displayName: "Select" },
        { type: "text", text: " + " },
        { type: "measure", name: "net_rev", displayName: "Net Revenue" },
      ],
    );
  });

  it("returns a single text token when nothing matches", () => {
    expect(tokenizeMeasureExpression("a + b", measures)).toEqual([
      { type: "text", text: "a + b" },
    ]);
    expect(tokenizeMeasureExpression("", measures)).toEqual([]);
  });
});

describe("serializeExpressionTokens", () => {
  it("round trips an expression, quoting names that need it", () => {
    const expression = 'total_revenue - "select" * (2 + net_rev)';
    expect(
      serializeExpressionTokens(
        tokenizeMeasureExpression(expression, measures),
      ),
    ).toBe(expression);
    expect(
      serializeExpressionTokens([
        { type: "measure", name: "select", displayName: "Select" },
        { type: "text", text: " " },
      ]),
    ).toBe('"select" ');
  });
});
