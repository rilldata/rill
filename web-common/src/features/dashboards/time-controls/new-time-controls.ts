// WIP as of 07/22/2025
// The intention of this file is to start from scratch building a new time control system
// The majority of this work is being implemented in the Canvas TimeControls class
// IntervalStore and MetricsTimeControls are WIP references, but are not currently being used
// The functions below UTILS are being used

import { fetchTimeRanges } from "@rilldata/web-common/features/dashboards/time-controls/rill-time-ranges.ts";
import {
  overrideRillTimeRef,
  parseRillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import type { V1ExploreTimeRange } from "@rilldata/web-common/runtime-client";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  DateTime,
  type DateTimeUnit,
  Duration,
  type DurationObjectUnits,
  IANAZone,
  Interval,
  type WeekdayNumbers,
} from "luxon";
import { get, writable, type Writable } from "svelte/store";

// CONSTANTS -> time-control-constants.ts

export const RILL_TO_UNIT: Record<
  RillPeriodToDate | RillPreviousPeriod,
  DateTimeUnit
> = {
  "rill-PDC": "day",
  "rill-PWC": "week",
  "rill-PMC": "month",
  "rill-PQC": "quarter",
  "rill-PYC": "year",
  "rill-TD": "day",
  "rill-WTD": "week",
  "rill-MTD": "month",
  "rill-QTD": "quarter",
  "rill-YTD": "year",
};

export const RILL_TO_LABEL: Record<
  RillPeriodToDate | RillPreviousPeriod | AllTime | CustomRange,
  string
> = {
  inf: "All Time",
  CUSTOM: "Custom",
  "rill-PDC": "Yesterday",
  "rill-PWC": "Previous week",
  "rill-PMC": "Previous month",
  "rill-PQC": "Previous quarter",
  "rill-PYC": "Previous year",
  "rill-TD": "Today",
  "rill-WTD": "Week to date",
  "rill-MTD": "Month to date",
  "rill-QTD": "Quarter to date",
  "rill-YTD": "Year to date",
};

export const RILL_PERIOD_TO_DATE = [
  "rill-TD",
  "rill-WTD",
  "rill-MTD",
  "rill-QTD",
  "rill-YTD",
] as const;

export const RILL_PREVIOUS_PERIOD = [
  "rill-PDC",
  "rill-PWC",
  "rill-PMC",
  "rill-PQC",
  "rill-PYC",
] as const;

export const RILL_LATEST = [
  "PT6H",
  "PT24H",
  "P7D",
  "P14D",
  "P4W",
  "P12M",
] as const;

export const TIME_GRAIN_TO_SHORTHAND: Record<V1TimeGrain, string> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "",
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "ms",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "s",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "m",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "H",
  [V1TimeGrain.TIME_GRAIN_DAY]: "D",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "W",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "M",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "Q",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "Y",
};

// TYPES -> time-control-types.ts

type RillPeriodToDateTuple = typeof RILL_PERIOD_TO_DATE;
export type RillPeriodToDate = RillPeriodToDateTuple[number];

type RillPreviousPeriodTuple = typeof RILL_PREVIOUS_PERIOD;
export type RillPreviousPeriod = RillPreviousPeriodTuple[number];

type RillLatestTuple = typeof RILL_LATEST;
export type RillLatest = RillLatestTuple[number];

export const CUSTOM_TIME_RANGE_ALIAS = "CUSTOM";
export const ALL_TIME_RANGE_ALIAS = "inf";
export type AllTime = typeof ALL_TIME_RANGE_ALIAS;
export type CustomRange = typeof CUSTOM_TIME_RANGE_ALIAS;
export type ISODurationString = string;

export type NamedRange =
  | RillPeriodToDate
  | RillPreviousPeriod
  | AllTime
  | CustomRange;

// STORES -> time-control-stores.ts

class IntervalStore {
  private _interval: Writable<Interval> = writable(
    Interval.invalid("Uninitialized"),
  );

  subscribe = this._interval.subscribe;

  clear = () => {
    this._interval.set(Interval.invalid("Uninitialized"));
  };

  updateInterval(interval: Interval) {
    this._interval.set(interval);
  }

  updateEnd(end: DateTime) {
    this._interval.update((i) => i.set({ end }));
  }

  updateStart(start: DateTime) {
    this._interval.update((i) => i.set({ start }));
  }

