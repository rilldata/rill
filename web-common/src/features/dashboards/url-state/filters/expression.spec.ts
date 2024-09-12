import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createOrExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/convertors";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";
import grammar from "./expression.js";
import nearley from "nearley";

const Cases = [
  {
    expr: "country IN ('US','IN') and state = 'ABC'",
    expected_parsed: [
      "AND",
      ["IN", "country", ["US", "IN"]],
      ["=", "state", "ABC"],
    ],
    expected_expression: createAndExpression([
      createInExpression("country", ["US", "IN"]),
      createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
    ]),
  },
  {
    expr: "country IN ('US','IN') and state = 'ABC' and lat >= 12.56",
    expected_parsed: [
      "AND",
      ["IN", "country", ["US", "IN"]],
      ["=", "state", "ABC"],
      [">=", "lat", 12.56],
    ],
    expected_expression: createAndExpression([
      createInExpression("country", ["US", "IN"]),
      createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
      createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
    ]),
  },
  {
    expr: "country IN ('US','IN') AND state = 'ABC' OR lat >= 12.56",
    expected_parsed: [
      "AND",
      ["IN", "country", ["US", "IN"]],
      ["OR", ["=", "state", "ABC"], [">=", "lat", 12.56]],
    ],
    expected_expression: createAndExpression([
      createInExpression("country", ["US", "IN"]),
      createOrExpression([
        createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
        createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
      ]),
    ]),
  },
  {
    expr: "country not in ('US','IN') and (state = 'ABC' or lat >= 12.56)",
    expected_parsed: [
      "AND",
      ["NIN", "country", ["US", "IN"]],
      ["OR", ["=", "state", "ABC"], [">=", "lat", 12.56]],
    ],
    expected_expression: createAndExpression([
      createInExpression("country", ["US", "IN"], true),
      createOrExpression([
        createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
        createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
      ]),
    ]),
  },
  {
    expr: "country NIN ('US','IN') and state having (lat >= 12.56)",
    expected_parsed: [
      "AND",
      ["NIN", "country", ["US", "IN"]],
      ["HAVING", "state", [">=", "lat", 12.56]],
    ],
    expected_expression: createAndExpression([
      createInExpression("country", ["US", "IN"], true),
      createSubQueryExpression(
        "state",
        ["lat"],
        createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
      ),
    ]),
  },
];

describe("expression", () => {
  for (const { expr, expected_parsed, expected_expression } of Cases) {
    it(expr, () => {
      const parser = new nearley.Parser(nearley.Grammar.fromCompiled(grammar));
      parser.feed(expr);
      expect(parser.results).length(1);
      expect(parser.results[0]).toEqual(expected_parsed);

      expect(convertFilterParamToExpression(expr)).toEqual(expected_expression);
    });
  }
});
