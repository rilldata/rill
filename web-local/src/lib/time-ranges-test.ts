import {
  queryServiceMetricsViewTimeRanges,
  queryServiceMetricsViewTimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import { Interval, DateTime, type DateTimeUnit } from "luxon";
import { GrainAliasToOrder } from "@rilldata/web-common/lib/time/new-grains";

const GRAINS = ["Y", "Q", "M", "W", "D", "H", "m", "s"] as const;

const REF = "ref";

type RillGrain = (typeof GRAINS)[number];

type TimeMetadata = {
  watermark: DateTime;
  now: DateTime;
  latest: DateTime;
};

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

interface Test {
  syntax: string;
  description: string;
  interval: Interval;
  log?: true;
}

function generateToTests(
  metadata: TimeMetadata,
  n1?: number,
  n2?: number,
): Test[] {
  if (!n1 || !n2) {
    n1 = Math.floor(Math.random() * 100);
    n2 = Math.floor(n1 * Math.random());
  }

  const tests: Test[] = [];

  GRAINS.forEach((grain, i) => {
    GRAINS.forEach((grain2, j) => {
      if (j < i) {
        return;
      }

      const unit1 = GRAIN_TO_LUXON[grain];
      const unit2 = GRAIN_TO_LUXON[grain2];

      tests.push({
        syntax: `-${n1}${grain}/${grain} to -${n2}${grain2}/${grain2}`,
        description: `${n1} ${unit1}(s) ago to ${n2} ${unit2}(s) ago, snapped to grain start`,

        interval: Interval.fromDateTimes(
          metadata.now.minus({ [unit1]: n1 }).startOf(unit1),
          metadata.now.minus({ [unit2]: n2 }).startOf(unit2),
        ),
      });
    });
  });

  return tests;
}

function lastNCompletePeriods(metadata: TimeMetadata, upTo: number = 100) {
  const tests: Test[] = [];

  for (let i = 1; i <= upTo; i++) {
    GRAINS.forEach((grain) => {
      const unit = GRAIN_TO_LUXON[grain];

      const interval = Interval.fromDateTimes(
        metadata.watermark.minus({ [unit]: i }).startOf(unit),
        metadata.watermark.startOf(unit),
      );

      tests.push({
        syntax: `-${i}${grain} to ${REF} as of watermark/${grain}`,
        description: `Last ${i} complete ${unit}(s) using ref syntax`,
        interval,
      });

      tests.push({
        syntax: `-${i}${grain}/${grain} to ${REF}/${grain} as of watermark`,
        description: `Last ${i} complete ${unit}(s) using ref syntax`,
        interval,
      });
    });
  }

  return tests;
}

export function generateStartingAndEndingTests(metadata: TimeMetadata) {
  return GRAINS.map((periodGrain) => {
    const periodUnit = GRAIN_TO_LUXON[periodGrain];
    const periodInteger = Math.floor(Math.random() * 100);
    const offsetInteger = Math.floor(Math.random() * 100);

    return GRAINS.map((offsetGrain) => {
      const offsetUnit = GRAIN_TO_LUXON[offsetGrain];

      return [
        <Test>{
          syntax: `${periodInteger}${periodGrain} as of latest-${offsetInteger}${offsetGrain}`,
          description: `The ${periodInteger} ${periodUnit} period starting exactly -${offsetInteger} ${offsetUnit}(s) ago`,
          interval: Interval.fromDateTimes(
            metadata.latest

              .minus({ [offsetUnit]: offsetInteger })
              .minus({ [periodUnit]: periodInteger }),
            metadata.latest.minus({ [offsetUnit]: offsetInteger }),
          ),
        },
      ];
    });
  })
    .flat()
    .flat();
}

function lastNPeriods(metadata: TimeMetadata, upTo: number = 100) {
  const tests: Test[] = [];

  for (let i = 1; i <= upTo; i++) {
    GRAINS.forEach((grain) => {
      const unit = GRAIN_TO_LUXON[grain];

      const interval = Interval.fromDateTimes(
        metadata.latest.minus({ [unit]: i - 1 }).startOf(unit),
        metadata.latest.plus({ [unit]: 1 }).startOf(unit),
      );

      tests.push({
        syntax: `-${i}${grain} to ${REF} as of latest/${grain}+1${grain}`,
        description: `Last ${i} complete ${unit}(s) using ref syntax`,
        interval,
      });

      tests.push({
        syntax: `-${i - 1}${grain}/${grain} to ${REF}/${grain}+1${grain} as of latest`,
        description: `Last ${i} complete ${unit}(s) using ref syntax`,
        interval,
      });
    });
  }

  return tests;
}

function lastNPeriodToNow(metadata: TimeMetadata, upTo: number = 100) {
  const tests: Test[] = [];

  for (let i = 1; i <= upTo; i++) {
    GRAINS.forEach((grain) => {
      const unit = GRAIN_TO_LUXON[grain];

      // const truncatedNow = metadata.now.startOf("second");

      const interval = Interval.fromDateTimes(
        metadata.now.minus({ [unit]: i }),
        metadata.now,
      );

      tests.push({
        syntax: `-${i}${grain} to ref as of ${metadata.now.toISO()}`,
        description: `Last ${i} ${unit}(s) to now `,
        interval,
      });

      tests.push({
        syntax: `-${i}${grain} to ref as of latest`,
        description: `Last ${i} ${unit}(s) to latest`,
        interval: Interval.fromDateTimes(
          metadata.latest.minus({ [unit]: i }),
          metadata.latest,
        ),
      });

      tests.push({
        syntax: `-${i}${grain} to ref as of latest`,
        description: `Last ${i} ${unit}(s) to watermark`,
        interval: Interval.fromDateTimes(
          metadata.watermark.minus({ [unit]: i }),
          metadata.watermark,
        ),
      });
    });
  }

  return tests;
}

function getHigherOrderGrain(grain: RillGrain): RillGrain | undefined {
  const index = GRAINS.indexOf(grain);

  const higherOrderGrain = GRAINS[index - 1];

  if (higherOrderGrain === "Q") {
    return "Y";
  }

  return higherOrderGrain;
}

function generateCurrentAndPreviousPeriods(metadata: TimeMetadata) {
  const tests: Test[] = [];

  GRAINS.forEach((grain) => {
    const unit = GRAIN_TO_LUXON[grain];

    const current = Interval.fromDateTimes(
      metadata.latest.startOf(unit),
      metadata.latest.plus({ [unit]: 1 }).startOf(unit),
    );

    const previous = Interval.fromDateTimes(
      metadata.watermark.startOf(unit).minus({ [unit]: 1 }),
      metadata.watermark.startOf(unit),
    );

    GRAINS.forEach((grain2) => {
      const unit2 = GRAIN_TO_LUXON[grain2];
      if (GrainAliasToOrder[grain2] >= GrainAliasToOrder[grain]) return;

      tests.push({
        syntax: `${grain}TD as of latest/${grain2}+1${grain2}`,
        description: `The current ${unit} up the latest ${unit2}`,
        interval: Interval.fromDateTimes(
          metadata.latest.startOf(unit),
          metadata.latest.plus({ [unit2]: 1 }).startOf(unit2),
        ),
      });

      tests.push({
        syntax: `ref/${grain} to ref as of latest/${grain2}+1${grain2}`,
        description: `The current ${unit} up the latest ${unit2}`,
        interval: Interval.fromDateTimes(
          metadata.latest.startOf(unit),
          metadata.latest.plus({ [unit2]: 1 }).startOf(unit2),
        ),
      });
    });

    tests.push({
      syntax: `ref/${grain} to ref/${grain}+1${grain} as of latest`,
      description: `The full current ${unit}`,
      interval: current,
    });

    tests.push({
      syntax: `1${grain} as of latest/${grain}+1${grain}`,
      description: `The full current ${unit}`,
      interval: current,
    });

    tests.push({
      syntax: `-0${grain}/${grain} to +1${grain}/${grain} as of latest`,
      description: `The full current ${unit}`,
      interval: current,
    });

    tests.push({
      syntax: `-1${grain}/${grain} to ref/${grain} as of latest`,
      description: `The previous ${unit} from latest`,
      interval: Interval.fromDateTimes(
        metadata.latest.minus({ [unit]: 1 }).startOf(unit),
        metadata.latest.startOf(unit),
      ),
    });

    //Complete
    tests.push({
      syntax: `1${grain} as of watermark/${grain}`,
      description: `The last complete ${unit}`,
      interval: previous,
    });

    tests.push({
      syntax: `-1${grain}/${grain} to ref/${grain} as of watermark`,
      description: `The last complete ${unit}`,
      interval: previous,
    });
  });

  return tests;
}

function generateFirstNPeriodTests(metadata: TimeMetadata, n: number = 3) {
  const tests: Test[] = [];

  GRAINS.forEach((grain) => {
    const unit = GRAIN_TO_LUXON[grain];

    const higherOrderGrain = getHigherOrderGrain(grain);

    if (!higherOrderGrain) return;
    const higherOrderUnit = GRAIN_TO_LUXON[higherOrderGrain];

    const ref = metadata.now.startOf(higherOrderUnit).startOf("second");
    let start = ref;

    if (unit === "week") {
      if (start.weekday >= 5) {
        start = start.startOf("week").plus({ week: 1 });
      } else {
        start = start.startOf("week");
      }

      tests.push({
        syntax: `${REF}/${higherOrderGrain}/W to ${REF}/${higherOrderGrain}/W+${n}${grain} as of ${ref.toISO({ suppressMilliseconds: true })}`,
        description: `First ${n} ${unit}(s) of the current ${higherOrderUnit}`,
        interval: Interval.fromDateTimes(start, start.plus({ [unit]: n })),
      });
    } else {
      tests.push({
        syntax: `${REF}/${higherOrderGrain} to ${REF}/${higherOrderGrain}+${n}${grain} as of ${ref.toISO({ suppressMilliseconds: true })}`,
        description: `First ${n} ${unit}(s) of the current ${higherOrderUnit}`,
        interval: Interval.fromDateTimes(start, start.plus({ [unit]: n })),
      });
    }
  });

  return tests;
}

export async function runTests(metricsViewName: string) {
  const instanceId = get(runtime).instanceId;

  let failures = 1;

  const okay = await queryServiceMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    {},
  );

  const { timeRangeSummary } = okay;
  if (
    !timeRangeSummary ||
    !timeRangeSummary.min ||
    !timeRangeSummary.watermark
  ) {
    console.log("No time range summary");
    return;
  }

  const allTime = Interval.fromDateTimes(
    DateTime.fromISO(timeRangeSummary.min).setZone("UTC"),
    DateTime.fromISO(timeRangeSummary.watermark).setZone("UTC"),
  );

  if (!allTime.isValid) {
    console.log("Invalid time range summary");
    return;
  }

  console.log(allTime.toLocaleString(DateTime.DATETIME_FULL_WITH_SECONDS));

  const latest = allTime.end;
  const watermark = DateTime.fromISO(timeRangeSummary.watermark).setZone("UTC");
  const now = DateTime.now().setZone("UTC");

  const metadata = {
    latest,
    watermark,
    now,
  };

  console.log({
    latest: latest.toISO(),
    watermark: watermark.toISO(),
    now: now.toISO(),
  });

  const testCases = generateTestCases(metadata);

  const expressions = testCases.map((testCase) => testCase.syntax);

  const response = await queryServiceMetricsViewTimeRanges(
    instanceId,
    metricsViewName,
    { expressions },
  );

  testCases.forEach(({ interval, log }, i) => {
    const apiFormat = {
      start: interval.start
        ?.setZone("UTC")
        .toISO({ suppressMilliseconds: true }),
      end: interval.end?.toISO({ suppressMilliseconds: true }),
    };

    if (response?.resolvedTimeRanges?.[i]?.start !== apiFormat.start) {
      console.log(testCases[i].description);
      console.log(
        `${failures.toLocaleString("en-US", {
          minimumIntegerDigits: 2,
          useGrouping: false,
        })}. FAIL (start) // ${testCases[i].syntax.padEnd(36)} // Expected: ${String(apiFormat.start).padEnd(20)} // Got: ${response?.timeRanges?.[i]?.start}`,
      );
      failures++;
    } else if (response?.resolvedTimeRanges?.[i]?.end !== apiFormat.end) {
      console.log(testCases[i].description);
      console.log(
        `${failures.toLocaleString("en-US", {
          minimumIntegerDigits: 2,
          useGrouping: false,
        })}. FAIL (end)   // ${testCases[i].syntax.padEnd(36)} // Expected: ${apiFormat.end?.padEnd(20)} // Got: ${response?.timeRanges?.[i]?.end}`,
      );
      failures++;
    } else if (log) {
      console.log(
        `--. PASS (start) // ${testCases[i].syntax.padEnd(36)} // Expected: ${apiFormat.start} // Got: ${response?.timeRanges?.[i]?.start} `,
      );
      console.log(
        `--. PASS (end)   // ${testCases[i].syntax.padEnd(36)} // Expected: ${apiFormat.end} // Got: ${response?.timeRanges?.[i]?.end}`,
      );
    }
  });
}