  updateZone(zone: IANAZone, keepLocalTime = false) {
    this._interval.update((i) =>
      i.mapEndpoints((e) => e.setZone(zone, { keepLocalTime: keepLocalTime })),
    );
  }
}

class MetricsTimeControls {
  private _maxRange = new IntervalStore();
  private _zone: Writable<IANAZone> = writable(new IANAZone("UTC"));
  private _selected = writable<NamedRange | ISODurationString>(
    ALL_TIME_RANGE_ALIAS,
  );
  private _visibleRange = new IntervalStore();
  private _subrange = new IntervalStore();
  private _comparisonRange = new IntervalStore();
  private _showComparison: Writable<boolean> = writable(false);
  private _metricsViewName: string;

  constructor(maxStart: DateTime, maxEnd: DateTime, metricsViewName: string) {
    this._metricsViewName = metricsViewName;
    const maxInterval = Interval.fromDateTimes(
      maxStart.setZone("UTC"),
      maxEnd.setZone("UTC"),
    );
    this._maxRange.updateInterval(maxInterval);
    this._visibleRange.updateInterval(maxInterval);
  }

  private applySubrange = () => {
    this._visibleRange.updateInterval(get(this._subrange));
    this._selected.set(CUSTOM_TIME_RANGE_ALIAS);
    this._subrange.clear();
  };

  private applyISODuration = async (iso: ISODurationString) => {
    const rightAnchor = get(this._maxRange).end;
    if (rightAnchor) {
      const interval = await deriveInterval(
        iso,

        this._metricsViewName,
        get(this._zone).name,
      );
      if (interval?.interval.isValid) {
        this._visibleRange.updateInterval(interval.interval);
        this._selected.set(iso);
      }
    }
  };

  private applyNamedRange = async (name: NamedRange) => {
    const rightAnchor = get(this._maxRange).end;
    if (rightAnchor) {
      const interval = await deriveInterval(
        name,

        this._metricsViewName,
        get(this._zone).name,
      );
      if (interval?.interval.isValid) {
        this._visibleRange.updateInterval(interval.interval);
        this._selected.set(name);
      }
    }
  };

  private applyCustomRange = (start: DateTime, end: DateTime) => {
    this._visibleRange.updateInterval(Interval.fromDateTimes(start, end));
    this._selected.set(CUSTOM_TIME_RANGE_ALIAS);
  };

  private applyRange = (string: NamedRange | ISODurationString | undefined) => {
    if (!string) return;

    if (string === ALL_TIME_RANGE_ALIAS) {
      this.applyAllTime();
    } else if (isRillPeriodToDate(string) || isRillPreviousPeriod(string)) {
      this.applyNamedRange(string);
    } else if (isValidISODuration(string)) {
      this.applyISODuration(string);
    } else if (string === CUSTOM_TIME_RANGE_ALIAS) {
      throw new Error("Custom time range requires start and end dates");
    } else {
      throw new Error("Invalid time range");
    }
  };

  private applyAllTime = () => {
    const maxInterval = get(this._maxRange);
    if (maxInterval) {
      this._visibleRange.updateInterval(maxInterval);
      this._selected.set(ALL_TIME_RANGE_ALIAS);
    }
  };

  updateZone = (zone: IANAZone) => {
    this._zone.set(zone);
    const rangeAlias = get(this._selected);
    this._maxRange.updateZone(zone);

    if (rangeAlias === CUSTOM_TIME_RANGE_ALIAS) {
      // If you've specified a custom range
      // We want to maintain the local time of the start and end
      this._visibleRange.updateZone(zone, true);
    } else {
      // Otherwise, we need to re-derive the interval based on the selection
      this.applyRange(rangeAlias);
    }
  };

  apply = {
    subrange: this.applySubrange,
    customRange: this.applyCustomRange,
    range: this.applyRange,
    allTime: this.applyAllTime,
  };

  switchComparison(bool?: boolean) {
    if (bool === undefined) {
      this._showComparison.update((b) => !b);
    } else {
      this._showComparison.set(bool);
    }
  }

  zone = this._zone;
  selected = this._selected;
  subrange = this._subrange;
  visibleRange = this._visibleRange;
  comparisonRange = this._comparisonRange;
}

