import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { GrainAliasToV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
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

type TestCase = [
  syntax: string,
  label: string,
  complete: boolean,
  rangeGrain: V1TimeGrain | undefined,
  byGrain: V1TimeGrain | undefined,
];

function getSinglePeriodTestCases(): TestCase[] {
  return GRAINS.map((g) => {
    const protoGrain = GrainAliasToV1TimeGrain[g];

    const current = `This ${GRAIN_TO_LUXON[g]}`;
    const previous = `Previous ${GRAIN_TO_LUXON[g]}`;
    const next = `Next ${GRAIN_TO_LUXON[g]}`;

    return <TestCase[]>[
      [`ref/${g} to ref/${g}+1${g}`, current, false, protoGrain, undefined],
      [`-${g}^ to +${g}$`, current, false, protoGrain, undefined],
      [`-0${g}^ to +0${g}$`, current, false, protoGrain, undefined],
      [`-1${g}$ to +1${g}^`, current, false, protoGrain, undefined],
      [`-0${g}/${g}^ to +0${g}/${g}$`, current, false, protoGrain, undefined],
      [`1${g}`, current, false, protoGrain, undefined],

      [`-1${g}^ to ${g}^`, previous, true, protoGrain, undefined],
      [`-1${g}^ to -1${g}$`, previous, true, protoGrain, undefined],
      [`-2${g}$ to -1${g}$`, previous, true, protoGrain, undefined],
      [`-1${g}/${g}^ to ${g}/${g}^`, previous, true, protoGrain, undefined],
      [
        `1${g} starting -1${g}^`,
        `1${g} starting -1${g}^`,
        false,
        protoGrain,
        undefined,
      ], // TODO: this should be complete
      [
        `1${g} ending -1${g}$`,
        `1${g} ending -1${g}$`,
        true,
        protoGrain,
        undefined,
      ],
      [`-1${g}#`, previous, true, protoGrain, undefined],
      [`1${g}!`, previous, true, protoGrain, undefined],

      [`${g}$ to +1${g}$`, next, false, protoGrain, undefined],
      [`+1${g}^ to +1${g}$`, next, false, protoGrain, undefined],
      [`+1${g}#`, next, false, protoGrain, undefined],
    ];
  }).flat();
}

function getMultiPeriodTestCases(n: number): TestCase[] {
  return GRAINS.map((g) => {
    const protoGrain = GrainAliasToV1TimeGrain[g];
    const last = `Last ${n} ${GRAIN_TO_LUXON[g]}s`;
    const next = `Next ${n} ${GRAIN_TO_LUXON[g]}s`;
    return <TestCase[]>[
      [`-7${g}$ to ${g}$`, last, false, protoGrain, undefined],
      [`-7${g}^ to ${g}^`, last, true, protoGrain, undefined],
      [`7${g}`, last, false, protoGrain, undefined],
      [`7${g}!`, last, true, protoGrain, undefined],

      [`${g}^ to +7${g}^`, next, false, protoGrain, undefined],
      [`${g}$ to +7${g}$`, next, false, protoGrain, undefined],
    ];
  }).flat();
}

describe("rill time", () => {
  describe("positive cases", () => {
    const Cases: TestCase[] = [
      ...getSinglePeriodTestCases(),
      ...getMultiPeriodTestCases(7),
      [
        "-5W4M3Q2Y to -4W3M2Q1Y",
        "-5W4M3Q2Y to -4W3M2Q1Y",
        true,
        V1TimeGrain.TIME_GRAIN_WEEK,
        undefined,
      ],
      [
        "-5W-4M-3Q-2Y to -4W-3M-2Q-1Y",
        "-5W-4M-3Q-2Y to -4W-3M-2Q-1Y",
        true,
        V1TimeGrain.TIME_GRAIN_WEEK,
        undefined,
      ],

      // Ordinal point in time variations
      [
        "M^ to W3^ of M",
        "M^ to W3^ of M",
        true,
        V1TimeGrain.TIME_GRAIN_WEEK,
        undefined,
      ],
      [
        "W1^ of M to M$",
        "W1^ of M to M$",
        false,
        V1TimeGrain.TIME_GRAIN_WEEK,
        undefined,
      ],
      [
        "W1^ of M to W3^ of M",
        "W1^ of M to W3^ of M",
        true,
        V1TimeGrain.TIME_GRAIN_WEEK,
        undefined,
      ],
      [
        "ref/Y to ref/Y+1Y",
        "ref/Y to ref/Y+1Y",
        false,
        V1TimeGrain.TIME_GRAIN_WEEK,
        undefined,
      ],
    ];

    const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
    for (const [rillTime, label, complete, rangeGrain, byGrain] of Cases) {
      it(rillTime, () => {
        const parser = new nearley.Parser(compiledGrammar);
        parser.feed(rillTime);
        // assert that there is only match. this ensures unambiguous grammar.
        expect(parser.results).length(1);

        const rt = parseRillTime(rillTime);
        console.log(rt);
        expect(rt.getLabel()).toEqual(label);
        expect(rt.isComplete).toEqual(complete);
        expect(rt.rangeGrain).toEqual(rangeGrain);
        expect(rt.byGrain).toEqual(byGrain);
      });
    }
  });
});