export function generateISOTests(randomCount: number = 50): Test[] {
  const hardcoded = DateTime.fromISO("2025-05-03T13:22:45Z", { zone: "UTC" });

  const randomRefs: DateTime[] = Array.from({ length: randomCount }, () => {
    const year = Math.floor(Math.random() * 31) + 2000;
    const month = Math.floor(Math.random() * 12) + 1;
    const day =
      Math.floor(
        Math.random() * (DateTime.utc(year, month)?.daysInMonth ?? 28),
      ) + 1;
    const hour = Math.floor(Math.random() * 24);
    const minute = Math.floor(Math.random() * 60);
    const second = Math.floor(Math.random() * 60);

    return DateTime.fromObject(
      { year, month, day, hour, minute, second },
      { zone: "UTC" },
    );
  });

  return [hardcoded, ...randomRefs].flatMap((ref) =>
    generateIntervalTestCases(ref),
  );
}

const specs = [
  { unit: "year", format: "yyyy", desc: "Year" },
  { unit: "month", format: "yyyy-MM", desc: "Month" },
  { unit: "day", format: "yyyy-MM-dd", desc: "Day" },
  { unit: "hour", format: "yyyy-MM-dd'T'HH", desc: "Hour" },
  { unit: "minute", format: "yyyy-MM-dd'T'HH:mm", desc: "Minute" },
  { unit: "second", format: "yyyy-MM-dd'T'HH:mm:ss", desc: "Second" },
] as const;

