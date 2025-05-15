import {
  queryServiceMetricsViewTimeRanges,
  queryServiceMetricsViewTimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import { Interval, DateTime, type DateTimeUnit } from "luxon";

const GRAINS = ["Y", "Q", "M", "W", "D", "H", "m", "s"] as const;

const START_CHARACTER = "^";
const END_CHARACTER = "$";
const LITERAL_CHARACTER = "";
const INTERVAL_CHARACTER = "!";
const ENDING = "ending";
const STARTING = "starting";

type RillGrain = (typeof GRAINS)[number];

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

function generateToTests(now: DateTime, n1?: number, n2?: number): Test[] {
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
        syntax: `-${n1}${grain}${START_CHARACTER} to -${n2}${grain2}${START_CHARACTER}`,
        description: `${n1} ${unit1}(s) ago to ${n2} ${unit2}(s) ago, snapped to grain start`,

        interval: Interval.fromDateTimes(
          now.minus({ [unit1]: n1 }).startOf(unit1),
          now.minus({ [unit2]: n2 }).startOf(unit2),
        ),
      });
    });
  });

  return tests;
}

function generateEndTests(now: DateTime, n1?: number, n2?: number): Test[] {
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
        syntax: `-${n1}${grain}${END_CHARACTER} to -${n2}${grain2}${END_CHARACTER}`,
        description: `${n1} ${unit1}(s) ago to ${n2} ${unit2}(s) ago, snapped to grain end`,

        interval: Interval.fromDateTimes(
          now.minus({ [unit1]: n1 - 1 }).startOf(unit1),
          now.minus({ [unit2]: n2 - 1 }).startOf(unit2),
        ),
      });
    });
  });

  return tests;
}

function generateLiteralTests(now: DateTime, n1?: number, n2?: number): Test[] {
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
        syntax: `-${n1}${grain}${LITERAL_CHARACTER} to -${n2}${grain2}${LITERAL_CHARACTER}`,
        description: `${n1} ${unit1}(s) ago to ${n2} ${unit2}(s) ago`,

        interval: Interval.fromDateTimes(
          now.minus({ [unit1]: n1 }),
          now.minus({ [unit2]: n2 }),
        ),
      });
    });
  });

  return tests;
}

function lastNPeriodExcludingCurrentPeriod(now: DateTime, upTo: number = 100) {
  const tests: Test[] = [];

  for (let i = 1; i <= upTo; i++) {
    GRAINS.forEach((grain) => {
      const unit = GRAIN_TO_LUXON[grain];

      const interval = Interval.fromDateTimes(
        now.minus({ [unit]: i }).startOf(unit),
        now.startOf(unit),
      );

      // Short
      tests.push({
        syntax: `-${i}${grain}${START_CHARACTER} to ${grain}${START_CHARACTER}`,
        description: `Last ${i} complete ${unit}(s) using short syntax`,
        interval,
      });

      // Explicit
      tests.push({
        syntax: `-${i}${grain}/${grain}${START_CHARACTER} to -0${grain}/${grain}${START_CHARACTER}`,
        description: `Last ${i} complete ${unit}(s) using explicit syntax`,
        interval,
      });

      // Using ending terminology
      tests.push({
        syntax: `${i}${grain} ${ENDING} ${grain}${START_CHARACTER}`,
        description: `Last ${i} complete ${unit}(s) using ending terminology`,
        interval,
      });

      // Using starting terminology
      tests.push({
        syntax: `${i}${grain} ${STARTING} -${i}${grain}${START_CHARACTER}`,
        description: `Last ${i} complete ${unit}(s) using starting terminology`,
        interval,
      });
    });
  }

  return tests;
}

