import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  normalizeTimeFilters,
  timeFiltersToParams,
} from "../components/time-filters";
import {
  andControlFilters,
  applyTimeControls,
  controlFiltersExpression,
  withoutTimeControls,
} from "./apply-filters";
import type { AppliedControl, ControlValue } from "./controls";

function filterControl(
  target: string,
  values: ControlValue[],
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

it("retains the primitive types selected in filter controls", () => {
  expect(
    controlFiltersExpression(
      [filterControl("country", [1, true, null, "null"])],
      "sales",
      () => true,
    ),
  ).toEqual(
    createAndExpression([
      createInExpression("country", [1, true, null, "null"]),
    ]),
  );
});

describe("withoutTimeControls", () => {
  const range: AppliedControl = {
    control: { name: "range", kind: "time_range" },
    level: "canvas",
    scope: [],
    urlKey: "c.range",
    targetKey: "time_range:",
    values: ["P7D"],
  };

  // The `time_filters` written back when the widget's time state, given the YAML's time filters
  // with the controls applied, changes to `changed`.
  function writeBack(
    yaml: string,
    controls: AppliedControl[],
    changed: (shown: URLSearchParams) => void,
  ) {
    const shown = timeFiltersToParams(
      applyTimeControls(yaml, controls).toString(),
    );
    const params = new URLSearchParams(shown);
    changed(params);
    return normalizeTimeFilters(
      withoutTimeControls(params, shown, timeFiltersToParams(yaml)),
    );
  }

  it("keeps the control's range out of the YAML when another param changes", () => {
    expect(writeBack("", [range], (p) => p.set("grain", "week"))).toBe(
      "grain=week",
    );
    expect(
      writeBack("tr=P30D&grain=day", [range], (p) => p.set("grain", "week")),
    ).toBe("tr=P30D&grain=week");
  });

  it("writes a range picked over the control's", () => {
    expect(writeBack("", [range], (p) => p.set("tr", "P14D"))).toBe(
      "tr=P14D&compare_tr=inherit",
    );
  });

  it("keeps the zone a control's range names out of the YAML", () => {
    const zoned = {
      ...range,
      values: ["W1 as of watermark tz Asia/Kathmandu"],
    };
    // The time state writes the zone it reads from the range as `tz`.
    expect(
      writeBack("tr=P30D&grain=day", [zoned], (p) => {
        p.set("grain", "week");
        p.set("tz", "Asia/Kathmandu");
      }),
    ).toBe("tr=P30D&grain=week");
  });

  it("writes the change as is without time controls", () => {
    expect(writeBack("tr=P30D", [], (p) => p.set("grain", "week"))).toBe(
      "tr=P30D&grain=week",
    );
  });
});
