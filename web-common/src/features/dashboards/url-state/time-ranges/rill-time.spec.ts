import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { describe, expect, it } from "vitest";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

describe("rill time", () => {
  describe("positive cases", () => {
    const Cases: [rillTime: string, label: string, complete: boolean][] = [
      ["D", "Previous day", true],
      ["D~", "This day", false],
      ["7D", "Last 7 days", true],
      ["7D~", "Last 7 days", false],
      ["-7D", "7 days in the past", true],
      ["+7D", "7 days in the future", true],

      ["D3 of W1 of -1Y", "Day 3 of week 1 of previous year", true],
      ["D3 of W1 of -3Y", "Day 3 of week 1 of 3 years in the past", true],
      ["D3 of W1 of 2024", "Day 3 of week 1 of 2024", true],
      ["D3 of 2024-03", "Day 3 of Mar 2024", true],
      ["H3 of 2024-03-10", "Hour 3 of Mar 10, 2024", true],

      ["WTH", "Week to hour", true],
      ["WTH~", "Week to hour", false],
    ];

    const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
    for (const [rillTime, label, complete] of Cases) {
      it(rillTime, () => {
        const parser = new nearley.Parser(compiledGrammar);
        parser.feed(rillTime);
        // assert that there is only match. this ensures unambiguous grammar.
        expect(parser.results).length(1);

        const rt = parseRillTime(rillTime);
        expect(rt.getLabel()).toEqual(label);
        expect(rt.isComplete).toEqual(complete);

        const convertedRillTime = rt.toString();
        const convertedRillTimeParsed = parseRillTime(convertedRillTime);
        expect(convertedRillTimeParsed.toString()).toEqual(convertedRillTime);
      });
    }
  });
});
