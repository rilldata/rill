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
    expect(canvasExactRequestNames(["users"], NONE)).toEqual([]);
  });

  it("flags the surface's selection among the measures it shows", () => {
    expect(
      canvasExactRequestNames(
        ["users", "revenue", "revenue_per_user"],
        new Set(["users", "sessions"]),
      ),
    ).toEqual(["users"]);
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

  it("is undefined for a derived measure: it has no exact variant of its own", () => {
    const ctx = createExactMeasureContext(new Set(["users"]), MEASURES);
    expect(
      canvasPrecisionChip(
        "revenue_per_user",
        "Revenue per user",
        ctx,
        false,
        undefined,
        noop,
      ),
    ).toBeUndefined();
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
    const ctx = createExactMeasureContext(new Set(["users"]), MEASURES);
    expect(
      canvasChartPrecisionChip(CHART, "Users by day", ctx, true, "boom", noop),
    ).toMatchObject({
      state: "exact",
      loading: true,
      error: "boom",
      onToggle: noop,
    });
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
