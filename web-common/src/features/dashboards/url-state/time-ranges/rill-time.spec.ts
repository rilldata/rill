import {
  overrideRillTimeRef,
  parseRillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import {
  capitalizeFirstChar,
  type RillTimeAsOfLabel,
  RillTimeLabel,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";
import { GrainAliasToV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { DateTimeUnit } from "luxon";
import nearley from "nearley";
import { describe, expect, it } from "vitest";
import grammar from "./rill-time.cjs";

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

    return <TestCase[]>[
      [`ref/${g} to ref/${g}+1${g}`, current, false, protoGrain, undefined],
      [
        `1${g} as of watermark/${g}+1${g}`,
        current,
        false,
        protoGrain,
        undefined,
      ],
      [
        `1${g} as of +1${g} as of watermark/${g}`,
        current,
        false,
        protoGrain,
        undefined,
      ],

      [`ref/${g}-1${g} to ref/${g}`, previous, true, protoGrain, undefined],
      [`1${g} as of watermark/${g}`, previous, true, protoGrain, undefined],
      [
        `-1${g} to ref as of watermark/${g}`,
        previous,
        true,
        protoGrain,
        undefined,
      ],
    ];
  }).flat();
}

function getMultiPeriodTestCases(n: number): TestCase[] {
  return GRAINS.map((g) => {
    const protoGrain = GrainAliasToV1TimeGrain[g];
    const last = `Last ${n} ${GRAIN_TO_LUXON[g]}s`;
    return <TestCase[]>[
      [
        `-7${g} to ref as of watermark/${g}+1${g}`,
        last,
        false,
        protoGrain,
        undefined,
      ],
      [
        `-7${g} to ref as of +1${g} as of watermark/${g}`,
        last,
        false,
        protoGrain,
        undefined,
      ],
      [`-7${g} to ref as of watermark/${g}`, last, true, protoGrain, undefined],
      [`7${g}`, last, false, protoGrain, undefined],
    ];
  }).flat();
}

function getPeriodToDateTestCases(): TestCase[] {
  return GRAINS.map((g) => {
    const protoGrain = GrainAliasToV1TimeGrain[g];
    const label = capitalizeFirstChar(`${GRAIN_TO_LUXON[g]} to date`);
    return <TestCase[]>[
      [`${g}TD as of watermark/${g}`, label, true, protoGrain, undefined],
      [
        `${g}TD as of watermark/${g}+1${g}`,
        label,
        false,
        protoGrain,
        undefined,
      ],

      [`${g}TD as of watermark/h`, label, true, protoGrain, undefined],
    ];
  }).flat();
}

function getLegacyISOTestCases(): TestCase[] {
  return [
    [`P2M3D`, "Last 2 months and 3 days", false, undefined, undefined],
    [`PT2H3M`, "Last 2 hours and 3 minutes", false, undefined, undefined],
    [
      `P2M3DT2H3M`,
      "Last 2 months, 3 days, 2 hours and 3 minutes",
      false,
      undefined,
      undefined,
    ],
  ];
}

function getLegacyDAXTestCases(): TestCase[] {
  return [
    ["rill-TD", "Today", false, undefined, undefined],
    ["rill-WTD", "Week to Date", false, undefined, undefined],
    ["rill-QTD", "Quarter to Date", false, undefined, undefined],
    ["rill-MTD", "Month to Date", false, undefined, undefined],
    ["rill-YTD", "Year to Date", false, undefined, undefined],

    ["rill-PDC", "Yesterday", false, undefined, undefined],
    ["rill-PWC", "Previous week", false, undefined, undefined],
    ["rill-PQC", "Previous quarter", false, undefined, undefined],
    ["rill-PMC", "Previous month", false, undefined, undefined],
    ["rill-PYC", "Previous year", false, undefined, undefined],
  ];
}

describe("rill time", () => {
  describe("positive cases", () => {
    const Cases: TestCase[] = [
      ...getSinglePeriodTestCases(),
      ...getMultiPeriodTestCases(7),
      ...getPeriodToDateTestCases(),
      ...getLegacyISOTestCases(),
      ...getLegacyDAXTestCases(),
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

      [
        `7D as of watermark/h+1h`,
        `Last 7 days`,
        false,
        V1TimeGrain.TIME_GRAIN_DAY,
        undefined,
      ],

      [
        `2025-02-20T01:23:45Z,2025-07-15T02:34:50Z`,
        `Custom`,
        false,
        undefined,
        undefined,
      ],

      ["inf", "All time", false, undefined, undefined],
    ];

    const compiledGrammar = nearley.Grammar.fromCompiled(grammar);
    for (const [rillTime, label, complete, rangeGrain, byGrain] of Cases) {
      it(rillTime, () => {
        const parser = new nearley.Parser(compiledGrammar);
        parser.feed(rillTime);
        // assert that there is only match. this ensures unambiguous grammar.
        expect(parser.results).length(1);

        const rt = parseRillTime(rillTime);
        expect(rt).not.toBeUndefined();
        expect(rt.getLabel()).toEqual(label);
        expect(rt.isComplete).toEqual(complete);
        expect(rt.rangeGrain).toEqual(rangeGrain);
        expect(rt.byGrain).toEqual(byGrain);

        const serialisedRillTime = rt.toString();
        const newRt = parseRillTime(serialisedRillTime);
        expect(newRt.toString()).toEqual(serialisedRillTime);
      });
    }
  });

  describe("override ref", () => {
    const Cases: [
      rillTime: string,
      refOverride: string,
      updatedRillTime: string,
    ][] = [
      ["7D AS OF watermark/Y", "watermark/Y+1Y", "7D AS OF watermark/Y+1Y"],
      ["7D AS OF watermark/Y+1Y", "watermark/Y", "7D AS OF watermark/Y"],
      ["7D AS OF watermark/Y", "now/Y", "7D AS OF now/Y"],
    ];

    for (const [rillTime, refOverride, updatedRillTime] of Cases) {
      it(`${rillTime} <> ${refOverride}`, () => {
        const rt = parseRillTime(rillTime);
        overrideRillTimeRef(rt, refOverride);
        expect(rt.toString(), updatedRillTime);
      });
    }
  });

  describe("as of label", () => {
    const Cases: [
      rillTime: string,
      asOfLabel: RillTimeAsOfLabel | undefined,
    ][] = [
      ["7D", undefined],
      ["7D as of -2D", undefined],
      [
        "7D as of -2D as of watermark",
        { label: RillTimeLabel.Watermark, snap: undefined, offset: 0 },
      ],
      [
        "7D as of -2D as of watermark/D",
        { label: RillTimeLabel.Watermark, snap: "D", offset: 0 },
      ],
      [
        "7D as of -2D as of watermark/D+1D",
        { label: RillTimeLabel.Watermark, snap: "D", offset: 1 },
      ],
      [
        "7D as of -2D as of watermark/D+1h",
        { label: RillTimeLabel.Watermark, snap: "D", offset: 0 },
      ],
      [
        "7D as of -2D as of watermark/h+1h",
        { label: RillTimeLabel.Watermark, snap: "h", offset: 1 },
      ],
    ];

    for (const [rillTime, asOfLabel] of Cases) {
      it(rillTime, () => {
        const rt = parseRillTime(rillTime);
        expect(rt.asOfLabel).toEqual(asOfLabel);
      });
    }
  });
});
