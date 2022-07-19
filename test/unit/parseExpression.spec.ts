import { parseExpression } from "$common/expression-parser/parseExpression";
import { getMessageFromParseError } from "$common/expression-parser/getMessageFromParseError";

describe("parseExpression", () => {
  (
    [
      ["count(*)", ["*"]],
      ["avg(a::INT)", ["a"]],
      ["avg(a) - sum(b)", ["a", "b"]],
      ["avg(a) + 100 / 2.5", ["a"]],
    ] as Array<[string, Array<string>]>
  ).forEach(([expression, expectedColumns]) => {
    it(`Valid Expression: ${expression}`, () => {
      const parsedExpression = parseExpression(expression);
      expect(parsedExpression.isValid).toBe(true);
      expect(parsedExpression.columns).toMatchObject(expectedColumns);
    });
  });

  (
    [
      ["a", "ref", [0, 1]],
      ["a + b", "ref", [4, 5]],
      ["case when a = null then 0 else a end", "case", [0, 36]],
    ] as Array<[string, string, [number, number]]>
  ).forEach(([expression, disallowedSyntax, location]) => {
    it(`Disallowed Syntax: ${expression}`, () => {
      const parsedExpression = parseExpression(expression);
      expect(parsedExpression.isValid).toBe(false);
      expect(parsedExpression.error.disallowedSyntax).toBe(disallowedSyntax);
      expect(parsedExpression.error.location.start).toBe(location[0]);
      expect(parsedExpression.error.location.end).toBe(location[1]);
    });
  });

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
      const parsedExpression = parseExpression(expression);
      expect(parsedExpression.isValid).toBe(false);
      expect(parsedExpression.error.location.start).toBe(location[0]);
      expect(parsedExpression.error.location.end).toBe(location[1]);
      expect(getMessageFromParseError(expression, parsedExpression.error)).toBe(
        expectedMessage
      );
    });
  });
});
