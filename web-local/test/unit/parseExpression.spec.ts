import { parseExpression } from "@rilldata/web-local/common/expression-parser/parseExpression";
import { getMessageFromParseError } from "@rilldata/web-local/common/expression-parser/getMessageFromParseError";
import { TwoTableJoinQueryColumnsTestData } from "../data/ModelQuery.data";
import type { ProfileColumn } from "@rilldata/web-local/lib/types";

describe("parseExpression", () => {
  describe("Valid Expression", () => {
    (
      [
        ["count(*)", ["*"]],
        ["count(distinct a)", ["a"]],
        ["avg(a::INT)", ["a"]],
        ["avg(a) - sum(b)", ["a", "b"]],
        ["avg(a) + 100 / 2.5", ["a"]],
      ] as Array<[string, Array<string>]>
    ).forEach(([expression, expectedColumns]) => {
      it(`Valid Expression: ${expression}`, () => {
        const parsedExpression = parseExpression(expression, []);
        expect(parsedExpression.isValid).toBe(true);
        expect(parsedExpression.columns).toMatchObject(expectedColumns);
      });
    });
  });

  describe("Disallowed Syntax", () => {
    (
      [
        ["a", "ref", [0, 1]],
        ["a + b", "ref", [4, 5]],
        ["case when a = null then 0 else a end", "case", [0, 36]],
      ] as Array<[string, string, [number, number]]>
    ).forEach(([expression, disallowedSyntax, location]) => {
      it(`Disallowed Syntax: ${expression}`, () => {
        const parsedExpression = parseExpression(expression, []);
        expect(parsedExpression.isValid).toBe(false);
        expect(parsedExpression.error.disallowedSyntax).toBe(disallowedSyntax);
        expect(parsedExpression.error.location.start).toBe(location[0]);
        expect(parsedExpression.error.location.end).toBe(location[1]);
      });
    });
  });

  describe("Error syntax", () => {
    (
      [
        ["sum(1", [4, 4], "Unexpected end of input"],
        ["sum(a+)", [6, 7], 'Token ")" is not allowed at "sum(a+`==>)<==`".'],
        [
          "select * from A",
          [0, 6],
          'Token "select" is not allowed at "`==>select<==` * from A".',
        ],
      ] as Array<[string, [number, number], string]>
    ).forEach(([expression, location, expectedMessage]) => {
      it(`Error syntax: ${expression}`, () => {
        const parsedExpression = parseExpression(expression, []);
        expect(parsedExpression.isValid).toBe(false);
        expect(parsedExpression.error.location.start).toBe(location[0]);
        expect(parsedExpression.error.location.end).toBe(location[1]);
        expect(
          getMessageFromParseError(expression, parsedExpression.error)
        ).toBe(expectedMessage);
      });
    });
  });

  describe("Aggregate check with columns", () => {
    (
      [
        ["sm(impressions)", '"sm" is not a valid aggregate function'],
        [
          "sum(publisher)",
          "Invalid aggregate arguments. Expected: sum(NUMERICS) Actual: sum(VARCHAR)",
        ],
        [
          "arg_max(publisher, impressions)",
          "Invalid aggregate arguments. Expected: arg_max(NUMERICS, ANY) Actual: arg_max(VARCHAR, BIGINT)",
        ],
        [
          "approx_quantile(publisher, 1)",
          "Invalid aggregate arguments. Expected: approx_quantile(NUMERICS, UNKNOWN) Actual: approx_quantile(VARCHAR, UNKNOWN)",
        ],
      ] as Array<[string, string]>
    ).forEach(([expression, expectedMessage]) => {
      it(`Invalid aggregates: ${expression}`, () => {
        const parsedExpression = parseExpression(
          expression,
          TwoTableJoinQueryColumnsTestData as unknown as Array<ProfileColumn>
        );
        expect(parsedExpression.isValid).toBe(false);
        expect(
          getMessageFromParseError(expression, parsedExpression.error)
        ).toBe(expectedMessage);
      });
    });

    (
      [
        "sum(impressions)",
        "arg_max(impressions, publisher)",
        "approx_quantile(impressions, 1)",
      ] as Array<string>
    ).forEach((expression) => {
      it(`Valid aggregates: ${expression}`, () => {
        const parsedExpression = parseExpression(
          expression,
          TwoTableJoinQueryColumnsTestData as unknown as Array<ProfileColumn>
        );
        expect(parsedExpression.isValid).toBe(true);
      });
    });
  });
});
