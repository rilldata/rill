// WIP as of 04/19/2024

import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { writable, type Writable, get } from "svelte/store";
import {
  Interval,
  DateTime,
  DurationObjectUnits,
  DateTimeUnit,
  Duration,
  IANAZone,
} from "luxon";
import { MetricsViewSpecAvailableTimeRange } from "@rilldata/web-common/runtime-client";

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
  "rill-PWC": "Previous week complete",
  "rill-PMC": "Previous month complete",
  "rill-PQC": "Previous quarter complete",
  "rill-PYC": "Previous year complete",
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
  "P3M",
  "P12M",
] as const;

// TYPES -> time-control-types.ts

type RillPeriodToDateTuple = typeof RILL_PERIOD_TO_DATE;
export type RillPeriodToDate = RillPeriodToDateTuple[number];

type RillPreviousPeriodTuple = typeof RILL_PREVIOUS_PERIOD;
export type RillPreviousPeriod = RillPreviousPeriodTuple[number];

type RillLatestTuple = typeof RILL_LATEST;
export type RillLatest = RillLatestTuple[number];

const CUSTOM_TIME_RANGE_ALIAS = "CUSTOM" as const;
const ALL_TIME_RANGE_ALIAS = "inf" as const;
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

  constructor(start?: DateTime, end?: DateTime) {
    if (start && end) this._interval.update((i) => i.set({ start, end }));
  }

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

  updateZone(zone: IANAZone) {
    this._interval.update((i) => i.mapEndpoints((e) => e.setZone(zone)));
  }
}

class MetricsTimeControls {
  private _zone: Writable<IANAZone> = writable(new IANAZone("UTC"));
  private _selected = writable<NamedRange | ISODurationString>(
    ALL_TIME_RANGE_ALIAS,
  );
  private _subrange: IntervalStore = new IntervalStore();
  private _visibleRange: IntervalStore = new IntervalStore();
  private _comparisonRange: IntervalStore = new IntervalStore();
  private _showComparison: Writable<boolean> = writable(false);
  private _maxRange: Interval = Interval.fromDateTimes(
    DateTime.fromMillis(0),
    DateTime.utc(),
  );

  constructor(maxStart: DateTime, maxEnd: DateTime) {
    this._maxRange = Interval.fromDateTimes(maxStart, maxEnd);
    this._visibleRange.updateInterval(this._maxRange);
  }

  private applySubrange = () => {
    this._visibleRange.updateInterval(get(this._subrange));
    this._selected.set(CUSTOM_TIME_RANGE_ALIAS);
    this._subrange.clear();
  };

  private applyISODuration = (iso: ISODurationString) => {
    if (this._maxRange?.end) {
      const interval = deriveInterval(iso, this._maxRange.end);
      if (interval?.isValid) {
        this._visibleRange.updateInterval(interval);
        this._selected.set(iso);
      }
    }
  };

  private applyCustomRange = (start: DateTime, end: DateTime) => {
    this._visibleRange.updateInterval(Interval.fromDateTimes(start, end));
    this._selected.set(CUSTOM_TIME_RANGE_ALIAS);
  };

  private applyNamedRange = (name: NamedRange) => {
    if (this._maxRange?.end) {
      const interval = deriveInterval(name, this._maxRange.end);
      if (interval?.isValid) {
        this._visibleRange.updateInterval(interval);
        this._selected.set(name);
      }
    }
  };

  private applyUnknown = (string: string | undefined) => {
    if (!string) return;

    if (isValidISODuration(string)) {
      this.applyISODuration(string);
    } else {
      this.applyNamedRange(string as NamedRange);
    }
  };

  private applyAllTime = () => {
    if (this._maxRange.start && this._maxRange.end) {
      this._visibleRange.updateInterval(
        Interval.fromDateTimes(this._maxRange.start, this._maxRange.end),
      );
      this._selected.set(ALL_TIME_RANGE_ALIAS);
    }
  };

  updateZone = (zone: IANAZone) => {
    this._zone.set(zone);
    // Actual zone change logic is an active discussion with the team
    // But what this does is just say give me this relative interval in a different zone
    // Currently, we shift the underlying interval to the new zone, which seems wrong
    this._visibleRange.updateZone(zone);
  };

  apply = {
    subrange: this.applySubrange,
    isoDuration: this.applyISODuration,
    customRange: this.applyCustomRange,
    namedRange: this.applyNamedRange,
    unknown: this.applyUnknown,
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
      store = new MetricsTimeControls(maxStart, maxEnd);
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

export function deriveInterval(
  name: NamedRange | ISODurationString,
  anchor: DateTime,
) {
  if (isRillPeriodToDate(name)) {
    const period = RILL_TO_UNIT[name];
    return getPeriodToDate(anchor, period);
  }

  if (isRillPreviousPeriod(name)) {
    const period = RILL_TO_UNIT[name];
    return getPreviousPeriodComplete(anchor, period, 1);
  }

  const duration = isValidISODuration(name);

  if (duration) return getInterval(duration, anchor);
}

export function getPeriodToDate(date: DateTime, period: DateTimeUnit) {
  const periodStart = date.startOf(period);
  const exclusiveEnd = date.endOf("day").plus({ millisecond: 1 });

  return Interval.fromDateTimes(periodStart, exclusiveEnd);
}

export function getPreviousPeriodComplete(
  anchor: DateTime,
  period: DateTimeUnit,
  steps = 0,
) {
  const startOfCurrentPeriod = anchor.startOf(period);
  const shiftedStart = startOfCurrentPeriod.minus({ [period + "s"]: steps });
  const exclusiveEnd = shiftedStart.endOf(period).plus({ millisecond: 1 });

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
      ? endDate.endOf(smallestUnit).plus({ millisecond: 1 })
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

// BUCKETS FOR DISPLAYING IN DROPDOWN (yaml spec may make this unnecessary)

type RangeBuckets = {
  latest: string[];
  previous: RillPreviousPeriod[];
  periodToDate: RillPeriodToDate[];
};

const defaultBuckets = {
  previous: [...RILL_PREVIOUS_PERIOD],
  latest: [...RILL_LATEST],
  periodToDate: [...RILL_PERIOD_TO_DATE],
};

export function bucketYamlRanges(
  availableRanges: MetricsViewSpecAvailableTimeRange[],
): RangeBuckets {
  const showDefaults = !availableRanges.length;

  if (showDefaults) {
    return defaultBuckets;
  }

  return availableRanges.reduce(
    (record, { range }) => {
      if (!range) return record;

      if (isRillPeriodToDate(range)) {
        record.periodToDate.push(range);
      } else if (isRillPreviousPeriod(range)) {
        record.previous.push(range);
      } else {
        record.latest.push(range);
      }

      return record;
    },
    <RangeBuckets>{
      previous: [],
      latest: [],
      periodToDate: [],
    },
  );
}
