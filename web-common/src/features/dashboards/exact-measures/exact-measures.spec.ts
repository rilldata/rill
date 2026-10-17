import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  createExactMeasureContext,
  hasExactVariant,
  precisionStateFor,
  requestExactNames,
  toggleSurfaceInSelection,
  transitiveRefs,
  withExactVariant,
} from "./exact-measures";

const MEASURES: MetricsViewSpecMeasure[] = [
  {
    name: "users",
    expression: "approx_count_distinct(user_id)",
    exactExpression: "count(DISTINCT user_id)",
  },
  { name: "revenue", expression: "sum(revenue)" },
  {
    name: "revenue_per_user",
    expression: "revenue / users",
    referencedMeasures: ["revenue", "users"],
  },
  {
    name: "revenue_per_user_pct",
    expression: "revenue_per_user * 100",
    referencedMeasures: ["revenue_per_user"],
  },
  // A reference cycle must not hang the closure.
  { name: "a", expression: "b + 1", referencedMeasures: ["b"] },
  { name: "b", expression: "a + 1", referencedMeasures: ["a"] },
];

const EPHEMERAL = [
  {
    name: "double_users",
    displayName: "Double users",
    expression: "users * 2",
  },
];

// Two exact-capable measures that share a derived measure.
const SIBLING_MEASURES: MetricsViewSpecMeasure[] = [
  ...MEASURES,
  {
    name: "sessions",
    expression: "approx_count_distinct(session_id)",
    exactExpression: "count(DISTINCT session_id)",
  },
  {
    name: "users_per_session",
    expression: "users / sessions",
    referencedMeasures: ["users", "sessions"],
  },
];

function context(selected: string[], measures = MEASURES) {
  return createExactMeasureContext(new Set(selected), measures, EPHEMERAL);
}

describe("transitiveRefs", () => {
  it("includes the measure and every measure it references, transitively", () => {
    expect([...transitiveRefs("revenue_per_user_pct", context([]))]).toEqual([
      "revenue_per_user_pct",
      "revenue_per_user",
      "users",
      "revenue",
    ]);
  });

  it("follows ephemeral expression refs", () => {
    expect([...transitiveRefs("double_users", context([]))]).toEqual([
      "double_users",
      "users",
    ]);
  });

  it("terminates on reference cycles", () => {
    expect([...transitiveRefs("a", context([]))].sort()).toEqual(["a", "b"]);
  });
});

describe("hasExactVariant", () => {
  it("is true for a measure with its own exact expression", () => {
    expect(hasExactVariant("users", context([]))).toBe(true);
  });

  it("is true for measures that reference one, at any depth", () => {
    expect(hasExactVariant("revenue_per_user", context([]))).toBe(true);
    expect(hasExactVariant("revenue_per_user_pct", context([]))).toBe(true);
    expect(hasExactVariant("double_users", context([]))).toBe(true);
  });

  it("is false otherwise", () => {
    expect(hasExactVariant("revenue", context([]))).toBe(false);
    expect(hasExactVariant("unknown", context([]))).toBe(false);
  });
});

describe("resolved exact measures", () => {
  it("is empty without a selection", () => {
    expect(context([]).exact.size).toBe(0);
  });

  it("is the selected measures that define an exact expression", () => {
    expect([...context(["users"]).exact]).toEqual(["users"]);
    expect([...context(["users", "revenue"]).exact]).toEqual(["users"]);
  });

  it("is empty for a selection without exact variants", () => {
    expect(context(["revenue"]).exact.size).toBe(0);
  });

  it("resolves a selected derived measure to the exact measures it is calculated from", () => {
    // The runtime computes `users` exactly in every query that requests
    // `revenue_per_user` exactly, so the dashboard treats `users` as exact.
    expect([...context(["revenue_per_user"]).exact]).toEqual(["users"]);
    expect([...context(["revenue_per_user_pct"]).exact]).toEqual(["users"]);
    expect([...context(["double_users"]).exact]).toEqual(["users"]);
  });

  it("is closed under the request rule", () => {
    // Requests flag `users_per_session` because it is calculated from `users`,
    // and the runtime then computes `sessions` exactly in those requests too.
    // Resolving `sessions` as exact makes it exact in every request instead of
    // only in the ones that happen to contain the ratio.
    expect([...context(["users"], SIBLING_MEASURES).exact].sort()).toEqual([
      "sessions",
      "users",
    ]);
  });
});

