import nearley from "nearley";
import { describe, it, expect } from "vitest";
import grammar from "./measure-expression.js";
import {
  EPHEMERAL_MEASURE_FUNCTIONS,
  MAX_EPHEMERAL_EXPRESSION_LENGTH,
  formatMeasureRef,
  parseMeasureExpression,
} from "./expression-parser";

describe("parseMeasureExpression", () => {
  describe("accepts", () => {
    const cases: Array<[string, string[]]> = [
      ["revenue - cost", ["revenue", "cost"]],
      ["revenue-cost", ["revenue", "cost"]],
      ["(a + b) * 2", ["a", "b"]],
      ["-a", ["a"]],
      ["a % b", ["a", "b"]],
      ["a * -1", ["a"]],
      ["power(a, 2)", ["a"]],
      ["coalesce(a, 0)", ["a"]],
      ["coalesce(a, NULL)", ["a"]],
      ["nullif(cost, 0)", ["cost"]],
      ["round(a / b, 2)", ["a", "b"]],
      [
        "abs(a) + floor(b) + ceil(c) + sqrt(d) + ln(e) + exp(f)",
        ["a", "b", "c", "d", "e", "f"],
      ],
      ["greatest(a, b, c)", ["a", "b", "c"]],
      ['"count" - a', ["count", "a"]],
      ['"weird""name" + 1', ['weird"name']],
      ["1e6 * a", ["a"]],
      ["0.4 * revenue", ["revenue"]],
      ["revenue - revenue", ["revenue"]],
      ["((((a))))", ["a"]],
      ['"null" + 1', ["null"]],
      ["ROUND(a, 2)", ["a"]],
      ['"rank" + a', ["rank", "a"]],
      ["true + a", ["a"]],
      ["a\r+\rb", ["a", "b"]],
      ["1. * a", ["a"]],
      [".5 * a", ["a"]],
      ["1.e5 * a", ["a"]],
      ["  a  ", ["a"]],
    ];
    it.each(cases)("%s", (expr, refs) => {
      const res = parseMeasureExpression(expr);
      expect(res.error).toBeUndefined();
      expect(res.refs).toEqual(refs);
    });
  });

  describe("rejects", () => {
    const cases: Array<[string, string]> = [
      ["", "empty"],
      ["   ", "empty"],
      ["revenue); DROP TABLE x;--", "unexpected"],
      ["a; SELECT 1", "unexpected"],
      ["(select secret from t)", "reserved SQL word"],
      ["sum(revenue)", 'unsupported function "sum"'],
      ["count(a)", 'unsupported function "count"'],
      ["t.revenue", "unexpected character"],
      ["a > b", "unexpected"],
      ["a and b", "reserved SQL word"],
      ["'str'", "string literals"],
      ["nullif(a, 'x')", "string literals"],
      ["a || b", "unexpected"],
      ["@@version", "unexpected character"],
      ["foo(a)", 'unsupported function "foo"'],
      ["round(a, 1, 2)", "does not accept 3 argument"],
      ["coalesce(a)", "does not accept 1 argument"],
      ["100", "must reference at least one measure"],
      ["1 + 2", "must reference at least one measure"],
      ["NULL", "must reference at least one measure"],
      ["a as profit", "reserved SQL word"],
      ["a /* comment */ + b", "unexpected"],
      ['"unterminated', "unterminated quoted identifier"],
      ["a +", "unexpected end of expression"],
      ["(a", "expected closing parenthesis"],
      ["revenue -- cost", "comments are not allowed"],
      ["rank + 1", "reserved SQL word"],
      ["if(a)", "reserved SQL word"],
      ['"" + a', "empty quoted identifier"],
      ["1e + a", 'invalid number "1e"'],
      ["a * 2.5e+", 'invalid number "2.5e+"'],
      ["(", "expected closing parenthesis"],
      ["a)", 'unexpected ")"'],
      ['"a" "b"', 'unexpected "b"'],
      ["round()", "does not accept 0 argument"],
      ["a.b", "unexpected character"],
    ];
    it.each(cases)("%s", (expr, message) => {
      const res = parseMeasureExpression(expr);
      expect(res.error).toBeDefined();
      expect(res.error!.message).toContain(message);
    });
  });

  it("reports error positions", () => {
    const cases: Array<[string, number]> = [
      ["revenue - 'oops'", 10],
      ["a + select", 4],
      ["a + b @ c", 6],
      ["a + foo(b)", 4],
      ["a + round(b, 1, 2)", 4],
      ["(a + b", 6],
      ["a +", 3],
    ];
    for (const [expr, position] of cases) {
      expect(parseMeasureExpression(expr).error?.position, expr).toBe(position);
    }
  });

  // Lexical problems are found by a pre-scan of the whole input, so make sure
  // they never mask a grammar error that appears earlier in the expression.
  it("reports the first error in reading order", () => {
    expect(parseMeasureExpression('coalesce true "x').error).toEqual({
      message: 'unexpected "true"',
      position: 9,
    });
    expect(parseMeasureExpression("'x' + select").error?.message).toContain(
      "string literals",
    );
  });

  it("parses without ambiguity", () => {
    const compiled = nearley.Grammar.fromCompiled(grammar);
    for (const expr of [
      "revenue - cost",
      "a - -b",
      "round(a / b, 2) * -1.5e3 + (c % d)",
      'coalesce("a b", NULL, .5)',
      "1.",
      "greatest(a, b, c) - least(a, b)",
    ]) {
      const parser = new nearley.Parser(compiled);
      parser.feed(expr);
      expect(parser.results, expr).toHaveLength(1);
    }
  });

  it("rejects deeply nested expressions", () => {
    const expr = "(".repeat(100) + "a" + ")".repeat(100);
    const res = parseMeasureExpression(expr);
    expect(res.error).toBeDefined();
    expect(res.error!.message).toContain("deeply nested");
  });

  // Mirrors the server parser (maxMeasureExpressionDepth = 32), which fails
  // flat chains of more than 32 binary operators with "too deeply nested".
  it("limits total expression complexity like the server", () => {
    const ok = Array.from({ length: 33 }, (_, i) => `a${i}`).join("+");
    expect(parseMeasureExpression(ok).error).toBeUndefined();
    const tooLong = Array.from({ length: 34 }, (_, i) => `a${i}`).join("+");
    expect(parseMeasureExpression(tooLong).error?.message).toContain(
      "deeply nested",
    );
  });

  it("quotes measure refs that need it", () => {
    expect(formatMeasureRef("revenue")).toBe("revenue");
    expect(formatMeasureRef("rank")).toBe('"rank"');
    expect(formatMeasureRef("true")).toBe('"true"');
    expect(formatMeasureRef("total volume")).toBe('"total volume"');
    expect(formatMeasureRef('weird"name')).toBe('"weird""name"');
  });

  it("rejects expressions over the maximum length", () => {
    const expr = "a+".repeat(MAX_EPHEMERAL_EXPRESSION_LENGTH / 2) + "a";
    const res = parseMeasureExpression(expr);
    expect(res.error).toBeDefined();
    expect(res.error!.message).toContain("maximum length");
  });

  // The function allowlist must stay in sync with the server's
  // measureExpressionFuncs in runtime/metricsview/measure_expression.go.
  it("matches the server function allowlist", () => {
    expect(Object.keys(EPHEMERAL_MEASURE_FUNCTIONS).sort()).toEqual([
      "abs",
      "ceil",
      "coalesce",
      "exp",
      "floor",
      "greatest",
      "least",
      "ln",
      "nullif",
      "power",
      "round",
      "sqrt",
    ]);
  });
});
