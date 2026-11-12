import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { andControlFilters, controlFiltersExpression } from "./apply-filters";
import type { AppliedControl } from "./controls";

function filterControl(
  target: string,
  values: string[],
  fields?: Record<string, string>,
): AppliedControl {
  return {
    control: { name: target, kind: "filter", target, fields },
    level: "canvas",
    scope: [],
    urlKey: `c.${target}`,
    targetKey: `filter:${target}`,
    values,
  };
}

describe("controlFiltersExpression", () => {
  const hasDimension = (name: string) =>
    ["country", "billing_country"].includes(name);

  it("filters the target dimension, or the one mapped for the metrics view", () => {
    const controls = [
      filterControl("country", ["US"], { sales: "billing_country" }),
    ];
    expect(controlFiltersExpression(controls, "bids", hasDimension)).toEqual(
      createAndExpression([createInExpression("country", ["US"])]),
    );
    expect(controlFiltersExpression(controls, "sales", hasDimension)).toEqual(
      createAndExpression([createInExpression("billing_country", ["US"])]),
    );
  });

  it("skips controls without values and dimensions the metrics view lacks", () => {
    expect(
      controlFiltersExpression(
        [filterControl("country", []), filterControl("device", ["phone"])],
        "bids",
        hasDimension,
      ),
    ).toBeUndefined();
  });
});

describe("andControlFilters", () => {
  const control = createAndExpression([createInExpression("country", ["US"])]);

  it("keeps every predicate of the existing filter", () => {
    const age = createBinaryExpression("age", V1Operation.OPERATION_GT, 18);
    const existing = createAndExpression([
      age,
      createInExpression("country", ["US", "CA"]),
    ]);
    expect(andControlFilters(existing, control)).toEqual(
      createAndExpression([
        age,
        createInExpression("country", ["US", "CA"]),
        createInExpression("country", ["US"]),
      ]),
    );
    // The inputs are not modified.
    expect(existing.cond?.exprs).toHaveLength(2);
  });

  it("wraps a filter that isn't a conjunction", () => {
    const either = createOrExpression([
      createInExpression("device", ["phone"]),
      createInExpression("device", ["tablet"]),
    ]);
    expect(andControlFilters(either, control)).toEqual(
      createAndExpression([either, createInExpression("country", ["US"])]),
    );
  });

  it("is the control filter when there is no other filter", () => {
    expect(andControlFilters(undefined, control)).toBe(control);
  });
});
