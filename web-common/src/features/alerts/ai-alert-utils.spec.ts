import {
  generateAIAlertName,
  getAIAlertResolverProps,
  getAlertAggregationRequest,
} from "@rilldata/web-common/features/alerts/ai-alert-utils";
import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import { describe, expect, it } from "vitest";

const Condition = "Revenue dropped unusually against the previous weeks";

function getFormValues(overrides: Partial<AlertFormValues>): AlertFormValues {
  return {
    aiCondition: `  ${Condition}  `,
    exploreName: "orders",
    measure: "",
    splitByDimension: "",
    ...overrides,
  } as AlertFormValues;
}

function getTimeFilterManager(
  overrides: Partial<TimeFilterManager>,
): TimeFilterManager {
  return {
    requestedTimeRange: undefined,
    timeDimension: undefined,
    timeZone: "UTC",
    ...overrides,
  } as TimeFilterManager;
}

describe("getAIAlertResolverProps", () => {
  it("writes the condition and the data it is narrowed to, which the alert reads back", () => {
    const where = createAndExpression([createInExpression("country", ["ES"])]);
    const props = getAIAlertResolverProps(
      getFormValues({ measure: "revenue", splitByDimension: "country" }),
      where,
      getTimeFilterManager({
        requestedTimeRange: "4W as of watermark/D+1D",
        timeDimension: "ordered_at",
        timeZone: "Europe/Madrid",
      }),
    );
    expect(props).toEqual({
      prompt: Condition,
      explore: "orders",
      measures: ["revenue"],
      dimensions: ["country"],
      where: {
        cond: {
          op: "and",
          exprs: [
            { cond: { op: "in", exprs: [{ name: "country" }, { val: "ES" }] } },
          ],
        },
      },
      time_range: {
        expression: "4W as of watermark/D+1D",
        time_dimension: "ordered_at",
      },
      time_zone: "Europe/Madrid",
    });

    expect(
      getAlertAggregationRequest(
        { resolver: "ai", resolverProperties: props },
        "orders_mv",
      ),
    ).toEqual({
      metricsView: "orders_mv",
      measures: [{ name: "revenue" }],
      dimensions: [{ name: "country" }],
      where,
      timeRange: {
        expression: "4W as of watermark/D+1D",
        timeDimension: "ordered_at",
        timeZone: "Europe/Madrid",
      },
    });
  });

  it("leaves out the data that does not narrow anything", () => {
    expect(
      getAIAlertResolverProps(
        getFormValues({}),
        createAndExpression([]),
        getTimeFilterManager({}),
      ),
    ).toEqual({ prompt: Condition, explore: "orders" });
  });
});

describe("generateAIAlertName", () => {
  it("shortens a long condition to a single line", () => {
    const name = generateAIAlertName(`${Condition},\n  ${Condition}`);
    expect(name).toHaveLength(60);
    expect(name).toBe(
      "Revenue dropped unusually against the previous weeks, Reven…",
    );
  });
});
