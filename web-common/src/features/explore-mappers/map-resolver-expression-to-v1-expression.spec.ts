import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { AD_BIDS_PUBLISHER_DIMENSION } from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-resolver-expression-to-v1-expression.ts";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import type { Expression } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { describe, expect, it } from "vitest";

describe("mapResolverExpressionToV1Expression", () => {
  const TestCases: {
    title: string;
    expression: Expression;
    expectedExpression: V1Expression;
  }[] = [
    {
      title: "array of values for in expression",
      expression: {
        cond: {
          op: "in",
          exprs: [
            { name: AD_BIDS_PUBLISHER_DIMENSION },
            { val: ["Facebook", "Google"] as any },
          ],
        },
      },
      expectedExpression: createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
        "Facebook",
        "Google",
      ]),
    },
    {
      title: "separate values for in expression",
      expression: {
        cond: {
          op: "in",
          exprs: [
            { name: AD_BIDS_PUBLISHER_DIMENSION },
            { val: "Facebook" as any },
            { val: "Google" as any },
          ],
        },
      },
      expectedExpression: createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
        "Facebook",
        "Google",
      ]),
    },
    {
      title: "eq expression",
      expression: {
        cond: {
          op: "eq",
          exprs: [
            { name: AD_BIDS_PUBLISHER_DIMENSION },
            { val: "Facebook" as any },
          ],
        },
      },
      expectedExpression: createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
        "Facebook",
      ]),
    },
    {
      title: "falsy scalar values are kept",
      expression: {
        cond: {
          op: "in",
          exprs: [
            { name: AD_BIDS_PUBLISHER_DIMENSION },
            { val: 0 as any },
            { val: false as any },
            { val: "" as any },
          ],
        },
      },
      expectedExpression: createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
        0,
        false,
        "",
      ]),
    },
  ];

  for (const { title, expression, expectedExpression } of TestCases) {
    it(title, () => {
      expect(mapResolverExpressionToV1Expression(expression)).toEqual(
        expectedExpression,
      );
    });
  }
});
