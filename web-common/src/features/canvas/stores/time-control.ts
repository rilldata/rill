import { fromTimeRangesParams } from "@rilldata/web-common/features/dashboards/url-state/convertURLToExplorePreset";
import { ToURLParamTimeGrainMapMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import {
  V1ExploreComparisonMode,
  V1TimeGrain,
  type V1ExploreTimeRange,
} from "@rilldata/web-common/runtime-client";
import { DateTime, Interval, Settings } from "luxon";
import {
  derived,
  get,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import {
  ALL_TIME_RANGE_ALIAS,
  deriveInterval,
  normalizeWeekday,
} from "../../dashboards/time-controls/new-time-controls";
import { type CanvasResponse } from "../selector";
import type { CanvasEntity, SearchParamsStore } from "./canvas-entity";
import { parseRillTime } from "../../dashboards/url-state/time-ranges/parser";
import {
  RillLegacyDaxInterval,
  RillPeriodToGrainInterval,
  RillTime,
} from "../../dashboards/url-state/time-ranges/RillTime";
import {
  DateTimeUnitToV1TimeGrain,
  getRangePrecision,
  isGrainAllowed,
  minTimeGrainToDefaultTimeRange,
} from "@rilldata/web-common/lib/time/new-grains";

export type MinMax = {
  min: DateTime<true>;
  max: DateTime<true>;
};

function maybeWritable<T>(value?: T): Writable<T | undefined> {
  return writable(value);
}

export class TimeManager {
  largestMinTimeGrain: Readable<V1TimeGrain>;
  minMaxTimeStamps = maybeWritable<MinMax>();
  defaultTimeRangeStore = maybeWritable<string>();
  defaultComparisonRangeStore = maybeWritable<string>();
  timeRangeOptionsStore = writable<V1ExploreTimeRange[]>([]);
  availableTimeZonesStore = writable<string[]>([]);
  allowCustomRangeStore = writable<boolean>(true);
  hasTimeSeriesMap = writable<Map<string, boolean>>(new Map());
  minTimeGrainMap = writable<Map<string, V1TimeGrain>>(new Map());
  hasTimeSeries: Readable<boolean>;
  specInitialized = false;
  global: TimeControls;

  constructor(
    public searchParamsStore: SearchParamsStore,
    public parent: CanvasEntity,
  ) {
    this.largestMinTimeGrain = derived(this.minTimeGrainMap, (map) => {
      let largest: V1TimeGrain = V1TimeGrain.TIME_GRAIN_UNSPECIFIED;

      for (const grain of map.values()) {
        if (isGrainBigger(grain, largest)) {
          largest = grain;
        }
      }
      return largest;
    });

    this.hasTimeSeries = derived(
      [this.hasTimeSeriesMap],
      ([$hasTimeSeriesMap]) => {
        const values = Array.from($hasTimeSeriesMap.values());
        return values.some((v) => v);
      },
    );

    this.global = new TimeControls(this.searchParamsStore, this.parent, this);
  }

  createLocalTimeControls = (
    componentName: string,
    searchParamsStore: SearchParamsStore,
  ) => {
    return new TimeControls(
      searchParamsStore,
      this.parent,
      this,
      componentName,
    );
  };

  onSpecChange = (spec: CanvasResponse) => {
    this.specInitialized = true;
    const defaultPreset = spec?.canvas?.defaultPreset;

    const ranges = this.checkAndSetTimeRangeOptions(spec);
    this.checkAndSetDefaultTimeRange(spec, ranges);
    this.checkAndSetAvailableTimeZones(spec);
    this.checkIfHasTimeSeries(spec);
    this.checkAndSetFirstDayOfWeek(spec);

    if (
      defaultPreset?.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
    ) {
      this.defaultComparisonRangeStore.set("rill-PP");
    } else {
      this.defaultComparisonRangeStore.set(undefined);
    }
  };

  checkAndSetFirstDayOfWeek(response: CanvasResponse) {
    // TODO: figure out a better way of handling this property
    // when it's not consistent across all metrics views - bgh
    const firstMetricsViewName = Object.keys(response.metricsViews)?.[0];
    const firstDayOfWeekOfFirstMetricsView =
      response.metricsViews[firstMetricsViewName]?.state?.validSpec
        ?.firstDayOfWeek;

    if (!firstMetricsViewName) return;

    Settings.defaultWeekSettings = {
      firstDay: normalizeWeekday(firstDayOfWeekOfFirstMetricsView),
      weekend: [6, 7],
      minimalDays: 4,
    };
  }

  checkIfHasTimeSeries(response: CanvasResponse) {
    const metricsViews = response.metricsViews || {};

    const timeGrainMap = new Map<string, V1TimeGrain>();
    const hasTimeSeriesMap = new Map<string, boolean>();

    Object.entries(metricsViews).forEach(([mvName, mv]) => {
      const metricsViewSpec = mv?.state?.validSpec;
      const timeGrain = metricsViewSpec?.smallestTimeGrain;
      timeGrainMap.set(mvName, timeGrain || V1TimeGrain.TIME_GRAIN_UNSPECIFIED);
      const hasTimeDimension = Boolean(metricsViewSpec?.timeDimension);
      hasTimeSeriesMap.set(mvName, hasTimeDimension);
    });

    this.minTimeGrainMap.set(timeGrainMap);
    this.hasTimeSeriesMap.set(hasTimeSeriesMap);
  }

  checkAndSetAllowCustomRange = (response: CanvasResponse) => {
    const allowCustomRange = response.canvas?.allowCustomTimeRange ?? true;

    this.allowCustomRangeStore.set(allowCustomRange);

    return allowCustomRange;
  };

  checkAndSetAvailableTimeZones(response: CanvasResponse) {
    const timeZones = response.canvas?.timeZones ?? [];

    this.availableTimeZonesStore.set(timeZones);

    return timeZones;
  }

  checkAndSetDefaultTimeRange(
    response: CanvasResponse,

    timeRanges: V1ExploreTimeRange[],
  ) {
    const defaultTimeRange = response.canvas?.defaultPreset?.timeRange;

    const finalRange = defaultTimeRange ?? timeRanges[0]?.range;

    const currentValue = get(this.defaultTimeRangeStore);
    if (currentValue !== finalRange && finalRange) {
      this.defaultTimeRangeStore.set(finalRange);
    }
  }

  checkAndSetTimeRangeOptions(response: CanvasResponse) {
    const timeRangeOptions = response.canvas?.timeRanges ?? [];

    this.timeRangeOptionsStore.set(timeRangeOptions);

    return timeRangeOptions;
  }
}

// This class is used for global and widget time controls on Canvas
// To avoid methods running in widgets, use isWidgetInstance flag
export class TimeControls {
  isWidgetInstance: boolean;

  interval: Readable<Interval<true> | undefined>;

  rangeStore: Readable<string | undefined>;
  grainStore: Readable<V1TimeGrain | undefined>;
  timeZoneStore: Readable<string>;

  private parsedRange: Readable<RillTime | undefined>;

  comparisonRangeStore = writable<string>("rill-PP");
  comparisonIntervalStore: Readable<Interval<true> | undefined>;
  showTimeComparisonStore = writable<boolean>(false);

  urlRangeStore = maybeWritable<string>();
  urlGrainStore = maybeWritable<V1TimeGrain>();
  urlTimeZoneStore = maybeWritable<string>();
  urlComparisonRangeStore = maybeWritable<string>();

  canPanStore: Readable<{ left: boolean; right: boolean }>;

  urlInitialized = false;

  constructor(
    public searchParamsStore: SearchParamsStore,
    public parent: CanvasEntity,
    public manager: TimeManager,
    public componentName?: string,
    public metricsViewName?: string,
  ) {
    this.isWidgetInstance = Boolean(componentName);

    this.rangeStore = derived(
      [
        this.urlRangeStore,
        this.manager.defaultTimeRangeStore,
        this.manager.largestMinTimeGrain,
        this.manager.timeRangeOptionsStore,
        this.manager.minTimeGrainMap,
      ],
      ([urlRange, defaultRange, minTimeGrain, timeRangeOptions]) => {
        if (!this.urlInitialized || !this.manager.specInitialized)
          return undefined;

        if (urlRange) return urlRange;

        if (this.isWidgetInstance) return undefined;

        if (defaultRange) return defaultRange;

        if (timeRangeOptions.length > 0) return timeRangeOptions[0].range;

        const derivedDefault = minTimeGrainToDefaultTimeRange[minTimeGrain];

        if (derivedDefault) return derivedDefault;

        return ALL_TIME_RANGE_ALIAS;
      },
    );

    this.parsedRange = derived(this.rangeStore, (range) => {
      if (!range) return undefined;

      try {
        const parsed = parseRillTime(range);
        return parsed;
      } catch {
        return undefined;
      }
    });

    this.timeZoneStore = derived([this.urlTimeZoneStore], ([urlTimeZone]) => {
      if (urlTimeZone) return urlTimeZone;

      return "UTC";
    });

    this.interval = derived<
      [
        typeof this.rangeStore,
        typeof this.parent._metricsViews,
        typeof this.timeZoneStore,
      ],
      Interval<true> | undefined
    >(
      [this.rangeStore, this.parent._metricsViews, this.timeZoneStore],
      ([range, allMetricsViews, timeZone], set) => {
        const allMetricsViewNames = Object.keys(allMetricsViews || {});

        if (!allMetricsViewNames.length || !range || !timeZone) {
          set(undefined);
          return;
        }

        const promises = allMetricsViewNames.map((mvName) => {
          return deriveInterval(range, mvName, timeZone);
        });

        Promise.all(promises)
          .then((results) => {
            let latestInterval: Interval<true> | undefined = undefined;
            let minDate: DateTime<true> | undefined = undefined;
            let maxDate: DateTime<true> | undefined = undefined;

            results.forEach((res) => {
              const interval = res.interval;
              const fullTimeRange = res.fullTimeRange;

              if (
                !interval?.end ||
                !interval?.start ||
                !fullTimeRange?.max ||
                !fullTimeRange?.min
              ) {
                return;
              }

              const min = DateTime.fromISO(fullTimeRange.min).setZone(timeZone);
              const max = DateTime.fromISO(fullTimeRange.max).setZone(timeZone);

              if (min.isValid && (!minDate || min < minDate)) {
                minDate = min;
              }
              if (max.isValid && (!maxDate || max > maxDate)) {
                maxDate = max;
              }

              if (
                interval?.isValid &&
                (!latestInterval || latestInterval.end < interval.end)
              ) {
                latestInterval = interval;
              }
            });

            set(latestInterval);
            this.manager.minMaxTimeStamps.set(
              minDate && maxDate ? { min: minDate, max: maxDate } : undefined,
            );
          })
          .catch(console.error);
      },
    );

    this.comparisonIntervalStore = derived(
      [
        this.interval,
        this.comparisonRangeStore,
        this.rangeStore,
        this.timeZoneStore,
      ],
      ([interval, comparisonRange, range, zone]) => {
        if (range === ALL_TIME_RANGE_ALIAS) return undefined;

        return getComparisonInterval(interval, comparisonRange, zone);
      },
    );

    this.canPanStore = derived(
      [this.interval, this.manager.minMaxTimeStamps],
      ([interval, minMaxTimeStamps]) => {
        if (!interval || !interval.start || !interval.end || !minMaxTimeStamps)
          return { left: false, right: false };

        const canPanLeft = interval?.start > minMaxTimeStamps.min;
        const canPanRight = interval?.end < minMaxTimeStamps.max;

        return { left: canPanLeft, right: canPanRight };
      },
    );

    this.grainStore = derived(
      [this.urlGrainStore, this.manager.largestMinTimeGrain, this.parsedRange],
      ([urlGrain, minTimeGrain, parsedRange]) => {
        if (urlGrain && isGrainAllowed(urlGrain, minTimeGrain)) {
          return urlGrain;
        } else if (parsedRange) {
          const parsedRangePrecision = getRangePrecision(parsedRange);

          if (isGrainAllowed(parsedRangePrecision, minTimeGrain)) {
            return parsedRangePrecision;
          } else {
            return minTimeGrain;
          }
        } else {
          return minTimeGrain;
        }
      },
    );
  }

  onUrlChange = (searchParams: URLSearchParams) => {
    this.urlInitialized = true;

    const { range, comparisonRange, zone, grain } =
      parseSearchParams(searchParams);

    // Component without local time range
    if (this.isWidgetInstance && !range) return;

    this.urlRangeStore.set(range);
    this.urlGrainStore.set(grain);
    this.urlTimeZoneStore.set(zone);
    this.urlComparisonRangeStore.set(comparisonRange);

    this.showTimeComparisonStore.set(!!comparisonRange);

    if (comparisonRange) {
      this.comparisonRangeStore.set(comparisonRange);
    }
  };

  set = {
    zone: (timeZone: string, checkIfSet = false, replaceState = false) => {
      const props = new Map([[ExploreStateURLParams.TimeZone, timeZone]]);
      return this.searchParamsStore.set(props, checkIfSet, replaceState);
    },
    range: (
      range: string,
      checkIfSet = false,
      replaceState = false,
      setComparisonToContinuous = false,
    ) => {
      const props = new Map([[ExploreStateURLParams.TimeRange, range]]);

      if (setComparisonToContinuous) {
        props.set(ExploreStateURLParams.ComparisonTimeRange, "rill-PP");
      }
      return this.searchParamsStore.set(props, checkIfSet, replaceState);
    },
    grain: (
      timeGrain: V1TimeGrain,
      checkIfSet = false,
      replaceState = false,
    ) => {
      const mappedTimeGrain = ToURLParamTimeGrainMapMap[timeGrain];
      const props = new Map<string, string>();
      if (mappedTimeGrain) {
        props.set(ExploreStateURLParams.TimeGrain, mappedTimeGrain);
        return this.searchParamsStore.set(props, checkIfSet, replaceState);
      }
    },
    comparison: (
      range: boolean | string,
      checkIfSet = false,
      replaceState = false,
    ) => {
      const showTimeComparison = Boolean(range);
      if (showTimeComparison) {
        if (range === true) {
          const selectedComparisonTimeRange = get(this.comparisonRangeStore);
          if (!selectedComparisonTimeRange) return;
          const props = new Map([
            [
              ExploreStateURLParams.ComparisonTimeRange,
              selectedComparisonTimeRange,
            ],
          ]);
          return this.searchParamsStore.set(props, checkIfSet, replaceState);
        } else if (typeof range === "string") {
          const props = new Map([
            [ExploreStateURLParams.ComparisonTimeRange, range],
          ]);
          return this.searchParamsStore.set(props, checkIfSet, replaceState);
        }
      } else {
        const props = new Map([
          [ExploreStateURLParams.ComparisonTimeRange, undefined],
        ]);
        return this.searchParamsStore.set(props, checkIfSet, replaceState);
      }
    },
  };

  clearAll = () => {
    const props = new Map([
      [ExploreStateURLParams.TimeRange, undefined],
      [ExploreStateURLParams.TimeGrain, undefined],
      [ExploreStateURLParams.ComparisonTimeRange, undefined],
      [ExploreStateURLParams.TimeZone, undefined],
    ]);

    return this.searchParamsStore.set(props);
  };
}

export function parseSearchParams(urlParams: URLSearchParams) {
  const { preset } = fromTimeRangesParams(urlParams, new Map());

  let range: string | undefined;
  let comparisonRange: string | undefined;
  let zone: string | undefined;
  let grain: V1TimeGrain | undefined = undefined;

  if (preset.timeRange) {
    range = preset.timeRange;
  }

  if (preset.timeGrain) {
    grain = DateTimeUnitToV1TimeGrain[preset.timeGrain];
  }

  if (preset.compareTimeRange) {
    comparisonRange = preset.compareTimeRange;
  }

  if (preset.timezone) {
    zone = preset.timezone;
  }

  return {
    range,
    zone,
    grain,
    comparisonRange,
  };
}

export function getComparisonTypeFromRangeString(
  range: string | undefined,
): TimeComparisonOption {
  if (!range) {
    return TimeComparisonOption.CONTIGUOUS;
  }
  try {
    const { interval, rangeGrain } = parseRillTime(range);

    if (
      interval instanceof RillLegacyDaxInterval ||
      interval instanceof RillPeriodToGrainInterval
    ) {
      return rangeGrain && rangeGrain in timeGrainToComparisonOptionMap
        ? timeGrainToComparisonOptionMap[rangeGrain]
        : TimeComparisonOption.CONTIGUOUS;
    } else {
      return TimeComparisonOption.CONTIGUOUS;
    }
  } catch {
    return TimeComparisonOption.CONTIGUOUS;
  }
}

const timeGrainToComparisonOptionMap: Record<
  V1TimeGrain,
  TimeComparisonOption
> = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: TimeComparisonOption.CONTIGUOUS,
  [V1TimeGrain.TIME_GRAIN_SECOND]: TimeComparisonOption.CONTIGUOUS,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: TimeComparisonOption.CONTIGUOUS,
  [V1TimeGrain.TIME_GRAIN_HOUR]: TimeComparisonOption.CONTIGUOUS,
  [V1TimeGrain.TIME_GRAIN_DAY]: TimeComparisonOption.DAY,
  [V1TimeGrain.TIME_GRAIN_WEEK]: TimeComparisonOption.WEEK,
  [V1TimeGrain.TIME_GRAIN_MONTH]: TimeComparisonOption.MONTH,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: TimeComparisonOption.QUARTER,
  [V1TimeGrain.TIME_GRAIN_YEAR]: TimeComparisonOption.YEAR,
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: TimeComparisonOption.CONTIGUOUS,
};

function getComparisonInterval(
  interval: Interval<true> | undefined,
  comparisonRange: string | undefined,
  activeTimeZone: string,
): Interval<true> | undefined {
  if (!interval || !comparisonRange) return undefined;

  let comparisonInterval: Interval | undefined = undefined;

  const COMPARISON_DURATIONS = {
    "rill-PP": interval.toDuration(),
    "rill-PD": { days: 1 },
    "rill-PW": { weeks: 1 },
    "rill-PM": { months: 1 },
    "rill-PQ": { quarter: 1 },
    "rill-PY": { years: 1 },
  };

  const duration =
    COMPARISON_DURATIONS[comparisonRange as keyof typeof COMPARISON_DURATIONS];

  if (duration) {
    comparisonInterval = Interval.fromDateTimes(
      interval.start.minus(duration),
      interval.end.minus(duration),
    );
  } else {
    const normalizedRange = comparisonRange.replace(",", "/");
    comparisonInterval = Interval.fromISO(normalizedRange).mapEndpoints((dt) =>
      dt.setZone(activeTimeZone),
    );
  }

  return comparisonInterval.isValid ? comparisonInterval : undefined;
}
