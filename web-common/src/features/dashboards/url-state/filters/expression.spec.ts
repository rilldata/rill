import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createOrExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  convertFilterParamToExpression,
  stripParserError,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";
import grammar from "./expression.cjs";
import nearley from "nearley";

describe("expression", () => {
  describe("positive cases", () => {
    const Cases = [
      {
        expr: "country IN ('US','IN')",
        expected_expression: createInExpression("country", ["US", "IN"]),
      },
      {
        expr: "country IN ('US','IN') and state eq 'ABC'",
        expected_expression: createAndExpression([
          createInExpression("country", ["US", "IN"]),
          createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
        ]),
      },
      {
        expr: "country IN ('US','IN') and state eq 'ABC' and lat gte 12.56",
        expected_expression: createAndExpression([
          createInExpression("country", ["US", "IN"]),
          createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
          createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
        ]),
      },
      {
        expr: "country IN ('US','IN') AND state eq 'ABC' OR lat gte 12.56",
        expected_expression: createAndExpression([
          createInExpression("country", ["US", "IN"]),
          createOrExpression([
            createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
            createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
          ]),
        ]),
      },
      {
        expr: "country not in ('US','IN') and (state eq 'ABC' or lat gte 12.56)",
        expected_expression: createAndExpression([
          createInExpression("country", ["US", "IN"], true),
          createOrExpression([
            createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
            createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
          ]),
        ]),
      },
      {
        expr: "country NIN ('US','IN') and state having (lat gte 12.56)",
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

    const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
    for (const { expr, expected_expression } of Cases) {
      it(expr, () => {
        const parser = new nearley.Parser(compiledGrammar);
        parser.feed(expr);
        // assert that there is only match. this ensures unambiguous grammar.
        expect(parser.results).length(1);

        expect(convertFilterParamToExpression(expr)).to.deep.eq(
          expected_expression,
        );
      });
    }
  });

  describe("negative cases", () => {
    const Cases = [
      {
        expr: "country ('US','IN') and state eq 'ABC'",
        err: `Syntax error at line 1 col 9:

1 country ('US','IN') and state eq 'ABC'
          ^

Unexpected "(".`,
      },
      {
        expr: "country IN (US,'IN') and state eq 'ABC'",
        err: `Syntax error at line 1 col 13:

1 country IN (US,'IN') and state eq 'ABC'
              ^

Unexpected "U".`,
      },
    ];

    for (const { expr, err } of Cases) {
      it(expr, () => {
        let hasError = false;
        try {
          convertFilterParamToExpression(expr);
        } catch (e) {
          hasError = true;
          expect(stripParserError(e)).toEqual(err);
        }
        expect(hasError).to.be.true;
      });
    }
  });
});