export function generateStartingAndEndingTests(now: DateTime) {
  return GRAINS.map((periodGrain) => {
    const periodUnit = GRAIN_TO_LUXON[periodGrain];
    const periodInteger = Math.floor(Math.random() * 100);
    const offsetInteger = Math.floor(Math.random() * 100);

    return GRAINS.map((offsetGrain) => {
      const offsetUnit = GRAIN_TO_LUXON[offsetGrain];

      return [
        <Test>{
          syntax: `${periodInteger}${periodGrain} ${STARTING} -${offsetInteger}${offsetGrain}`,
          description: `The ${periodInteger} ${periodUnit} period starting exactly -${offsetInteger} ${offsetUnit}(s) ago`,
          interval: Interval.fromDateTimes(
            now.minus({ [offsetUnit]: offsetInteger }),
            now
              .minus({ [offsetUnit]: offsetInteger })
              .plus({ [periodUnit]: periodInteger }),
          ),
        },
        <Test>{
          syntax: `${periodInteger}${periodGrain} ${ENDING} -${offsetInteger}${offsetGrain}`,
          description: `The ${periodInteger} ${periodUnit} period ending exactly -${offsetInteger} ${offsetUnit}(s) ago`,
          interval: Interval.fromDateTimes(
            now
              .minus({ [offsetUnit]: offsetInteger })
              .minus({ [periodUnit]: periodInteger }),
            now.minus({ [offsetUnit]: offsetInteger }),
          ),
        },
        <Test>{
          syntax: `${periodInteger}${periodGrain} ${STARTING} -${offsetInteger}${offsetGrain}/${periodGrain}${START_CHARACTER}`,
          description: `The ${periodInteger} ${periodUnit} period starting exactly -${offsetInteger} ${offsetUnit}(s) ago rounded to the ${periodUnit} start`,
          interval: Interval.fromDateTimes(
            now.minus({ [offsetUnit]: offsetInteger }).startOf(periodUnit),
            now
              .minus({ [offsetUnit]: offsetInteger })
              .startOf(periodUnit)
              .plus({ [periodUnit]: periodInteger }),
          ),
        },
      ];
    });
  })
    .flat()
    .flat();
}

function lastNPeriodIncludingCurrentPeriod(now: DateTime, upTo: number = 100) {
  const tests: Test[] = [];

  for (let i = 1; i <= upTo; i++) {
    GRAINS.forEach((grain) => {
      const unit = GRAIN_TO_LUXON[grain];

      const interval = Interval.fromDateTimes(
        now.minus({ [unit]: i - 1 }).startOf(unit),
        now.plus({ [unit]: 1 }).startOf(unit),
      );

      // Short
      tests.push({
        syntax: `-${i - 1}${grain}${START_CHARACTER} to ${grain}${END_CHARACTER}`,
        description: `Last ${i} ${unit}(s) inclduing this ${unit}`,
        interval,
      });

      // Explicit
      tests.push({
        syntax: `-${i - 1}${grain}/${grain}${START_CHARACTER} to -0${grain}/${grain}${END_CHARACTER}`,
        description: `Last ${i} ${unit}(s) inclduing this ${unit}`,
        interval,
      });

      // Using +1
      tests.push({
        syntax: `-${i - 1}${grain}${START_CHARACTER} to +1${grain}${START_CHARACTER}`,
        description: `Last ${i} ${unit}(s) inclduing this ${unit}`,
        interval,
      });

      // Using ending terminology
      tests.push({
        syntax: `${i}${grain} ${ENDING} ${grain}${END_CHARACTER}`,
        description: `Last ${i} ${unit}(s) inclduing this ${unit}`,
        interval,
      });
    });
  }

  return tests;
}

