import {
  getAggregationRequestFromAIAlertProperties,
  getAIAlertResolverProperties,
} from "@rilldata/web-common/features/alerts/ai-alert-utils";
import { getExistingAlertInitialFormValues } from "@rilldata/web-common/features/alerts/extract-alert-form-values";
import {
  alertFormValidationSchema,
  type AlertFormValues,
  checkIsTabValid,
  getAlertResolverOptions,
} from "@rilldata/web-common/features/alerts/form-utils";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  V1Operation,
  type V1AlertSpec,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const CountryIsSpain: V1Expression = {
  cond: {
    op: V1Operation.OPERATION_AND,
    exprs: [
      {
        cond: {
          op: V1Operation.OPERATION_IN,
          exprs: [{ ident: "country" }, { val: "Spain" }],
        },
      },
    ],
  },
};

// An AI alert as the API creates it: no dashboard annotations, only the ai resolver's properties.
const AIAlertSpec: V1AlertSpec = {
  displayName: "Revenue drop",
  resolver: "ai",
  resolverProperties: {
    prompt: "Revenue dropped unusually against the previous weeks",
    explore: "orders",
    measures: ["revenue"],
    comparison_time_range: { expression: "-4W to ref" },
  },
  annotations: {},
  notifiers: [
    { connector: "email", properties: { recipients: ["a@example.com"] } },
  ],
};

describe("getAIAlertResolverProperties", () => {
  it("writes the condition, the explore and the context that narrows it, and reads them back", () => {
    const props = getAIAlertResolverProperties({
      condition: "  Revenue dropped unusually against the previous weeks  ",
      exploreName: "orders",
      measure: "revenue",
      splitByDimension: "country",
      where: CountryIsSpain,
      selectedTimeRange: { name: "P4W" },
      timeZone: "Europe/Madrid",
    });
    expect(props).toEqual({
      prompt: "Revenue dropped unusually against the previous weeks",
      explore: "orders",
      measures: ["revenue"],
      dimensions: ["country"],
      where: {
        cond: {
          op: "and",
          exprs: [
            {
              cond: {
                op: "in",
                exprs: [{ name: "country" }, { val: "Spain" }],
              },
            },
          ],
        },
      },
      time_range: { expression: "P4W" },
      time_zone: "Europe/Madrid",
    });

    expect(
      getAggregationRequestFromAIAlertProperties(props, "orders_mv"),
    ).toEqual({
      metricsView: "orders_mv",
      measures: [{ name: "revenue" }],
      dimensions: [{ name: "country" }],
      where: CountryIsSpain,
      timeRange: { isoDuration: "P4W", timeZone: "Europe/Madrid" },
    });
  });

  it("leaves out the context that does not narrow anything", () => {
    expect(
      getAIAlertResolverProperties({
        condition: "Revenue dropped",
        exploreName: "orders",
        measure: "",
        splitByDimension: "",
        where: { cond: { op: V1Operation.OPERATION_AND, exprs: [] } },
        selectedTimeRange: { name: TimeRangePreset.ALL_TIME },
        timeZone: "UTC",
      }),
    ).toEqual({ prompt: "Revenue dropped", explore: "orders" });
  });

  it("keeps the properties the dialog does not edit", () => {
    const props = getAIAlertResolverProperties({
      condition: "Revenue dropped",
      exploreName: "orders",
      measure: "revenue",
      splitByDimension: "",
      where: undefined,
      selectedTimeRange: undefined,
      timeZone: "UTC",
      previous: {
        prompt: "old condition",
        measures: ["revenue", "orders"],
        comparison_time_range: { expression: "-4W to ref" },
        time_zone: "Europe/Madrid",
      },
    });
    expect(props).toEqual({
      prompt: "Revenue dropped",
      explore: "orders",
      // The dialog shows the first measure; the rest are kept while it is unchanged
      measures: ["revenue", "orders"],
      comparison_time_range: { expression: "-4W to ref" },
      // Without a time range, the dialog has no time zone to write, but the comparison is still resolved in this one
      time_zone: "Europe/Madrid",
    });
  });
});

describe("the alert form with an AI alert", () => {
  it("loads the condition and writes the explore's default time range explicitly", () => {
    const values = getExistingAlertInitialFormValues(AIAlertSpec, "orders_mv");
    expect(values.conditionType).toBe("ai");
    expect(values.aiCondition).toBe(
      "Revenue dropped unusually against the previous weeks",
    );
    expect(values.exploreName).toBe("orders");
    expect(values.measure).toBe("revenue");

    // What the user sees is what is judged, even if the explore's default changes later
    const options = getAlertResolverOptions(
      values,
      undefined,
      {
        selectedTimeRange: { name: TimeRangePreset.DEFAULT },
        selectedTimezone: "Europe/Madrid",
      } as unknown as TimeControlState,
      { defaultPreset: { timeRange: "P14D" } },
      AIAlertSpec,
    );
    expect(options).toEqual({
      resolver: "ai",
      resolverProperties: {
        prompt: "Revenue dropped unusually against the previous weeks",
        explore: "orders",
        measures: ["revenue"],
        comparison_time_range: { expression: "-4W to ref" },
        time_range: { expression: "P14D" },
        time_zone: "Europe/Madrid",
      },
    });
  });

  it("requires a condition instead of a measure and criteria", async () => {
    const values = {
      ...getExistingAlertInitialFormValues(AIAlertSpec, "orders_mv"),
      measure: "",
      name: "Revenue drop",
    } as AlertFormValues;
    expect((await alertFormValidationSchema.validate(values)).success).toBe(
      true,
    );
    expect(checkIsTabValid(0, values, {})).toBe(true);
    expect(checkIsTabValid(1, values, {})).toBe(true);

    const noCondition = { ...values, aiCondition: "  " };
    expect(
      (await alertFormValidationSchema.validate(noCondition)).success,
    ).toBe(false);
    expect(checkIsTabValid(1, noCondition, {})).toBe(false);

    const criteria = { ...values, conditionType: "criteria" } as const;
    expect((await alertFormValidationSchema.validate(criteria)).success).toBe(
      false,
    );
    expect(checkIsTabValid(0, criteria, {})).toBe(false);
  });
});
