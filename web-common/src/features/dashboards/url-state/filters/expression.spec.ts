import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createLikeExpression,
  createOrExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  convertExpressionToFilterParam,
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
        expectedExprString: "country IN ('US','IN')",
        expectedExprObject: createInExpression("country", ["US", "IN"]),
      },
      {
        expr: "country IN ('US','IN') and state eq 'ABC'",
        expectedExprString: "country IN ('US','IN') AND state EQ 'ABC'",
        expectedExprObject: createAndExpression([
          createInExpression("country", ["US", "IN"]),
          createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
        ]),
      },
      {
        expr: "country IN ('US','IN') and state eq 'ABC' and lat gte 12.56",
        expectedExprString:
          "country IN ('US','IN') AND state EQ 'ABC' AND lat GTE 12.56",
        expectedExprObject: createAndExpression([
          createInExpression("country", ["US", "IN"]),
          createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
          createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
        ]),
      },
      {
        expr: "country IN ('US','IN') AND state eq 'ABC' OR lat gte 12.56",
        expectedExprString:
          "country IN ('US','IN') AND (state EQ 'ABC' OR lat GTE 12.56)",
        expectedExprObject: createAndExpression([
          createInExpression("country", ["US", "IN"]),
          createOrExpression([
            createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
            createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
          ]),
        ]),
      },
      {
        expr: "country not in ('US','IN') and (state eq 'ABC' or lat gte 12.56)",
        expectedExprString:
          "country NIN ('US','IN') AND (state EQ 'ABC' OR lat GTE 12.56)",
        expectedExprObject: createAndExpression([
          createInExpression("country", ["US", "IN"], true),
          createOrExpression([
            createBinaryExpression("state", V1Operation.OPERATION_EQ, "ABC"),
            createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
          ]),
        ]),
      },
      {
        expr: "country NIN ('US','IN') and state having (lat gte 12.56)",
        expectedExprString:
          "country NIN ('US','IN') AND state having (lat GTE 12.56)",
        expectedExprObject: createAndExpression([
          createInExpression("country", ["US", "IN"], true),
          createSubQueryExpression(
            "state",
            ["lat"],
            createBinaryExpression("lat", V1Operation.OPERATION_GTE, 12.56),
          ),
        ]),
      },
      {
        expr: `"coun tr.y" IN ('U\\'S','I\\nN') and "st ate" having ("la t" gte 12.56)`,
        expectedExprString: `"coun tr.y" IN ('U\\'S','I\\nN') AND "st ate" having ("la t" GTE 12.56)`,
        expectedExprObject: createAndExpression([
          // values converted to V1Expression do not have escaped chars
          createInExpression("coun tr.y", ["U'S", "I\nN"]),
          createSubQueryExpression(
            "st ate",
            ["la t"],
            createBinaryExpression("la t", V1Operation.OPERATION_GTE, 12.56),
          ),
        ]),
      },
      {
        expr: `country IN (['U\\'S','I\\nN'])`,
        expectedExprString: `country IN (['U\\'S','I\\nN'])`,
        expectedExprObject: createInExpression("country", [["U'S", "I\nN"]]),
      },
      {
        expr: `country IN ({'US':['SF','LA'],'IN':['BLR','DLH'],'UK':'London'})`,
        expectedExprString: `country IN ({'US':['SF','LA'],'IN':['BLR','DLH'],'UK':'London'})`,
        expectedExprObject: createInExpression("country", [
          { US: ["SF", "LA"], IN: ["BLR", "DLH"], UK: "London" },
        ]),
      },

      {
        expr: "country MATCH ('US','IN')",
        expectedExprString: "country MATCH ('US','IN')",
        expectedExprObject: {
          ...createInExpression("country", ["US", "IN"]),
          isMatchList: true,
        },
      },

      {
        expr: "country like '%oo%'",
        expectedExprString: "country LIKE '%oo%'",
        expectedExprObject: createLikeExpression("country", "%oo%"),
      },
    ];

    const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
    for (const { expr, expectedExprString, expectedExprObject } of Cases) {
      it(expr, () => {
        const parser = new nearley.Parser(compiledGrammar);
        parser.feed(expr);
        // assert that there is only match. this ensures unambiguous grammar.
        expect(parser.results).length(1);

        const exprObject = convertFilterParamToExpression(expr);
        expect(exprObject).to.deep.eq(expectedExprObject);
        expect(convertExpressionToFilterParam(exprObject)).toEqual(
          expectedExprString,
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