function lastNPeriodToNow(now: DateTime, upTo: number = 100) {
  const tests: Test[] = [];

  for (let i = 1; i <= upTo; i++) {
    GRAINS.forEach((grain) => {
      const unit = GRAIN_TO_LUXON[grain];

      const interval = Interval.fromDateTimes(
        now.minus({ [unit]: i }).startOf(unit),
        now.plus({ millisecond: 1 }),
      );

      // Short
      tests.push({
        syntax: `-${i}${grain}${START_CHARACTER} to watermark`,
        description: `Last ${i} ${unit}(s) to watermark`,
        interval,
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

function generateCurrentAndPreviousPeriods(now: DateTime) {
  const tests: Test[] = [];

  GRAINS.forEach((grain) => {
    const unit = GRAIN_TO_LUXON[grain];

    const current = Interval.fromDateTimes(
      now.startOf(unit),
      now.plus({ [unit]: 1 }).startOf(unit),
    );

    const previous = Interval.fromDateTimes(
      now.minus({ [unit]: 1 }).startOf(unit),
      now.startOf(unit),
    );

    const toNow = Interval.fromDateTimes(
      now.startOf(unit),
      now.plus({ millisecond: 1 }),
    );

    // tests.push({
    //   syntax: `${grain}`,
    //   description: `The full current ${unit}`,
    //   interval: current,
    // });

    tests.push({
      syntax: `${grain}${START_CHARACTER} to watermark`,
      description: `Start of ${unit} to watermark`,
      interval: toNow,
    });

    tests.push({
      syntax: `1${grain} ${STARTING} ${grain}${START_CHARACTER}`,
      description: `The full current ${unit}`,
      interval: current,
    });

    tests.push({
      syntax: `1${grain} ${ENDING} ${grain}${END_CHARACTER}`,
      description: `The full current ${unit}`,
      interval: current,
    });

    tests.push({
      syntax: `-0${grain}/${grain}${START_CHARACTER} to -0${grain}/${grain}${END_CHARACTER}`,
      description: `The full current ${unit}`,
      interval: current,
    });

    tests.push({
      syntax: `-1${grain}${START_CHARACTER} to ${grain}${START_CHARACTER}`,
      description: `The last complete ${unit}`,
      interval: previous,
    });

    tests.push({
      syntax: `-1${grain}${INTERVAL_CHARACTER} `,
      description: `The last complete ${unit}`,
      interval: previous,
    });
  });

  return tests;
}

function generateFirstNPeriodTests(now: DateTime, n: number = 3) {
  const tests: Test[] = [];

  GRAINS.forEach((grain) => {
    const unit = GRAIN_TO_LUXON[grain];

    const higherOrderGrain = getHigherOrderGrain(grain);

    if (!higherOrderGrain) return;
    const higherOrderUnit = GRAIN_TO_LUXON[higherOrderGrain];

    let start = now.startOf(higherOrderUnit);

    if (unit === "week") {
      if (start.weekday >= 5) {
        start = start.startOf("week").plus({ week: 1 });
      } else {
        start = start.startOf("week");
      }

      tests.push({
        syntax: `${higherOrderGrain}/${higherOrderGrain}W${START_CHARACTER} to ${higherOrderGrain}/${higherOrderGrain}W${START_CHARACTER}+${n}${grain}`,
        description: `First ${n} ${unit}(s) of the current ${higherOrderUnit}`,
        interval: Interval.fromDateTimes(start, start.plus({ [unit]: n })),
      });

      tests.push({
        syntax: `${n}${grain} starting ${higherOrderGrain}/${higherOrderGrain}W${START_CHARACTER}`,
        description: `First ${n} ${unit}(s) of the current ${higherOrderUnit} using starting terminology`,
        interval: Interval.fromDateTimes(start, start.plus({ [unit]: n })),
      });
    } else {
      tests.push({
        syntax: `${higherOrderGrain}${START_CHARACTER} to ${higherOrderGrain}${START_CHARACTER}+${n}${grain}`,
        description: `First ${n} ${unit}(s) of the current ${higherOrderUnit}`,
        interval: Interval.fromDateTimes(start, start.plus({ [unit]: n })),
      });

      tests.push({
        syntax: `${n}${grain} starting ${higherOrderGrain}${START_CHARACTER}`,
        description: `First ${n} ${unit}(s) of the current ${higherOrderUnit} using starting terminology`,
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

  const testCases = generateTestCases(allTime.end?.setZone("UTC"));

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

    if (response?.timeRanges?.[i]?.start !== apiFormat.start) {
      console.log(testCases[i].description);
      console.log(
        `${failures.toLocaleString("en-US", {
          minimumIntegerDigits: 2,
          useGrouping: false,
        })}. FAIL (start) // ${testCases[i].syntax.padEnd(36)} // Expected: ${String(apiFormat.start).padEnd(20)} // Got: ${response?.timeRanges?.[i]?.start}`,
      );
      failures++;
    } else if (response?.timeRanges?.[i]?.end !== apiFormat.end) {
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

function generateTestCases(now: DateTime): Test[] {
  return [
    ...generateToTests(now, 6, 3),
    ...generateToTests(now, 4, 1),
    ...generateToTests(now, 9, 7),
    ...generateToTests(now),
    ...lastNPeriodExcludingCurrentPeriod(now),
    ...lastNPeriodIncludingCurrentPeriod(now),
    ...generateFirstNPeriodTests(now),
    ...generateLiteralTests(now),
    ...generateEndTests(now),
    ...generateCurrentAndPreviousPeriods(now),
    ...lastNPeriodToNow(now),
    ...generateStartingAndEndingTests(now),
    ...generateISOTests(),

    {
      syntax: `-6d${START_CHARACTER} to h${START_CHARACTER}`,
      description: "Six days to date excluding the current hour",
      interval: Interval.fromDateTimes(
        now.minus({ day: 6 }).startOf("day"),
        now.startOf("hour"),
      ),
    },

    {
      syntax: `-2h/h${START_CHARACTER}-3d${LITERAL_CHARACTER} to -2h/h${START_CHARACTER}`,
      description: "The 3 day period starting two hour boundaries ago",
      interval: Interval.fromDateTimes(
        now.minus({ hour: 2 }).startOf("hour").minus({ day: 3 }),
        now.minus({ hour: 2 }).startOf("hour"),
      ),
    },
    {
      syntax: `-2s-3m-4h-5d-6w-7M-8Q-9Y to -1s-2m-3h-4d-5w-6M-7Q-8Y`,
      description: "Six days to date excluding the current hour",
      interval: Interval.fromDateTimes(
        now.minus({
          second: 2,
          minute: 3,
          hour: 4,
          day: 5,
          week: 6,
          month: 7,
          quarter: 8,
          year: 9,
        }),
        now.minus({
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
      syntax: `-2s3m4h5d6w7M8Q9Y to -1s2m3h4d5w6M7Q8Y`,
      description:
        "Six days to date excluding the current hour (without minuses",
      interval: Interval.fromDateTimes(
        now.minus({
          second: 2,
          minute: 3,
          hour: 4,
          day: 5,
          week: 6,
          month: 7,
          quarter: 8,
          year: 9,
        }),
        now.minus({
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
      syntax: `-42W-3M/W${START_CHARACTER} to -2M-3h/M${END_CHARACTER}`,
      description:
        "The start of the week 42 weeks and 3 minutes ago to the end of the month 2 months and 3 hours ago",
      interval: Interval.fromDateTimes(
        now
          .minus({
            week: 42,
            month: 3,
          })
          .startOf("week"),

        now
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
      syntax: `-3Y/d${START_CHARACTER}+6h-6W/h${START_CHARACTER} to -4d/M${END_CHARACTER}-23h-7m`,
      description:
        "The day start three days ago plus 6 hours minus 6 weeks rounded to the hour start TO the month end three years ago minus 23 hours and 7 minutes",
      interval: Interval.fromDateTimes(
        now
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

        now
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
      syntax: `3W18D23h ${STARTING} -3Y${LITERAL_CHARACTER}`,
      description:
        "The three week, 18 day, and 23 hour period starting exactly three years ago",
      interval: Interval.fromDateTimes(
        now.minus({
          year: 3,
        }),
        now
          .minus({
            year: 3,
          })
          .plus({
            week: 3,
            day: 18,
            hour: 23,
          }),
      ),
    },

    {
      syntax: `2h ${STARTING} -4h${LITERAL_CHARACTER}`,
      description: "The two hour period starting exactly four hours ago",
      interval: Interval.fromDateTimes(
        now.minus({
          hour: 4,
        }),
        now
          .minus({
            hour: 4,
          })
          .plus({
            hour: 2,
          }),
      ),
    },

    {
      syntax: `-2h to latest`,
      description: "The two hour period ending at the latest data point",
      interval: Interval.fromDateTimes(
        now.minus({
          hour: 2,
        }),
        now.plus({
          millisecond: 1,
        }),
      ),
    },

    {
      syntax: `3W18D23h ${ENDING} -3Y${LITERAL_CHARACTER}`,
      description:
        "The three week, 18 day, and 23 hour period ending exactly three years ago",
      interval: Interval.fromDateTimes(
        now
          .minus({
            year: 3,
          })
          .minus({
            week: 3,
            day: 18,
            hour: 23,
          }),
        now.minus({
          year: 3,
        }),
      ),
    },

    {
      syntax: `<6H of D25 as of -3M`,
      description:
        "The first 6 hours of the 25th day of the month three months ago",
      interval: Interval.fromDateTimes(
        now.minus({ month: 3 }).set({ day: 25 }).startOf("day"),
        now
          .minus({ month: 3 })
          .set({ day: 25 })
          .startOf("day")
          .plus({ hours: 6 }),
      ),
    },
    {
      syntax: `6h ${STARTING} D25^ as of -3M`,
      description:
        "The first 6 hours of the 25th day of the month three months ago",
      interval: Interval.fromDateTimes(
        now.minus({ month: 3 }).set({ day: 25 }).startOf("day"),
        now
          .minus({ month: 3 })
          .set({ day: 25 })
          .startOf("day")
          .plus({ hours: 6 }),
      ),
    },
    {
      syntax: `7h ${ENDING} -1d${END_CHARACTER}`,
      description: "The last 7 hours of the previous day",
      interval: Interval.fromDateTimes(
        now.startOf("day").minus({ hour: 7 }),
        now.startOf("day"),
      ),
    },
    {
      syntax: `7h ${ENDING} -1d${END_CHARACTER}`,
      description: "The last 7 hours of the previous day",
      interval: Interval.fromDateTimes(
        now.startOf("day").minus({ hour: 7 }),
        now.startOf("day"),
      ),
    },
    {
      syntax: `>7h of -1d${INTERVAL_CHARACTER}`,
      description: "The last 7 hours of the previous day",
      interval: Interval.fromDateTimes(
        now.startOf("day").minus({ hour: 7 }),
        now.startOf("day"),
      ),
    },
    {
      syntax: `H4 as of -1d`,
      description: "The fourth hour of the previous day",
      interval: Interval.fromDateTimes(
        now.minus({ day: 1 }).set({ hour: 3 }).startOf("hour"),
        now.minus({ day: 1 }).set({ hour: 4 }).startOf("hour"),
      ),
    },
    {
      syntax: `H4 of -1d${INTERVAL_CHARACTER}`,
      description: "The fourth hour of the previous day",
      interval: Interval.fromDateTimes(
        now.minus({ day: 1 }).set({ hour: 3 }).startOf("hour"),
        now.minus({ day: 1 }).set({ hour: 4 }).startOf("hour"),
      ),
    },
    {
      syntax: `-2d${START_CHARACTER} to -2d${START_CHARACTER}+3h`,
      description: "The first 3 hours of two days ago",
      interval: Interval.fromDateTimes(
        now.minus({ day: 2 }).startOf("day"),
        now.minus({ day: 2 }).set({ hour: 3 }).startOf("hour"),
      ),
    },
    ...[1, 2, 4, 13, 26, 39, 52, 53, 54].map((value) => {
      return {
        syntax: `W${value} of Y`,
        description: `Week ${value} of the current year`,
        interval: Interval.fromDateTimes(
          DateTime.fromObject(
            {
              weekYear: now.year,
              weekNumber: value,
            },
            { zone: now.zone },
          ),
          DateTime.fromObject(
            {
              weekYear: now.year,
              weekNumber: value,
            },
            { zone: now.zone },
          )
            .plus({ week: 1 })
            .startOf("week"),
        ),
      };
    }),

    {
      syntax: `W1 of -1M${INTERVAL_CHARACTER}`,
      description:
        "The first week of the last complete month following ISO rules",
      interval: Interval.fromDateTimes(
        now.minus({ month: 1 }).startOf("month").weekday >= 5
          ? now
              .minus({ month: 1 })
              .startOf("month")
              .startOf("week")
              .plus({ week: 1 })
          : now.minus({ month: 1 }).startOf("month").startOf("week"),
        (now.minus({ month: 1 }).startOf("month").weekday >= 5
          ? now
              .minus({ month: 1 })
              .startOf("month")
              .plus({ week: 1 })
              .startOf("week")
          : now.minus({ month: 1 }).startOf("month").startOf("week")
        )
          .plus({ week: 1 })
          .startOf("week"),
      ),
    },
    {
      syntax: `W1 of -1M${START_CHARACTER} to -1M${END_CHARACTER}`,
      description:
        "The first week of the last complete month following ISO rules",
      interval: Interval.fromDateTimes(
        now.minus({ month: 1 }).startOf("month").weekday >= 5
          ? now
              .minus({ month: 1 })
              .startOf("month")
              .startOf("week")
              .plus({ week: 1 })
          : now.minus({ month: 1 }).startOf("month").startOf("week"),
        (now.minus({ month: 1 }).startOf("month").weekday >= 5
          ? now
              .minus({ month: 1 })
              .startOf("month")
              .plus({ week: 1 })
              .startOf("week")
          : now.minus({ month: 1 }).startOf("month").startOf("week")
        )
          .plus({ week: 1 })
          .startOf("week"),
      ),
    },
    {
      syntax: "W1 of M",
      description: "The first week of the current month following ISO rules",
      interval: Interval.fromDateTimes(
        now.startOf("month").weekday >= 5
          ? now.startOf("month").plus({ week: 1 }).startOf("week")
          : now.startOf("month").startOf("week"),
        (now.startOf("month").weekday >= 5
          ? now.startOf("month").plus({ week: 1 }).startOf("week")
          : now.startOf("month").startOf("week")
        )
          .plus({ week: 1 })
          .startOf("week"),
      ),
    },

    {
      syntax: `>3s of m4 of H2 of D4 of -1M${INTERVAL_CHARACTER}`,
      description:
        "The last three seconds of the fourth minute of the second hour of the 4th day of last month",
      interval: Interval.fromDateTimes(
        now
          .minus({ month: 1 })
          .set({ day: 4 })
          .set({ hour: 1 })
          .set({ minute: 3 })
          .set({ second: 57 })
          .startOf("second"),
        now
          .minus({ month: 1 })
          .set({ day: 4 })
          .set({ hour: 1 })
          .set({ minute: 4 })
          .startOf("minute"),
      ),
    },
    {
      syntax: `-2d${START_CHARACTER} to d${END_CHARACTER} as of -1Q`,
      description:
        "The 3 day period that includes the current incomplete day and the 2 previous complete days offset into the previous quarter",
      interval: Interval.fromDateTimes(
        now.minus({ day: 2 }).startOf("day").minus({ quarter: 1 }),
        now
          .minus({ day: 2 })
          .startOf("day")
          .minus({ quarter: 1 })
          .plus({ day: 3 })
          .startOf("day"),
      ),
    },
    {
      syntax: `3d ${ENDING} -1Q/d${END_CHARACTER}`,
      description:
        "The 3 day period that includes the current incomplete day and the 2 previous complete days offset into the previous quarter",
      interval: Interval.fromDateTimes(
        now.minus({ day: 2 }).startOf("day").minus({ quarter: 1 }),
        now
          .minus({ day: 2 })
          .startOf("day")
          .minus({ quarter: 1 })
          .plus({ day: 3 })
          .startOf("day"),
      ),
    },
    {
      syntax: `3d ${ENDING} -1Q/d${END_CHARACTER}`,
      description:
        "The 3 day period that includes the current incomplete day and the 2 previous complete days offset into the previous quarter",
      interval: Interval.fromDateTimes(
        now.minus({ day: 2 }).startOf("day").minus({ quarter: 1 }),
        now
          .minus({ day: 2 })
          .startOf("day")
          .minus({ quarter: 1 })
          .plus({ day: 3 })
          .startOf("day"),
      ),
    },
    {
      syntax: "<6h of W4 of Q2",
      description:
        "The first 6 hours of week 4 of quarter 2 of the current year",
      interval: Interval.fromDateTimes(
        now.startOf("year").plus({ quarter: 1 }).weekday >= 5
          ? now
              .startOf("year")
              .plus({ quarter: 1 })
              .plus({ week: 4 })
              .startOf("week")
          : now
              .startOf("year")
              .plus({ quarter: 1 })
              .plus({ week: 3 })
              .startOf("week"),
        (now.startOf("year").plus({ quarter: 1 }).weekday >= 5
          ? now
              .startOf("year")
              .plus({ quarter: 1 })
              .plus({ week: 4 })
              .startOf("week")
          : now
              .startOf("year")
              .plus({ quarter: 1 })
              .plus({ week: 3 })
              .startOf("week")
        )
          .plus({ hour: 6 })
          .startOf("hour"),
      ),
    },
    {
      syntax: `m30 of H12 of D5 of >1W of Q3 as of -2Y`,
      description:
        "The 30th minute of the 12th hour of the fifth day of the last week of quarter three of two years ago",
      interval: Interval.fromDateTimes(
        (now.minus({ year: 2 }).startOf("year").plus({ quarter: 3 }).weekday >=
        5
          ? now.minus({ year: 2 }).startOf("year").plus({ quarter: 3 })
          : now
              .minus({ year: 2 })
              .startOf("year")
              .plus({ quarter: 3 })
              .minus({ week: 1 })
        )
          .startOf("week")
          .plus({ day: 4 })
          .set({ hour: 11 })
          .set({ minute: 29 })
          .startOf("minute"),
        (now.minus({ year: 2 }).startOf("year").plus({ quarter: 3 }).weekday >=
        5
          ? now.minus({ year: 2 }).startOf("year").plus({ quarter: 3 })
          : now
              .minus({ year: 2 })
              .startOf("year")
              .plus({ quarter: 3 })
              .minus({ week: 1 })
        )
          .startOf("week")
          .plus({ day: 4 })
          .set({ hour: 11 })
          .set({ minute: 30 })
          .startOf("minute"),
      ),
    },

    {
      syntax: `Y${START_CHARACTER} to M${START_CHARACTER}`,
      description: "All complete months of the current year",
      interval: Interval.fromDateTimes(
        now.startOf("year"),
        now.startOf("month"),
      ),
    },
    {
      syntax: `Y${START_CHARACTER} to M${END_CHARACTER}`,
      description:
        "All months of the current year, including the current incomplete one",
      interval: Interval.fromDateTimes(
        now.startOf("year"),
        now.plus({ month: 1 }).startOf("month"),
      ),
    },
    {
      syntax: `Y${START_CHARACTER} to D${START_CHARACTER}`,
      description: "All complete days of the current year",
      interval: Interval.fromDateTimes(now.startOf("year"), now.startOf("day")),
    },
    {
      syntax: `Y${START_CHARACTER} to D${END_CHARACTER}`,
      description:
        "All days of the current year, including the current incomplete one",
      interval: Interval.fromDateTimes(
        now.startOf("year"),
        now.plus({ day: 1 }).startOf("day"),
      ),
    },

    {
      syntax: `Y/YW${START_CHARACTER} to W${START_CHARACTER}`,
      description:
        "All weeks of the current year, excluding the current incomplete one",
      interval: Interval.fromDateTimes(
        DateTime.fromObject(
          {
            weekYear: now.year,
            weekNumber: 1,
          },
          { zone: now.zone },
        ),
        now.startOf("week"),
      ),
    },

    {
      syntax: `-127d/W${START_CHARACTER} to -89d/W${START_CHARACTER}`,
      description:
        "The current week start 127 days ago to the current week start 89 days ago",
      interval: Interval.fromDateTimes(
        now.minus({ day: 127 }).startOf("week"),
        now.minus({ day: 89 }).startOf("week"),
      ),
    },

    {
      syntax: `-15Y/W${START_CHARACTER} to -4Y/W${START_CHARACTER}`,
      description:
        "The current week start 15 years ago to the current week start 4 years ago",
      interval: Interval.fromDateTimes(
        now.minus({ year: 15 }).startOf("week"),
        now.minus({ year: 4 }).startOf("week"),
      ),
    },

    {
      syntax: `-2Y/YW${START_CHARACTER} to -2Y/W${START_CHARACTER}`,
      description: "All weeks of the current year, offset into two years ago",
      interval: Interval.fromDateTimes(
        DateTime.fromObject(
          {
            weekYear: now.year - 2,
            weekNumber: 1,
          },
          { zone: now.zone },
        ),
        now.minus({ year: 2 }).startOf("week"),
      ),
    },
    {
      syntax: `-2Y/YW${START_CHARACTER} to -2Y/YW${END_CHARACTER}`,
      description: "The year two years ago by ISO week rules",
      interval: Interval.fromDateTimes(
        DateTime.fromObject(
          {
            weekYear: now.year - 2,
            weekNumber: 1,
          },
          { zone: now.zone },
        ),
        DateTime.fromObject(
          {
            weekYear: now.year - 1,
            weekNumber: 1,
          },
          { zone: now.zone },
        ),
      ),
    },

    {
      syntax: `D3 of W2 of -2M${INTERVAL_CHARACTER}`,
      description: "Day 3 of week 2 of two months ago",
      interval: Interval.fromDateTimes(
        (now.minus({ month: 2 }).startOf("month").weekday >= 5
          ? now
              .minus({ month: 2 })
              .startOf("month")
              .startOf("week")
              .plus({ week: 1 })
          : now.minus({ month: 2 }).startOf("month").startOf("week")
        ).plus({ day: 2, week: 1 }),

        (now.minus({ month: 2 }).startOf("month").weekday >= 5
          ? now
              .minus({ month: 2 })
              .startOf("month")
              .startOf("week")
              .plus({ week: 1 })
          : now.minus({ month: 2 }).startOf("month").startOf("week")
        ).plus({ day: 3, week: 1 }),
      ),
    },

    {
      syntax: `94m ${STARTING} -9d`,
      description: "The 94 minute period starting exactly 9 days ago",
      interval: Interval.fromDateTimes(
        now.minus({ day: 9 }),
        now.minus({ day: 9 }).plus({ minute: 94 }),
      ),
    },

    {
      syntax: `94Q ${STARTING} -9d`,
      description: "The 94 quarter period starting exactly 9 days ago",
      interval: Interval.fromDateTimes(
        now.minus({ day: 9 }),
        now.minus({ day: 9 }).plus({ quarter: 94 }),
      ),
    },

    {
      syntax: `94m ${ENDING} -9d/m^`,
      description:
        "The 94 minute period ending 9 days ago, rounded to the minute",
      interval: Interval.fromDateTimes(
        now.minus({ day: 9 }).startOf("minute").minus({ minute: 94 }),
        now.minus({ day: 9 }).startOf("minute"),
      ),
    },
    {
      syntax: `94m ${ENDING} -9d`,
      description: "The 94 minute period ending exactly 9 days ago",
      interval: Interval.fromDateTimes(
        now.minus({ day: 9 }).minus({ minute: 94 }),
        now.minus({ day: 9 }),
      ),
    },
    // This fails
    {
      syntax: `-1y/W^ to -1y/W$ as of 2025-05-17T13:43:00Z`,
      description: "The one week period starting two years ago from 05/14/2025",
      interval: Interval.fromDateTimes(
        DateTime.fromISO("2025-05-17T13:43:00Z", { zone: "UTC" })
          .minus({ year: 1 })
          .startOf("week"),
        DateTime.fromISO("2025-05-17T13:43:00Z", { zone: "UTC" })
          .minus({ year: 1 })
          .plus({ week: 1 })
          .startOf("week"),
      ),
    },
    // This doesn't fail (but they should both be the same)
    {
      syntax: `-1y/W^ to -1y/W$ as of 2025-05-15T13:43:00Z`,
      description: "The one week period starting two years ago from 05/14/2025",
      interval: Interval.fromDateTimes(
        DateTime.fromISO("2025-05-15T13:43:00Z", { zone: "UTC" })
          .minus({ year: 1 })
          .startOf("week"),
        DateTime.fromISO("2025-05-15T13:43:00Z", { zone: "UTC" })
          .minus({ year: 1 })
          .plus({ week: 1 })
          .startOf("week"),
      ),
    },
  ];
}