class TimeControls {
  private _timeControls = new Map<string, MetricsTimeControls>();

  get(metricsViewName: string, maxStart?: DateTime, maxEnd?: DateTime) {
    let store = this._timeControls.get(metricsViewName);

    if (!store && maxStart && maxEnd) {
      store = new MetricsTimeControls(maxStart, maxEnd, metricsViewName);
      this._timeControls.set(metricsViewName, store);
    } else if (!store) {
      throw new Error("TimeControls.get() called without maxStart and maxEnd");
    }

    return store;
  }
}

export const timeControls = new TimeControls();

// UTILS -> time-control-utils.ts

export function isRillPreviousPeriod(
  value: string,
): value is RillPreviousPeriod {
  return RILL_PREVIOUS_PERIOD.includes(value as RillPreviousPeriod);
}

export function isRillPeriodToDate(value: string): value is RillPeriodToDate {
  return RILL_PERIOD_TO_DATE.includes(value as RillPeriodToDate);
}

import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  getAllowedGrains,
  GrainAliasToV1TimeGrain,
  V1TimeGrainToAlias,
} from "@rilldata/web-common/lib/time/new-grains";
import {
  RillLegacyDaxInterval,
  RillLegacyIsoInterval,
  RillPeriodToGrainInterval,
  RillShorthandInterval,
  RillTimeLabel,
  RillTimeStartEndInterval,
  type RillTime,
} from "../url-state/time-ranges/RillTime";
import { getDefaultRangeBuckets } from "@rilldata/web-common/lib/time/defaults";

export async function deriveInterval(
  name: RillPeriodToDate | RillPreviousPeriod | ISODurationString | string,
  metricsViewName: string,
  activeTimeZone: string,
): Promise<{
  interval: Interval;
  grain?: V1TimeGrain | undefined;
  error?: string;
}> {
  if (name === CUSTOM_TIME_RANGE_ALIAS) {
    return {
      interval: Interval.invalid("Cannot derive interval for custom range"),
      grain: undefined,
      error: "Cannot derive interval for custom range",
    };
  }

  try {
    const parsed = parseRillTime(name);

    // We have a RillTime string
    const instanceId = get(runtime).instanceId;
    const cacheBust = name.includes("now");

    const response = await fetchTimeRanges({
      instanceId,
      metricsViewName,
      rillTimes: [name],
      timeZone: activeTimeZone,
      cacheBust,
    });

    const timeRange = response.resolvedTimeRanges?.[0];

    if (!timeRange?.start || !timeRange?.end) {
      return { interval: Interval.invalid("Invalid time range") };
    }

    return {
      interval: Interval.fromDateTimes(
        DateTime.fromISO(timeRange.start).setZone(activeTimeZone),
        DateTime.fromISO(timeRange.end).setZone(activeTimeZone),
      ),
      grain: parsed.asOfLabel?.snap
        ? GrainAliasToV1TimeGrain[parsed.asOfLabel?.snap]
        : parsed.rangeGrain,
    };
  } catch (error) {
    console.error("Error deriving interval:", error);
    return {
      interval: Interval.invalid("Unable to derive interval"),
      grain: undefined,
      error: "Error deriving interval",
    };
  }
}

export function getPeriodToDate(date: DateTime, period: DateTimeUnit) {
  const periodStart = date.startOf(period, { useLocaleWeeks: true });
  const exclusiveEnd = date.endOf("day").plus({ millisecond: 1 });

  return Interval.fromDateTimes(periodStart, exclusiveEnd);
}

export function normalizeWeekday(
  possibleWeekday: number | undefined,
): WeekdayNumbers {
  if (
    possibleWeekday === undefined ||
    possibleWeekday <= 0 ||
    possibleWeekday >= 8
  )
    return 1;

  return possibleWeekday as WeekdayNumbers;
}

export function getPreviousPeriodComplete(
  anchor: DateTime,
  period: DateTimeUnit,
  steps = 0,
) {
  const startOfCurrentPeriod = anchor.startOf(period, { useLocaleWeeks: true });
  const shiftedStart = startOfCurrentPeriod.minus({ [period + "s"]: steps });
  const exclusiveEnd = shiftedStart
    .endOf(period, { useLocaleWeeks: true })
    .plus({ millisecond: 1 });

  return Interval.fromDateTimes(shiftedStart, exclusiveEnd);
}

