import { describe, it, expect } from "vitest";
import {
  getFiltersFromText,
  splitDimensionSearchText,
} from "web-common/src/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
import { V1Operation } from "web-common/src/runtime-client";

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

describe("getFiltersFromText", () => {
  it("should flatten array-wrapped values in IN expressions", () => {
    const { expr } = getFiltersFromText("username IN (['Ashwani Yadav'])");
    const inExpr = expr.cond?.exprs?.[0];
    expect(inExpr?.cond?.op).toBe(V1Operation.OPERATION_IN);
    expect(inExpr?.cond?.exprs).toEqual([
      { ident: "username" },
      { val: "Ashwani Yadav" },
    ]);
  });

  it("should flatten array-wrapped values in NIN expressions", () => {
    const { expr } = getFiltersFromText("username NIN (['Ashwani Yadav'])");
    const ninExpr = expr.cond?.exprs?.[0];
    expect(ninExpr?.cond?.op).toBe(V1Operation.OPERATION_NIN);
    expect(ninExpr?.cond?.exprs).toEqual([
      { ident: "username" },
      { val: "Ashwani Yadav" },
    ]);
  });

  it("should flatten multi-value array-wrapped IN expressions", () => {
    const { expr } = getFiltersFromText(
      "username IN (['Ashwani Yadav','John Doe'])",
    );
    const inExpr = expr.cond?.exprs?.[0];
    expect(inExpr?.cond?.op).toBe(V1Operation.OPERATION_IN);
    expect(inExpr?.cond?.exprs).toEqual([
      { ident: "username" },
      { val: "Ashwani Yadav" },
      { val: "John Doe" },
    ]);
  });

  it("should leave non-array IN expressions unchanged", () => {
    const { expr } = getFiltersFromText(
      "username IN ('Ashwani Yadav','John Doe')",
    );
    const inExpr = expr.cond?.exprs?.[0];
    expect(inExpr?.cond?.op).toBe(V1Operation.OPERATION_IN);
    expect(inExpr?.cond?.exprs).toEqual([
      { ident: "username" },
      { val: "Ashwani Yadav" },
      { val: "John Doe" },
    ]);
  });
});
