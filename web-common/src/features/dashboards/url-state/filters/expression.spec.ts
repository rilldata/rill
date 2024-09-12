import { describe, it, expect } from "vitest";
import grammar from "./expression.js";
import nearley from "nearley";

describe("expression", () => {
  const Cases = [
    {
      expr: "country IN ('US','IN') and state = 'ABC'",
      expected: ["AND", ["IN", "country", ["US", "IN"]], ["=", "state", "ABC"]],
    },
    {
      expr: "country IN ('US','IN') and state = 'ABC' and lat >= 12.56",
      expected: [
        "AND",
        ["IN", "country", ["US", "IN"]],
        ["=", "state", "ABC"],
        [">=", "lat", 12.56],
      ],
    },
    {
      expr: "country not in ('US','IN') and (state = 'ABC' or lat >= 12.56)",
      expected: [
        "AND",
        ["NIN", "country", ["US", "IN"]],
        ["OR", ["=", "state", "ABC"], [">=", "lat", 12.56]],
      ],
    },
    {
      expr: "country NIN ('US','IN') and state having (lat >= 12.56)",
      expected: [
        "AND",
        ["NIN", "country", ["US", "IN"]],
        ["HAVING", "state", [">=", "lat", 12.56]],
      ],
    },
  ];

  for (const { expr, expected } of Cases) {
    it(expr, () => {
      const parser = new nearley.Parser(nearley.Grammar.fromCompiled(grammar));
      parser.feed(expr);
      console.log(JSON.stringify(parser.results, null, 2));
      expect(parser.results).length(1);
      expect(parser.results[0]).toEqual(expected);
    });
  }
});