export function generateIntervalTestCases(reference: DateTime): Test[] {
  const tests: Test[] = [];

  for (const { unit, format, desc } of specs) {
    const start = reference.startOf(unit);

    const end = start.plus({ [unit]: 1 });

    const syntax = start.toFormat(format) + (unit === "second" ? "Z" : "");

    tests.push({
      syntax,
      description: `${desc}‐level interval (${syntax})`,
      interval: Interval.fromDateTimes(start, end),
    });
  }

  for (const { unit, desc } of specs) {
    const start = reference.startOf(unit);
    const end = start.plus({ [unit]: 1 });

    const syntax = `${start.toISO({ suppressMilliseconds: true })} to ${end.toISO({ suppressMilliseconds: true })}`;

    tests.push({
      syntax,
      description: `Range from start of ${desc.toLowerCase()} to next`,
      interval: Interval.fromDateTimes(start, end),
    });
  }

  return tests;
}

export const forceTestImport = "test";

function generateTestCases(metadata: TimeMetadata): Test[] {
  return [
    ...generateToTests(metadata, 6, 3),
    ...generateToTests(metadata, 4, 1),
    ...generateToTests(metadata, 9, 7),
    ...generateToTests(metadata),
    ...lastNCompletePeriods(metadata),
    ...lastNPeriods(metadata),
    ...lastNPeriodToNow(metadata),
    ...generateFirstNPeriodTests(metadata),
    ...generateCurrentAndPreviousPeriods(metadata),
    ...generateStartingAndEndingTests(metadata),
    ...generateISOTests(),

    {
      syntax: `-6d/d to ${REF}/h as of latest`,
      description: "Six days to date excluding the current hour",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({ day: 6 }).startOf("day"),
        metadata.latest.startOf("hour"),
      ),
    },
    {
      syntax: `-2h/h-3d to -2h/h as of latest`,
      description: "The 3 day period starting two hour boundaries ago",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({ hour: 2 }).startOf("hour").minus({ day: 3 }),
        metadata.latest.minus({ hour: 2 }).startOf("hour"),
      ),
    },
    {
      syntax: `-2s-3m-4h-5d-6W-7M-8Q-9Y to -1s-2m-3h-4d-5W-6M-7Q-8Y as of latest`,
      description: "Stacking offsets",
      interval: Interval.fromDateTimes(
        metadata.latest
          .minus({
            second: 2,
          })
          .minus({
            minute: 3,
          })
          .minus({
            hour: 4,
          })
          .minus({
            day: 5,
          })
          .minus({
            week: 6,
          })
          .minus({
            month: 7,
          })
          .minus({
            quarter: 8,
          })
          .minus({
            year: 9,
          }),

        metadata.latest
          .minus({
            second: 1,
          })
          .minus({
            minute: 2,
          })
          .minus({
            hour: 3,
          })
          .minus({
            day: 4,
          })
          .minus({
            week: 5,
          })
          .minus({
            month: 6,
          })
          .minus({
            quarter: 7,
          })
          .minus({
            year: 8,
          }),
      ),
    },
    {
      syntax: `-2s3m4h5d6w7M8Q9Y to -1s2m3h4d5w6M7Q8Y as of latest`,
      description: "Offsetting by compound duration",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({
          second: 2,
          minute: 3,
          hour: 4,
          day: 5,
          week: 6,
          month: 7,
          quarter: 8,
          year: 9,
        }),
        metadata.latest.minus({
          second: 1,
          minute: 2,
          hour: 3,
          day: 4,
          week: 5,
          month: 6,
          quarter: 7,
          year: 8,
        }),
      ),
    },

    {
      syntax: `-42W-3M/W to -2M-3h/M+1M as of latest`,
      description:
        "The start of the week 42 weeks and 3 minutes ago to the end of the month 2 months and 3 hours ago",
      interval: Interval.fromDateTimes(
        metadata.latest
          .minus({
            week: 42,
            month: 3,
          })
          .startOf("week"),

        metadata.latest
          .minus({
            month: 2,
            hour: 3,
          })
          .plus({
            month: 1,
          })
          .startOf("month"),
      ),
    },
    {
      syntax: `-3Y/d+6h-6W/h to -4d/M+1M-23h-7m as of latest`,
      description:
        "The day start three days ago plus 6 hours minus 6 weeks rounded to the hour start TO the month end three years ago minus 23 hours and 7 minutes",
      interval: Interval.fromDateTimes(
        metadata.latest
          .minus({
            year: 3,
          })
          .startOf("day")
          .plus({
            hour: 6,
          })
          .minus({
            week: 6,
          })
          .startOf("hour"),

        metadata.latest
          .minus({
            day: 4,
          })
          .plus({ month: 1 })
          .startOf("month")
          .minus({
            hour: 23,
          })
          .minus({
            minute: 7,
          }),
      ),
    },

    {
      syntax: `3W18D23h as of latest-3Y`,
      description:
        "The three week, 18 day, and 23 hour period starting exactly three years ago",
      interval: Interval.fromDateTimes(
        metadata.latest
          .minus({
            year: 3,
          })
          .minus({
            week: 3,
            day: 18,
            hour: 23,
          }),
        metadata.latest.minus({
          year: 3,
        }),
      ),
    },

    {
      syntax: `-2h to ref as of latest`,
      description: "The two hour period ending at the latest data point",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({
          hour: 2,
        }),
        metadata.latest,
      ),
    },

    {
      syntax: `3W as of latest-3Y`,
      description: "The three week period ending exactly three years ago",
      interval: Interval.fromDateTimes(
        metadata.latest
          .minus({
            year: 3,
          })
          .minus({
            week: 3,
          }),
        metadata.latest.minus({
          year: 3,
        }),
      ),
    },
    {
      syntax: `7h as of watermark/d`,
      description: "The last 7 hours of the last complete day",
      interval: Interval.fromDateTimes(
        metadata.watermark.startOf("day").minus({ hour: 7 }),
        metadata.watermark.startOf("day"),
      ),
    },

    {
      syntax: `H4 as of latest-1d`,
      description: "The fourth hour of the previous day",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({ day: 1 }).set({ hour: 3 }).startOf("hour"),
        metadata.latest.minus({ day: 1 }).set({ hour: 4 }).startOf("hour"),
      ),
    },

    {
      syntax: `-2d/d to -2d/d+3h as of latest`,
      description: "The first 3 hours of two days ago",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({ day: 2 }).startOf("day"),
        metadata.latest.minus({ day: 2 }).set({ hour: 3 }).startOf("hour"),
      ),
    },
    {
      syntax: `-2d/d to ref/d+1d as of latest-1Q`,
      description:
        "The 3 day period that includes the current incomplete day and the 2 previous complete days offset into the previous quarter",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({ quarter: 1 }).minus({ day: 2 }).startOf("day"),
        metadata.latest.minus({ quarter: 1 }).startOf("day").plus({ day: 1 }),
      ),
    },
    {
      syntax: `3d as of latest-1Q/d+1d`,
      description:
        "The 3 day period that includes the current incomplete day and the 2 previous complete days offset into the previous quarter",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({ quarter: 1 }).startOf("day").minus({ day: 2 }),
        metadata.latest.minus({ quarter: 1 }).startOf("day").plus({ day: 1 }),
      ),
    },
    {
      syntax: `m30 of H12 of D5 of Q3 as of latest-2Y`,
      description:
        "The 30th minute of the 12th hour of the fifth day of quarter three of two years ago",
      interval: Interval.fromDateTimes(
        metadata.latest
          .minus({ year: 2 })
          .startOf("year")
          .plus({ quarter: 2 })
          .plus({ day: 4 })
          .set({ hour: 11 })
          .set({ minute: 29 })
          .startOf("minute"),
        metadata.latest
          .minus({ year: 2 })
          .startOf("year")
          .plus({ quarter: 2 })
          .plus({ day: 4 })
          .set({ hour: 11 })
          .set({ minute: 30 })
          .startOf("minute"),
      ),
    },
    {
      syntax: `-127d/W to -89d/W as of watermark`,
      description:
        "The week start 127 days ago to the current week start 89 days ago",
      interval: Interval.fromDateTimes(
        metadata.watermark.minus({ day: 127 }).startOf("week"),
        metadata.watermark.minus({ day: 89 }).startOf("week"),
      ),
    },
    {
      syntax: `-15Y/W to -4Y/W as of latest`,
      description:
        "The current week start 15 years ago to the current week start 4 years ago",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({ year: 15 }).startOf("week"),
        metadata.latest.minus({ year: 4 }).startOf("week"),
      ),
    },
    {
      syntax: `-2Y/Y/W to -2Y/W as of latest`,
      description: "All weeks of the current year, offset into two years ago",
      interval: Interval.fromDateTimes(
        DateTime.fromObject(
          {
            weekYear: metadata.latest.year - 2,
            weekNumber: 1,
          },
          { zone: metadata.latest.zone },
        ),
        metadata.latest.minus({ year: 2 }).startOf("week"),
      ),
    },
    {
      syntax: `-2Y/Y/W to -1Y/Y/W`,
      description: "The year two years ago by ISO week rules",
      interval: Interval.fromDateTimes(
        DateTime.fromObject(
          {
            weekYear: metadata.latest.year - 2,
            weekNumber: 1,
          },
          { zone: metadata.latest.zone },
        ),
        DateTime.fromObject(
          {
            weekYear: metadata.latest.year - 1,
            weekNumber: 1,
          },
          { zone: metadata.latest.zone },
        ),
      ),
    },
    {
      syntax: `94m as of latest/m-9d`,
      description:
        "The 94 minute period ending 9 days ago, rounded to the minute",
      interval: Interval.fromDateTimes(
        metadata.latest
          .minus({ day: 9 })
          .startOf("minute")
          .minus({ minute: 94 }),
        metadata.latest.minus({ day: 9 }).startOf("minute"),
      ),
    },
    {
      syntax: `94m as of latest-9d`,
      description: "The 94 minute period ending exactly 9 days ago",
      interval: Interval.fromDateTimes(
        metadata.latest.minus({ day: 9 }).minus({ minute: 94 }),
        metadata.latest.minus({ day: 9 }),
      ),
    },
    {
      syntax: `-1y/W to -1y/W+1W as of 2025-05-14T13:43:00Z`,
      description: "The one week period starting two years ago from 05/14/2025",
      interval: Interval.fromDateTimes(
        DateTime.fromISO("2025-05-14T13:43:00Z", { zone: "UTC" })
          .minus({ year: 1 })
          .startOf("week"),
        DateTime.fromISO("2025-05-14T13:43:00Z", { zone: "UTC" })
          .minus({ year: 1 })
          .plus({ week: 1 })
          .startOf("week"),
      ),
    },
    {
      syntax: `1Q as of 2025-02-25T09:00:00Z-8W `,
      description: "The 1 quarter period ending exactly 8 weeks ago",
      interval: Interval.fromDateTimes(
        DateTime.fromISO("2025-02-25T09:00:00Z", { zone: "UTC" })
          .minus({ week: 8 })
          .minus({ quarter: 1 }),
        DateTime.fromISO("2025-02-25T09:00:00Z", { zone: "UTC" }).minus({
          week: 8,
        }),
      ),
    },
    {
      syntax: `1Q as of 2025-02-25T09:00:00Z-8W/Q`,
      description:
        "The 1 quarter period ending exactly 8 weeks ago, snapped to quarter",
      interval: Interval.fromDateTimes(
        DateTime.fromISO("2025-02-25T09:00:00Z", { zone: "UTC" })
          .minus({ week: 8 })
          .startOf("quarter")
          .minus({ quarter: 1 }),
        DateTime.fromISO("2025-02-25T09:00:00Z", { zone: "UTC" })
          .minus({ week: 8 })
          .startOf("quarter"),
      ),
    },
    {
      syntax: `-1W to ref as of watermark/W-1Y`,
      description:
        "The one week period starting one year ago, rounded to the week",
      interval: Interval.fromDateTimes(
        metadata.watermark
          .startOf("week")
          .minus({ year: 1 })
          .minus({ week: 1 }),
        metadata.watermark.startOf("week").minus({ year: 1 }),
      ),
    },
    {
      syntax: `6M as of latest/M+1M tz America/New_York`,
      description: "Last 6 months snapped",
      interval: Interval.fromDateTimes(
        metadata.latest
          .setZone("America/New_York")
          .startOf("month")
          .minus({ month: 5 }),
        metadata.latest
          .setZone("America/New_York")
          .startOf("month")
          .plus({ month: 1 }),
      ),
    },

    /*
    The following tests are commented out because they may not fully supported.
    They can be uncommented and adjusted as needed when the functionality is implemented.
    */
    //  {
    //   syntax: `W1 of 2024`,
    //   description:
    //     "The one week period starting one year ago, rounded to the week",
    //   interval: Interval.fromDateTimes(
    //     metadata.latest.startOf("week").minus({ year: 1 }).minus({ week: 1 }),
    //     metadata.latest.startOf("week").minus({ year: 1 }),
    //   ),
    // },
    // ...[1, 2, 4, 13, 26, 39, 52, 53, 54].map((value) => {
    //   return {
    //     syntax: `W${value} of Y as of latest`,
    //     description: `Week ${value} of the current year`,
    //     interval: Interval.fromDateTimes(
    //       DateTime.fromObject(
    //         {
    //           weekYear: metadata.latest.year,
    //           weekNumber: value,
    //         },
    //         { zone: metadata.latest.zone },
    //       ),
    //       DateTime.fromObject(
    //         {
    //           weekYear: metadata.latest.year,
    //           weekNumber: value,
    //         },
    //         { zone: metadata.latest.zone },
    //       )
    //         .plus({ week: 1 })
    //         .startOf("week"),
    //     ),
    //   };
    // }),

    // {
    //   syntax: `W1 of -1M${INTERVAL_CHARACTER}`,
    //   description:
    //     "The first week of the last complete month following ISO rules",
    //   interval: Interval.fromDateTimes(
    //     now.minus({ month: 1 }).startOf("month").weekday >= 5
    //       ? now
    //           .minus({ month: 1 })
    //           .startOf("month")
    //           .startOf("week")
    //           .plus({ week: 1 })
    //       : now.minus({ month: 1 }).startOf("month").startOf("week"),
    //     (now.minus({ month: 1 }).startOf("month").weekday >= 5
    //       ? now
    //           .minus({ month: 1 })
    //           .startOf("month")
    //           .plus({ week: 1 })
    //           .startOf("week")
    //       : now.minus({ month: 1 }).startOf("month").startOf("week")
    //     )
    //       .plus({ week: 1 })
    //       .startOf("week"),
    //   ),
    // },
    // {
    //   syntax: `W1 of -1M${START_CHARACTER} to -1M${END_CHARACTER}`,
    //   description:
    //     "The first week of the last complete month following ISO rules",
    //   interval: Interval.fromDateTimes(
    //     now.minus({ month: 1 }).startOf("month").weekday >= 5
    //       ? now
    //           .minus({ month: 1 })
    //           .startOf("month")
    //           .startOf("week")
    //           .plus({ week: 1 })
    //       : now.minus({ month: 1 }).startOf("month").startOf("week"),
    //     (now.minus({ month: 1 }).startOf("month").weekday >= 5
    //       ? now
    //           .minus({ month: 1 })
    //           .startOf("month")
    //           .plus({ week: 1 })
    //           .startOf("week")
    //       : now.minus({ month: 1 }).startOf("month").startOf("week")
    //     )
    //       .plus({ week: 1 })
    //       .startOf("week"),
    //   ),
    // },
    // {
    //   syntax: "W1 of M",
    //   description: "The first week of the current month following ISO rules",
    //   interval: Interval.fromDateTimes(
    //     now.startOf("month").weekday >= 5
    //       ? now.startOf("month").plus({ week: 1 }).startOf("week")
    //       : now.startOf("month").startOf("week"),
    //     (now.startOf("month").weekday >= 5
    //       ? now.startOf("month").plus({ week: 1 }).startOf("week")
    //       : now.startOf("month").startOf("week")
    //     )
    //       .plus({ week: 1 })
    //       .startOf("week"),
    //   ),
    // },

    // {
    //   syntax: `>3s of m4 of H2 of D4 of -1M${INTERVAL_CHARACTER}`,
    //   description:
    //     "The last three seconds of the fourth minute of the second hour of the 4th day of last month",
    //   interval: Interval.fromDateTimes(
    //     now
    //       .minus({ month: 1 })
    //       .set({ day: 4 })
    //       .set({ hour: 1 })
    //       .set({ minute: 3 })
    //       .set({ second: 57 })
    //       .startOf("second"),
    //     now
    //       .minus({ month: 1 })
    //       .set({ day: 4 })
    //       .set({ hour: 1 })
    //       .set({ minute: 4 })
    //       .startOf("minute"),
    //   ),
    // },

    // {
    //   syntax: `<6h of D25 as of watermark-3M`,
    //   description:
    //     "The first 6 hours of the 25th day of the month three months ago",
    //   interval: Interval.fromDateTimes(
    //     metadata.watermark.minus({ month: 3 }).set({ day: 25 }).startOf("day"),
    //     metadata.watermark
    //       .minus({ month: 3 })
    //       .set({ day: 25 })
    //       .startOf("day")
    //       .plus({ hours: 6 }),
    //   ),
    // },
    // {
    //   syntax: `6h ${STARTING} D25^ as of -3M`,
    //   description:
    //     "The first 6 hours of the 25th day of the month three months ago",
    //   interval: Interval.fromDateTimes(
    //     now.minus({ month: 3 }).set({ day: 25 }).startOf("day"),
    //     now
    //       .minus({ month: 3 })
    //       .set({ day: 25 })
    //       .startOf("day")
    //       .plus({ hours: 6 }),
    //   ),
    // },

    // {
    //   syntax: "<6h of W4 of Q2",
    //   description:
    //     "The first 6 hours of week 4 of quarter 2 of the current year",
    //   interval: Interval.fromDateTimes(
    //     now.startOf("year").plus({ quarter: 1 }).weekday >= 5
    //       ? now
    //           .startOf("year")
    //           .plus({ quarter: 1 })
    //           .plus({ week: 4 })
    //           .startOf("week")
    //       : now
    //           .startOf("year")
    //           .plus({ quarter: 1 })
    //           .plus({ week: 3 })
    //           .startOf("week"),
    //     (now.startOf("year").plus({ quarter: 1 }).weekday >= 5
    //       ? now
    //           .startOf("year")
    //           .plus({ quarter: 1 })
    //           .plus({ week: 4 })
    //           .startOf("week")
    //       : now
    //           .startOf("year")
    //           .plus({ quarter: 1 })
    //           .plus({ week: 3 })
    //           .startOf("week")
    //     )
    //       .plus({ hour: 6 })
    //       .startOf("hour"),
    //   ),
    // },

    // {
    //   syntax: `Y${START_CHARACTER} to M${START_CHARACTER}`,
    //   description: "All complete months of the current year",
    //   interval: Interval.fromDateTimes(
    //     now.startOf("year"),
    //     now.startOf("month"),
    //   ),
    // },
    // {
    //   syntax: `Y${START_CHARACTER} to M${END_CHARACTER}`,
    //   description:
    //     "All months of the current year, including the current incomplete one",
    //   interval: Interval.fromDateTimes(
    //     now.startOf("year"),
    //     now.plus({ month: 1 }).startOf("month"),
    //   ),
    // },
    // {
    //   syntax: `Y${START_CHARACTER} to D${START_CHARACTER}`,
    //   description: "All complete days of the current year",
    //   interval: Interval.fromDateTimes(now.startOf("year"), now.startOf("day")),
    // },
    // {
    //   syntax: `Y${START_CHARACTER} to D${END_CHARACTER}`,
    //   description:
    //     "All days of the current year, including the current incomplete one",
    //   interval: Interval.fromDateTimes(
    //     now.startOf("year"),
    //     now.plus({ day: 1 }).startOf("day"),
    //   ),
    // },

    // {
    //   syntax: `Y/Y/W${START_CHARACTER} to W${START_CHARACTER}`,
    //   description:
    //     "All weeks of the current year, excluding the current incomplete one",
    //   interval: Interval.fromDateTimes(
    //     DateTime.fromObject(
    //       {
    //         weekYear: now.year,
    //         weekNumber: 1,
    //       },
    //       { zone: now.zone },
    //     ),
    //     now.startOf("week"),
    //   ),
    // },

    // {
    //   syntax: `D3 of W2 of -2M${INTERVAL_CHARACTER}`,
    //   description: "Day 3 of week 2 of two months ago",
    //   interval: Interval.fromDateTimes(
    //     (now.minus({ month: 2 }).startOf("month").weekday >= 5
    //       ? now
    //           .minus({ month: 2 })
    //           .startOf("month")
    //           .startOf("week")
    //           .plus({ week: 1 })
    //       : now.minus({ month: 2 }).startOf("month").startOf("week")
    //     ).plus({ day: 2, week: 1 }),

    //     (now.minus({ month: 2 }).startOf("month").weekday >= 5
    //       ? now
    //           .minus({ month: 2 })
    //           .startOf("month")
    //           .startOf("week")
    //           .plus({ week: 1 })
    //       : now.minus({ month: 2 }).startOf("month").startOf("week")
    //     ).plus({ day: 3, week: 1 }),
    //   ),
    // },

    // {
    //   syntax: `94m ${STARTING} -9d`,
    //   description: "The 94 minute period starting exactly 9 days ago",
    //   interval: Interval.fromDateTimes(
    //     now.plus({ [smallestTimeGrain]: 1 }).minus({ day: 9 }),
    //     now
    //       .plus({ [smallestTimeGrain]: 1 })
    //       .minus({ day: 9 })
    //       .plus({ minute: 94 }),
    //   ),
    // },

    // {
    //   syntax: `94Q ${STARTING} -9d`,
    //   description: "The 94 quarter period starting exactly 9 days ago",
    //   interval: Interval.fromDateTimes(
    //     now.plus({ [smallestTimeGrain]: 1 }).minus({ day: 9 }),
    //     now
    //       .plus({ [smallestTimeGrain]: 1 })
    //       .minus({ day: 9 })
    //       .plus({ quarter: 94 }),
    //   ),
    // },
  ];
}