export function getInterval(
  luxonDuration: Duration,
  endDate: DateTime,
  full = true,
) {
  const durationUnits = luxonDuration.toObject();
  const smallestUnit = getSmallestUnit(durationUnits);

  const end =
    smallestUnit && full
      ? endDate
          .endOf(smallestUnit, { useLocaleWeeks: true })
          .plus({ millisecond: 1 })
      : endDate;

  return Interval.before(end, durationUnits);
}

export function getSmallestUnit(
  units: DurationObjectUnits,
): DateTimeUnit | null {
  if (units.milliseconds) return "millisecond";
  if (units.seconds) return "second";
  if (units.minutes) return "minute";
  if (units.hours) return "hour";
  if (units.days) return "day";
  if (units.weeks) return "week";
  if (units.months) return "month";
  if (units.quarters) return "quarter";
  if (units.years) return "year";

  return null;
}

export function getSmallestUnitInDateTime(time: DateTime): DateTimeUnit | null {
  if (time.millisecond) return "millisecond";
  if (time.second) return "second";
  if (time.minute) return "minute";
  if (time.hour) return "hour";
  if (time.day) return "day";
  if (time.month) return "month";
  if (time.quarter) return "quarter";
  if (time.year) return "year";

  return null;
}

export function isValidISODuration(duration: string) {
  const luxonDuration = Duration.fromISO(duration);

  if (luxonDuration.isValid) return luxonDuration;
  return null;
}

export function getDurationLabel(isoDuration: string): string {
  if (!isValidISODuration(isoDuration)) {
    throw new Error("Invalid ISO duration");
  }

  return `Last ${humaniseISODuration(isoDuration)}`;
}

export function getRangeLabel(range: string | undefined): string {
  if (!range) return "Custom";
  if (isRillPeriodToDate(range) || isRillPreviousPeriod(range)) {
    return RILL_TO_LABEL[range];
  }

  if (range === ALL_TIME_RANGE_ALIAS || range === CUSTOM_TIME_RANGE_ALIAS) {
    return RILL_TO_LABEL[range];
  }

  if (isValidISODuration(range)) {
    return getDurationLabel(range);
  }

  try {
    const rt = parseRillTime(range);

    const label = rt.getLabel();

    return label;
  } catch (e) {
    console.error("Error parsing RillTime", e);
    return "Custom";
  }
}

export type RangeBuckets = {
  custom: RillTime[];
  latest: RillTime[];
  periodToDate: RillTime[];
  previous: RillTime[];
  allTime: boolean;
};

const defaultBuckets: RangeBuckets = {
  latest: RILL_LATEST.map((r) => parseRillTime(r)),
  periodToDate: RILL_PERIOD_TO_DATE.map((r) => parseRillTime(r)),
  previous: RILL_PREVIOUS_PERIOD.map((r) => parseRillTime(r)),
  custom: [],
  allTime: false,
};

const previousPeriodRegex =
  /-\d+[sSmMhHdDwWqQYy]\/[sSmMhHdDwWqQYy]\s+to\s+ref\/[sSmMhHdDwWqQYy]/;

export function bucketYamlRanges(
  yamlRanges: V1ExploreTimeRange[],
  minTimeGrain: V1TimeGrain | undefined,
  usingRillTime: boolean,
): RangeBuckets {
  const showDefaults = !yamlRanges.length;

  if (!minTimeGrain) {
    minTimeGrain = V1TimeGrain.TIME_GRAIN_SECOND;
  }

  if (showDefaults) {
    if (!usingRillTime) return defaultBuckets;

    const timeGrainOptions = getAllowedGrains(minTimeGrain);

    return getDefaultRangeBuckets(timeGrainOptions);
  }

  const skeleton: RangeBuckets = {
    previous: [],
    latest: [],
    periodToDate: [],
    custom: [],
    allTime: false,
  };

  yamlRanges.forEach(({ range }) => {
    if (!range) return;

    if (range === "inf") {
      skeleton.allTime = true;
      return;
    }

    try {
      const parsed = parseRillTime(range);

      const { interval } = parsed;

      if (
        interval instanceof RillLegacyIsoInterval ||
        interval instanceof RillShorthandInterval
      ) {
        skeleton.latest.push(parsed);
      } else if (interval instanceof RillTimeStartEndInterval) {
        if (previousPeriodRegex.test(range)) {
          skeleton.previous.push(parsed);
        } else {
          skeleton.custom.push(parsed);
        }
      } else if (interval instanceof RillPeriodToGrainInterval) {
        skeleton.periodToDate.push(parsed);
      } else if (interval instanceof RillLegacyDaxInterval) {
        if (isRillPreviousPeriod(range)) {
          skeleton.previous.push(parsed);
        } else if (isRillPeriodToDate(range)) {
          skeleton.periodToDate.push(parsed);
        } else {
          skeleton.custom.push(parsed);
        }
      } else {
        skeleton.custom.push(parsed);
      }
    } catch (e) {
      console.error("Error parsing RillTime", e);
    }
  });

  return skeleton;
}

