import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-resolver-expression-to-v1-expression";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  convertV1ExpressionToMapstructure,
  metricsSQLUsesTimeRange,
} from "./expression-utils";

describe("convertV1ExpressionToMapstructure", () => {
  it("converts a measure filter's subquery to the shape of Go's metricsview.Subquery", () => {
    const expression = createAndExpression([
      createInExpression("country", ["US"]),
      createSubQueryExpression(
        "publisher",
        ["impressions"],
        createAndExpression([
          createBinaryExpression("impressions", V1Operation.OPERATION_GT, 10),
        ]),
      ),
    ]);

    const converted = convertV1ExpressionToMapstructure(expression);
    expect(converted).toEqual({
      cond: {
        op: "and",
        exprs: [
          {
            cond: {
              op: "in",
              exprs: [{ name: "country" }, { val: "US" }],
            },
          },
          {
            cond: {
              op: "in",
              exprs: [
                { name: "publisher" },
                {
                  subquery: {
                    dimension: { name: "publisher" },
                    measures: [{ name: "impressions" }],
                    having: {
                      cond: {
                        op: "and",
                        exprs: [
                          {
                            cond: {
                              op: "gt",
                              exprs: [{ name: "impressions" }, { val: 10 }],
                            },
                          },
                        ],
                      },
                    },
                  },
                },
              ],
            },
          },
        ],
      },
    });
    // The resolver shape reads back as the expression it was converted from.
    expect(mapResolverExpressionToV1Expression(converted)).toEqual(expression);
  });
});

describe("metricsSQLUsesTimeRange", () => {
  const timeless = new Map([["products", false]]);
  const mixed = new Map([
    ["products", false],
    ["orders", true],
  ]);

  it.each([
    {
      name: "a dashboard without a time dimension never resolves a time range",
      views: ["products"],
      dashboard: timeless,
      expected: false,
    },
    {
      name: "a metrics view without a time dimension is not filtered by another one's",
      views: ["products"],
      dashboard: mixed,
      expected: false,
    },
    {
      name: "one query on a metrics view with a time dimension is enough",
      views: ["products", "orders"],
      dashboard: mixed,
      expected: true,
    },
    {
      name: "an unknown metrics view follows a dashboard with a time dimension",
      views: [undefined],
      dashboard: mixed,
      expected: true,
    },
    {
      name: "an unknown metrics view does not wait on a dashboard without a time dimension",
      views: [undefined, "missing"],
      dashboard: timeless,
      expected: false,
    },
  ])("$name", ({ views, dashboard, expected }) => {
    expect(metricsSQLUsesTimeRange(views, dashboard)).toBe(expected);
  });
});
