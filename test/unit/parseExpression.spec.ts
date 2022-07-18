import { parseExpression } from "$common/utils/parseExpression";

describe("parseExpression", () => {
  (
    [
      ["count(*)", ["*"]],
      ["avg(a::INT)", ["a"]],
      ["avg(a) - sum(b)", ["a", "b"]],
      ["avg(a) + 100 / 2.5", ["a"]],
    ] as Array<[string, Array<string>]>
  ).forEach((validExpressionTest) => {
    it(`Valid Expression: ${validExpressionTest[0]}`, () => {
      const parsedExpression = parseExpression(validExpressionTest[0]);
      expect(parsedExpression.isValid).toBe(true);
      expect(parsedExpression.columns).toMatchObject(validExpressionTest[1]);
    });
  });

  (
    [
      ["a", "ref", [0, 1]],
      ["a + b", "ref", [4, 5]],
      ["case when a = null then 0 else a end", "case", [0, 36]],
    ] as Array<[string, string, [number, number]]>
  ).forEach((disallowedExpressionTest) => {
    it(`Disallowed Syntax: ${disallowedExpressionTest[0]}`, () => {
      const parsedExpression = parseExpression(disallowedExpressionTest[0]);
      expect(parsedExpression.isValid).toBe(false);
      expect(parsedExpression.error.disallowedSyntax).toBe(
        disallowedExpressionTest[1]
      );
      expect(parsedExpression.error.location.start).toBe(
        disallowedExpressionTest[2][0]
      );
      expect(parsedExpression.error.location.end).toBe(
        disallowedExpressionTest[2][1]
      );
    });
  });

  (
    [
      ["sum(1", [4, 4]],
      ["sum(a+)", [6, 7]],
    ] as Array<[string, [number, number]]>
  ).forEach((errorExpressionTest) => {
    it(`Error syntax: ${errorExpressionTest}`, () => {
      const parsedExpression = parseExpression(errorExpressionTest[0]);
      expect(parsedExpression.isValid).toBe(false);
      expect(parsedExpression.error.location.start).toBe(
        errorExpressionTest[1][0]
      );
      expect(parsedExpression.error.location.end).toBe(
        errorExpressionTest[1][1]
      );
    });
  });
});
