import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { describe, expect, it } from "vitest";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

describe("rill time", () => {
  describe("positive cases", () => {
    const Cases = [
      {
        rillTime: "m : |s|",
        label: "Minute to date, incomplete",
      },
      {
        rillTime: "-5m : |m|",
        label: "Last 5 minutes, incomplete",
      },
      {
        rillTime: "-5m, 0m : |m|",
        label: "Last 5 minutes, complete",
      },
      {
        rillTime: "-7d, 0d : |h|",
        label: "Last 7 days, complete",
      },
      {
        rillTime: "-6d, now : |h|",
        label: "Last 7 days, incomplete",
      },
      {
        rillTime: "-6d, now : h",
        label: "Last 7 days, incomplete",
      },
      {
        rillTime: "d : h",
        label: "Today, incomplete",
      },
    ];

    const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
    for (const { rillTime, label } of Cases) {
      it(rillTime, () => {
        const parser = new nearley.Parser(compiledGrammar);
        parser.feed(rillTime);
        // assert that there is only match. this ensures unambiguous grammar.
        expect(parser.results).length(1);

        const rt = parseRillTime(rillTime);
        expect(
          rt.getLabel({
            completeness: true,
          }),
        ).toEqual(label);
      });
    }
  });
});
