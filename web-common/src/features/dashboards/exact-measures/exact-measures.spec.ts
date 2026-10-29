import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  createExactMeasureContext,
  hasExactVariant,
  precisionStateFor,
  requestExactNames,
  toggleSurfaceInSelection,
  withExactVariant,
} from "./exact-measures";

const MEASURES: MetricsViewSpecMeasure[] = [
  {
    name: "users",
    expression: "approx_count_distinct(user_id)",
    exactExpression: "count(DISTINCT user_id)",
  },
  {
    name: "sessions",
    expression: "approx_count_distinct(session_id)",
    exactExpression: "count(DISTINCT session_id)",
  },
  { name: "revenue", expression: "sum(revenue)" },
  // Derived measures cannot define an exact expression: they are calculated
  // from the other measures as the query computes them.
  {
    name: "revenue_per_user",
    expression: "revenue / users",
    referencedMeasures: ["revenue", "users"],
  },
];

function context(selected: string[]) {
  return createExactMeasureContext(new Set(selected), MEASURES);
}

describe("hasExactVariant", () => {
  it("is true for a measure with its own exact expression", () => {
    expect(hasExactVariant("users", context([]))).toBe(true);
  });

  it("is false for derived measures and measures without one", () => {
    expect(hasExactVariant("revenue", context([]))).toBe(false);
    expect(hasExactVariant("revenue_per_user", context([]))).toBe(false);
    expect(hasExactVariant("unknown", context([]))).toBe(false);
  });
});

describe("resolved exact measures", () => {
  it("is empty without a selection", () => {
    expect(context([]).exact.size).toBe(0);
  });

  it("is the selected measures that define an exact expression", () => {
    expect([...context(["users"]).exact]).toEqual(["users"]);
    expect([...context(["users", "revenue", "sessions"]).exact]).toEqual([
      "users",
      "sessions",
    ]);
  });

  it("ignores a selection without exact variants, including derived measures", () => {
    expect(context(["revenue"]).exact.size).toBe(0);
    expect(context(["revenue_per_user"]).exact.size).toBe(0);
  });
});

describe("precisionStateFor", () => {
  it("returns null for measures without an exact variant, so they render no chip", () => {
    expect(precisionStateFor("revenue", context(["users"]))).toBeNull();
    expect(
      precisionStateFor("revenue_per_user", context(["users"])),
    ).toBeNull();
  });

  it("is approximate by default", () => {
    expect(precisionStateFor("users", context([]))).toBe("approx");
  });

  it("is exact for selected measures", () => {
    expect(precisionStateFor("users", context(["users"]))).toBe("exact");
  });

  it("stays approximate for measures outside the selection", () => {
    expect(precisionStateFor("sessions", context(["users"]))).toBe("approx");
  });
});

describe("requestExactNames", () => {
  it("is empty when nothing is selected", () => {
    expect(
      requestExactNames(["users", "revenue_per_user"], context([])),
    ).toEqual(new Set());
  });

  it("is empty when the selection has no exact variant", () => {
    expect(requestExactNames(["revenue"], context(["revenue"]))).toEqual(
      new Set(),
    );
  });

  it("flags the requested measures that are selected", () => {
    expect(
      requestExactNames(
        ["users", "revenue", "revenue_per_user"],
        context(["users", "sessions"]),
      ),
    ).toEqual(new Set(["users"]));
  });
});

describe("withExactVariant", () => {
  it("keeps the measures with an exact expression of their own", () => {
    expect(
      withExactVariant(
        ["users", "revenue", "revenue_per_user", "sessions"],
        context([]),
      ),
    ).toEqual(["users", "sessions"]);
  });
});

describe("toggleSurfaceInSelection", () => {
  const key = "ctx";

  it("selects every measure of the surface at once", () => {
    expect(
      toggleSurfaceInSelection(undefined, key, "chart", ["users", "sessions"]),
    ).toEqual({ contextKey: key, surfaces: { chart: ["users", "sessions"] } });
  });

  it("clears the surface when all of its measures are selected, keeping other surfaces", () => {
    const selection = {
      contextKey: key,
      surfaces: { chart: ["users", "sessions"], kpi: ["users"] },
    };
    expect(
      toggleSurfaceInSelection(selection, key, "chart", ["sessions", "users"]),
    ).toEqual({ contextKey: key, surfaces: { kpi: ["users"] } });
  });

  it("completes a partial selection instead of clearing it", () => {
    const selection = { contextKey: key, surfaces: { chart: ["users"] } };
    expect(
      toggleSurfaceInSelection(selection, key, "chart", ["users", "sessions"]),
    ).toEqual({ contextKey: key, surfaces: { chart: ["users", "sessions"] } });
  });

  it("returns undefined when nothing is selected any more", () => {
    const selection = { contextKey: key, surfaces: { chart: ["users"] } };
    expect(
      toggleSurfaceInSelection(selection, key, "chart", ["users"]),
    ).toBeUndefined();
    expect(
      toggleSurfaceInSelection(undefined, key, "chart", []),
    ).toBeUndefined();
  });

  it("discards a selection made in another query context", () => {
    const selection = { contextKey: "old", surfaces: { kpi: ["users"] } };
    expect(
      toggleSurfaceInSelection(selection, key, "chart", ["users"]),
    ).toEqual({ contextKey: key, surfaces: { chart: ["users"] } });
  });
});
