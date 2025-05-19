import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import type { DateTimeUnit } from "luxon";
import { describe, expect, it } from "vitest";
import grammar from "./rill-time.cjs";
import nearley from "nearley";

const GRAINS = ["Y", "Q", "M", "W", "D", "H", "m", "s"] as const;
const GRAIN_TO_LUXON: Record<string, DateTimeUnit> = {
  s: "second",
  m: "minute",
  H: "hour",
  D: "day",
  W: "week",
  M: "month",
  Q: "quarter",
  Y: "year",
};

type TestCase = [syntax: string, label?: string, complete?: boolean];

function getSinglePeriodTestCases(): TestCase[] {
  return GRAINS.map((g) => {
    const current = `This ${GRAIN_TO_LUXON[g]}`;
    const previous = `Previous ${GRAIN_TO_LUXON[g]}`;
    const next = `Next ${GRAIN_TO_LUXON[g]}`;
    return <TestCase[]>[
      [`${g}^ to ${g}$`, current, false],
      [`-${g}^ to +${g}$`, current, false],
      [`-0${g}^ to +0${g}$`, current, false],
      [`-1${g}$ to +1${g}^`, current, false],
      [`-0${g}/${g}^ to +0${g}/${g}$`, current, false],
      [`1${g} starting ${g}^`, `1${g} starting ${g}^`, false],
      [`1${g} ending ${g}$`, `1${g} ending ${g}$`, false],
      [`${g}!`, current, false],

      [`-1${g}^ to ${g}^`, previous, true],
      [`-1${g}^ to -1${g}$`, previous, true],
      [`-2${g}$ to -1${g}$`, previous, true],
      [`-1${g}/${g}^ to ${g}/${g}^`, previous, true],
      [`1${g} starting -1${g}^`, `1${g} starting -1${g}^`, false], // TODO: this should be complete
      [`1${g} ending -1${g}$`, `1${g} ending -1${g}$`, true],
      [`-1${g}!`, previous, true],

      [`${g}$ to +1${g}$`, next, false],
      [`+1${g}^ to +1${g}$`, next, false],
      [`+1${g}!`, next, false],
    ];
  }).flat();
}

function getMultiPeriodTestCases(n: number): TestCase[] {
  return GRAINS.map((g) => {
    const last = `Last ${n} ${GRAIN_TO_LUXON[g]}s`;
    const next = `Next ${n} ${GRAIN_TO_LUXON[g]}s`;
    return <TestCase[]>[
      [`-7${g}$ to ${g}$`, last, false],
      [`-7${g}^ to ${g}^`, last, true],

      [`${g}^ to +7${g}^`, next, false],
      [`${g}$ to +7${g}$`, next, false],
    ];
  }).flat();
}

describe("rill time", () => {
  describe("positive cases", () => {
    const Cases: [rillTime: string, label?: string, complete?: boolean][] = [
      ...getSinglePeriodTestCases(),
      ...getMultiPeriodTestCases(7),
      ["-5W4M3Q2Y to -4W3M2Q1Y", "-5W4M3Q2Y to -4W3M2Q1Y", true],
      ["-5W-4M-3Q-2Y to -4W-3M-2Q-1Y", "-5W-4M-3Q-2Y to -4W-3M-2Q-1Y", true],
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

        // const convertedRillTime = rt.toString();
        // const convertedRillTimeParsed = parseRillTime(convertedRillTime);
        // expect(convertedRillTimeParsed.toString()).toEqual(convertedRillTime);
      });
    }
  });
});