describe("precisionStateFor", () => {
  it("returns null for measures without an exact variant", () => {
    expect(precisionStateFor("revenue", context(["users"]))).toBeNull();
    expect(precisionStateFor("revenue", context(["revenue"]))).toBeNull();
  });

  it("is approximate by default", () => {
    expect(precisionStateFor("users", context([]))).toBe("approx");
    expect(precisionStateFor("revenue_per_user", context([]))).toBe("approx");
  });

  it("is exact for selected measures", () => {
    expect(precisionStateFor("users", context(["users"]))).toBe("exact");
    expect(
      precisionStateFor("revenue_per_user", context(["revenue_per_user"])),
    ).toBe("exact");
  });

  it("is inherited for measures calculated from a selected exact measure", () => {
    expect(precisionStateFor("revenue_per_user", context(["users"]))).toBe(
      "inherited",
    );
    expect(precisionStateFor("revenue_per_user_pct", context(["users"]))).toBe(
      "inherited",
    );
    expect(precisionStateFor("double_users", context(["users"]))).toBe(
      "inherited",
    );
  });

  it("is inherited for the exact measures a selected derived measure is calculated from", () => {
    const ctx = context(["revenue_per_user"]);
    expect(precisionStateFor("users", ctx)).toBe("inherited");
    // Other measures calculated from `users` are exact as well.
    expect(precisionStateFor("revenue_per_user_pct", ctx)).toBe("inherited");
    expect(precisionStateFor("double_users", ctx)).toBe("inherited");
  });

  it("is inherited for a sibling the runtime computes exactly alongside", () => {
    const ctx = context(["users"], SIBLING_MEASURES);
    expect(precisionStateFor("sessions", ctx)).toBe("inherited");
    expect(precisionStateFor("users_per_session", ctx)).toBe("inherited");
  });

  it("stays approximate for measures unrelated to the selection", () => {
    const ctx = context(
      ["sessions"],
      [
        ...MEASURES,
        {
          name: "sessions",
          expression: "approx_count_distinct(session_id)",
          exactExpression: "count(DISTINCT session_id)",
        },
      ],
    );
    expect(precisionStateFor("users", ctx)).toBe("approx");
    expect(precisionStateFor("revenue_per_user", ctx)).toBe("approx");
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

  it("flags the selected measures and every requested measure calculated from them", () => {
    expect(
      requestExactNames(
        ["users", "revenue", "revenue_per_user", "revenue_per_user_pct"],
        context(["users"]),
      ),
    ).toEqual(new Set(["users", "revenue_per_user", "revenue_per_user_pct"]));
  });

  it("flags only the requested measures", () => {
    expect(requestExactNames(["revenue_per_user"], context(["users"]))).toEqual(
      new Set(["revenue_per_user"]),
    );
    expect(requestExactNames(["revenue"], context(["users"]))).toEqual(
      new Set(),
    );
  });

  it("flags the exact measures a selected derived measure is calculated from, even when the derived measure is not requested", () => {
    const ctx = context(["revenue_per_user"]);
    expect(requestExactNames(["users", "revenue"], ctx)).toEqual(
      new Set(["users"]),
    );
    expect(requestExactNames(["revenue_per_user_pct"], ctx)).toEqual(
      new Set(["revenue_per_user_pct"]),
    );
  });

  it("flags a sibling in every request, not only next to the ratio", () => {
    const ctx = context(["users"], SIBLING_MEASURES);
    expect(requestExactNames(["sessions"], ctx)).toEqual(new Set(["sessions"]));
    expect(requestExactNames(["users_per_session", "revenue"], ctx)).toEqual(
      new Set(["users_per_session"]),
    );
  });
});

describe("withExactVariant", () => {
  it("keeps the measures that are, or are calculated from, an exact-capable measure", () => {
    expect(
      withExactVariant(
        ["users", "revenue", "revenue_per_user", "double_users"],
        context([]),
      ),
    ).toEqual(["users", "revenue_per_user", "double_users"]);
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