function convertIsoToRillTime(iso: string): string {
  const upper = iso.toUpperCase();

  if (!upper.startsWith("P")) {
    throw new Error("Invalid ISO duration: must start with P");
  }

  const result: string[] = [];

  const [datePartRaw, timePartRaw] = upper.slice(1).split("T");
  const datePart = datePartRaw || "";
  const timePart = timePartRaw || "";

  const dateUnits: Record<string, string> = {
    Y: "Y",
    M: "M",
    W: "W",
    D: "D",
  };

  const timeUnits: Record<string, string> = {
    H: "H",
    M: "m",
    S: "S",
  };

  for (const [unit, rill] of Object.entries(dateUnits)) {
    const match = datePart.match(new RegExp(`(\\d+(\\.\\d+)?)${unit}`));
    if (match) result.push(`${match[1]}${rill}`);
  }

  for (const [unit, rill] of Object.entries(timeUnits)) {
    const match = timePart.match(new RegExp(`(\\d+(\\.\\d+)?)${unit}`));
    if (match) result.push(`${match[1]}${rill}`);
  }

  return result.join("");
}

const previousCompleteMap = {
  PHC: "-1H/H to ref/H",
  PDC: "-1D/D to ref/D",
  PWC: "-1W/W to ref/W",
  PMC: "-1M/M to ref/M",
  PQC: "-1Q/Q to ref/Q",
  PYC: "-1Y/Y to ref/Y",
};

export function convertLegacyTime(timeString: string) {
  if (timeString.startsWith("rill-")) {
    const stripped = timeString.replace("rill-", "");
    if (timeString === "rill-TD") return "DTD";
    if (previousCompleteMap[stripped]) return previousCompleteMap[stripped];
    return timeString.replace("rill-", "");
  } else if (timeString.startsWith("P") || timeString.startsWith("p")) {
    return convertIsoToRillTime(timeString);
  }
  return timeString;
}

export function constructAsOfString(
  asOf: RillTimeLabel | undefined,
  grain: V1TimeGrain | undefined | null,
  pad: boolean,
): string {
  if (!grain) {
    return asOf ?? RillTimeLabel.Now;
  }

  const alias = V1TimeGrainToAlias[grain];

  let base: string;

  if (asOf === RillTimeLabel.Latest || asOf === undefined) {
    base = `latest/${alias}`;
  } else if (asOf === RillTimeLabel.Watermark) {
    base = `watermark/${alias}`;
  } else if (asOf === RillTimeLabel.Now) {
    base = `now/${alias}`;
  } else {
    base = `${asOf}/${alias}`;
  }

  if (pad) {
    return `${base}+1${alias}`;
  } else {
    return base;
  }
}

export function isUsingLegacyTime(timeString: string | undefined): boolean {
  return (
    timeString?.startsWith("rill") ||
    timeString?.startsWith("P") ||
    timeString?.startsWith("p") ||
    false
  );
}

export function constructNewString({
  currentString,
  truncationGrain,
  snapToEnd,
  ref,
}: {
  currentString: string;
  truncationGrain: V1TimeGrain | undefined | null;
  snapToEnd: boolean;
  ref: RillTimeLabel | undefined;
}): string {
  const legacy = isUsingLegacyTime(currentString);

  const rillTime = parseRillTime(
    legacy ? convertLegacyTime(currentString) : currentString,
  );

  const newAsOfString = constructAsOfString(ref, truncationGrain, snapToEnd);

  overrideRillTimeRef(rillTime, newAsOfString);

  return rillTime.toString();
}
