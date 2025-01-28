import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { describe, expect, it } from "vitest";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

describe("rill time", () => {
  describe("positive cases", () => {
    const Cases: [rillTime: string, label: string][] = [
      ["m : |s|", "Minute to date, incomplete"],
      ["-5m : |m|", "Last 5 minutes, incomplete"],
      ["-5m, 0m : |m|", "Last 5 minutes"],
      ["-7d, 0d : |h|", "Last 7 days"],
      ["-7d, now/d : |h|", "Last 7 days"],
      ["-6d, now : |h|", "Last 7 days, incomplete"],
      ["-6d, now : h", "Last 7 days, incomplete"],
      ["d : h", "Today, incomplete"],

      // TODO: correct label for the below
      ["-7d, -5d : h", "-7d, -5d : h"],
      ["-2d, now/d : h @ -5d", "-2d, now/d : h @ -5d"],
      ["-2d, now/d @ -5d", "-2d, now/d @ -5d"],

      [
        "-7d, now/d : h @ {Asia/Kathmandu}",
        "-7d, now/d : h @ {Asia/Kathmandu}",
      ],
      [
        "-7d, now/d : |h| @ {Asia/Kathmandu}",
        "-7d, now/d : |h| @ {Asia/Kathmandu}",
      ],
      [
        "-7d, now/d : |h| @ -5d {Asia/Kathmandu}",
        "-7d, now/d : |h| @ -5d {Asia/Kathmandu}",
      ],

      // TODO: should these be something different when end is latest vs now?
      ["-7d, latest/d : |h|", "Last 7 days"],
      ["-6d, latest : |h|", "Last 6 days, incomplete"],
      ["-6d, latest : h", "Last 6 days, incomplete"],

      ["2024-01-01, 2024-03-31", "2024-01-01, 2024-03-31"],
      [
        "2024-01-01 10:00, 2024-03-31 18:00",
        "2024-01-01 10:00, 2024-03-31 18:00",
      ],

      ["-7W+2d", "-7W+2d"],
      ["-7W,latest+2d", "-7W,latest+2d"],
      ["-7W,-7W+2d", "-7W,-7W+2d"],
      ["-7W+2d", "-7W+2d"],
      ["2024-01-01-1W,2024-01-01", "2024-01-01-1W,2024-01-01"],
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
      });
    }
  });
});
