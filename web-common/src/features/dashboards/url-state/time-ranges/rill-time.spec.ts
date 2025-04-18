import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { describe, expect, it } from "vitest";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

describe("rill time", () => {
  describe("positive cases", () => {
    const Cases: [rillTime: string, label: string][] = [
      ["D3 of W1 of -1Y", "Day 3 of Week 1 of Previous year"],
      ["D3 of W1 of -3Y", "Day 3 of Week 1 of 3 years in the past"],
    ];

    const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
    for (const [rillTime, label] of Cases) {
      it(rillTime, () => {
        const parser = new nearley.Parser(compiledGrammar);
        parser.feed(rillTime);
        // assert that there is only match. this ensures unambiguous grammar.
        expect(parser.results).length(1);

        const rt = parseRillTime(rillTime);
        expect(rt.getLabel()).toEqual(label);

        const convertedRillTime = rt.toString();
        const convertedRillTimeParsed = parseRillTime(convertedRillTime);
        expect(convertedRillTimeParsed.toString()).toEqual(convertedRillTime);
      });
    }
  });
});
