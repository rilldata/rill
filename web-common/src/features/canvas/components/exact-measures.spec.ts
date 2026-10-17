import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import { createExactMeasureContext } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  canvasChartPrecisionChip,
  canvasExactRequestNames,
  canvasPrecisionChip,
  chartMeasureNames,
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
];

const NONE: ReadonlySet<string> = new Set();

describe("canvasExactRequestNames", () => {
  it("is empty without a selection", () => {
    expect(
      canvasExactRequestNames(["users"], NONE, MEASURES, undefined),
    ).toEqual([]);
  });

  it("flags the surface's selection and the measures calculated from it", () => {
    expect(
      canvasExactRequestNames(
        ["users", "revenue", "revenue_per_user"],
        new Set(["users"]),
        MEASURES,
        undefined,
      ),
    ).toEqual(["users", "revenue_per_user"]);
  });

  it("flags the measures a selected ratio is calculated from, in a component that shows only them", () => {
    expect(
      canvasExactRequestNames(
        ["users", "revenue"],
        new Set(["revenue_per_user"]),
        MEASURES,
        undefined,
      ),
    ).toEqual(["users"]);
  });

  it("is a no-op for a selection without exact variants", () => {
    expect(
      canvasExactRequestNames(
        ["revenue"],
        new Set(["revenue"]),
        MEASURES,
        undefined,
      ),
    ).toEqual([]);
  });

  it("follows ephemeral measures calculated from an exact measure", () => {
    expect(
      canvasExactRequestNames(["double_users"], new Set(["users"]), MEASURES, [
        {
          name: "double_users",
          displayName: "Double users",
          expression: "users * 2",
        },
      ]),
    ).toEqual(["double_users"]);
  });
});

describe("chartMeasureNames", () => {
  const params = {
    x: { type: "positional", meta: { chartFieldInput: { type: "dimension" } } },
    y: { type: "positional", meta: { chartFieldInput: { type: "measure" } } },
    measures: {
      type: "multi_fields",
      meta: { chartFieldInput: { type: "measure" } },
    },
    color: { type: "mark" },
  } as unknown as Record<string, ComponentInputParam>;

  it("collects the field and fields of measure inputs only", () => {
    expect(
      chartMeasureNames(params, {
        x: { field: "publisher" },
        y: { field: "users" },
        measures: { fields: ["users", "revenue"] },
        color: { field: "domain" },
      }),
    ).toEqual(["users", "revenue"]);
  });

  it("tolerates missing inputs", () => {
    expect(chartMeasureNames(params, {})).toEqual([]);
    expect(chartMeasureNames(undefined, { y: { field: "users" } })).toEqual([]);
  });
});

describe("canvasPrecisionChip", () => {
  const noop = () => {};

  it("is undefined for a measure without an exact variant", () => {
    const ctx = createExactMeasureContext(new Set(["users"]), MEASURES);
    expect(
      canvasPrecisionChip("revenue", "Revenue", ctx, false, undefined, noop),
    ).toBeUndefined();
  });

  it("renders the surface's selection as exact, with the query status", () => {
    const ctx = createExactMeasureContext(new Set(["users"]), MEASURES);
    expect(
      canvasPrecisionChip("users", "Users", ctx, true, "boom", noop),
    ).toMatchObject({
      state: "exact",
      loading: true,
      error: "boom",
      onToggle: noop,
    });
  });

  it("renders the measures a selected ratio is calculated from as inherited", () => {
    const ctx = createExactMeasureContext(
      new Set(["revenue_per_user"]),
      MEASURES,
    );
    expect(
      canvasPrecisionChip("users", "Users", ctx, false, undefined, noop),
    ).toMatchObject({ state: "inherited" });
  });

  it("does not report loading or errors for an approximate measure", () => {
    const ctx = createExactMeasureContext(NONE, MEASURES);
    expect(
      canvasPrecisionChip("users", "Users", ctx, true, "boom", noop),
    ).toMatchObject({ state: "approx", loading: false, error: undefined });
  });
});

describe("canvasChartPrecisionChip", () => {
  const noop = () => {};
  const CHART = ["users", "revenue", "revenue_per_user"];
  const SESSIONS: MetricsViewSpecMeasure = {
    name: "sessions",
    expression: "approx_count_distinct(session_id)",
    exactExpression: "count(DISTINCT session_id)",
  };

  it("is undefined for a chart without exact-capable measures", () => {
    const ctx = createExactMeasureContext(NONE, MEASURES);
    expect(
      canvasChartPrecisionChip(
        ["revenue"],
        "Revenue",
        ctx,
        false,
        undefined,
        noop,
      ),
    ).toBeUndefined();
  });

  it("is approximate when none of the chart's measures is exact", () => {
    const ctx = createExactMeasureContext(NONE, MEASURES);
    expect(
      canvasChartPrecisionChip(CHART, "Users by day", ctx, true, "boom", noop),
    ).toMatchObject({
      state: "approx",
      measureLabel: "Users by day",
      loading: false,
      error: undefined,
    });
  });

  it("is exact when every exact-capable measure of the chart is exact, with the query status", () => {
    const ctx = createExactMeasureContext(
      new Set(["users", "revenue_per_user"]),
      MEASURES,
    );
    expect(
      canvasChartPrecisionChip(CHART, "Users by day", ctx, true, "boom", noop),
    ).toMatchObject({
      state: "exact",
      loading: true,
      error: "boom",
      onToggle: noop,
    });
  });

  it("counts the measures that are exact through a related measure", () => {
    // Only the ratio is selected; the users it is calculated from are exact too.
    const ctx = createExactMeasureContext(
      new Set(["revenue_per_user"]),
      MEASURES,
    );
    expect(
      canvasChartPrecisionChip(
        CHART,
        "Users by day",
        ctx,
        false,
        undefined,
        noop,
      ),
    ).toMatchObject({ state: "exact" });
  });

  it("is partial when only some exact-capable measures are exact", () => {
    const ctx = createExactMeasureContext(new Set(["users"]), [
      ...MEASURES,
      SESSIONS,
    ]);
    expect(
      canvasChartPrecisionChip(
        ["users", "sessions"],
        "Users, Sessions",
        ctx,
        false,
        undefined,
        noop,
      ),
    ).toMatchObject({ state: "partial" });
  });
});
