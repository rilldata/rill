import { describe, it, expect } from "vitest";
import grammar from "./expression.js";
import nearley from "nearley";

describe("expression", () => {
  const Cases = [
    {
      expr: "country IN ('US','IN') and state = 'ABC'",
      expected: [["country", "IN", ["US", "IN"]], "and", ["state", "=", "ABC"]],
    },
    {
      expr: "country NIN ('US','IN') and (state = 'ABC' or lat >= 12.56)",
      expected: [
        ["country", "NIN", ["US", "IN"]],
        "and",
        [["state", "=", "ABC"], "or", ["lat", ">=", 12.56]],
      ],
    },
    {
      expr: "country NIN ('US','IN') and state having (lat >= 12.56)",
      expected: [
        ["country", "NIN", ["US", "IN"]],
        "and",
        ["state", "having", ["lat", ">=", 12.56]],
      ],
    },
  ];

  for (const { expr, expected } of Cases) {
    it(expr, () => {
      const parser = new nearley.Parser(nearley.Grammar.fromCompiled(grammar));
      parser.feed(expr);
      expect(parser.results).length(1);
      expect(parser.results[0]).toEqual(expected);
    });
  }
});
