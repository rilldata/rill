import { fromTimeRangesParams } from "@rilldata/web-common/features/dashboards/url-state/convertURLToExplorePreset";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";
import { derived, get, writable, type Readable } from "svelte/store";
import {
  ALL_TIME_RANGE_ALIAS,
  deriveInterval,
} from "../../dashboards/time-controls/new-time-controls";
import type { CanvasEntity, SearchParamsStore } from "./canvas-entity";
import { parseRillTime } from "../../dashboards/url-state/time-ranges/parser";
import {
  RillLegacyDaxInterval,
  RillPeriodToGrainInterval,
  RillTime,
} from "../../dashboards/url-state/time-ranges/RillTime";
import {
  DateTimeUnitToV1TimeGrain,
  minTimeGrainToDefaultTimeRange,
  V1TimeGrainToDateTimeUnit,
} from "@rilldata/web-common/lib/time/new-grains";
import { maybeWritable } from "@rilldata/web-common/lib/store-utils";
import type { TimeManager } from "./time-manager";
import { getComparisonInterval } from "@rilldata/web-common/lib/time/comparisons";
import { getValidatedTimeGrain } from "@rilldata/web-common/lib/time/grains";

export type MinMax = {
  min: DateTime<true>;
  max: DateTime<true>;
};

export class TimeState {
  // This class is used for global and widget time controls on Canvas
  // To avoid methods running in widgets, use isWidgetInstance flag
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

  urlInitialized = writable<boolean>(false);

  minMaxTimeStamps = maybeWritable<MinMax>();

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
        this.urlInitialized,
        this.manager.hasTimeSeriesMap,
      ],
      ([
        urlRange,
        defaultRange,
        minTimeGrain,
        timeRangeOptions,
        urlInitialized,
        hasTimeSeriesMap,
      ]) => {
        const hasTimeSeries = this.metricsViewName
          ? hasTimeSeriesMap.get(this.metricsViewName)
          : Array.from(hasTimeSeriesMap.values()).some((v) => v);

        if (!hasTimeSeries || !urlInitialized || !this.manager.specInitialized)
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
        typeof this.manager.hasTimeSeriesMap,
      ],
      Interval<true> | undefined
    >(
      [
        this.rangeStore,
        this.parent._metricsViews,
        this.timeZoneStore,
        this.manager.hasTimeSeriesMap,
      ],
      ([range, allMetricsViews, timeZone, hasTimeSeriesMap], set) => {
        const allMetricsViewNames = this.metricsViewName
          ? [this.metricsViewName]
          : Object.keys(allMetricsViews || {});

        const metricsViewsWithTimeSeries = allMetricsViewNames.filter(
          (mvName) => {
            return hasTimeSeriesMap.get(mvName);
          },
        );

        if (!metricsViewsWithTimeSeries.length || !range || !timeZone) {
          set(undefined);
          return;
        }

        const promises = metricsViewsWithTimeSeries.map((mvName) => {
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

            if (minDate && maxDate) {
              this.minMaxTimeStamps.set({
                min: minDate,
                max: maxDate,
              });
            }
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
      [this.interval, this.minMaxTimeStamps],
      ([interval, minMaxTimeStamps]) => {
        if (!interval || !interval.start || !interval.end || !minMaxTimeStamps)
          return { left: false, right: false };

        const canPanLeft = interval?.start > minMaxTimeStamps.min;
        const canPanRight = interval?.end < minMaxTimeStamps.max;

        return { left: canPanLeft, right: canPanRight };
      },
    );

    this.grainStore = derived(
      [
        this.urlGrainStore,
        this.manager.largestMinTimeGrain,
        this.parsedRange,
        this.interval,
      ],
      ([urlGrain, minTimeGrain, parsedRange, interval]) => {
        return getValidatedTimeGrain(
          interval,
          minTimeGrain,
          urlGrain,
          parsedRange,
        );
      },
    );
  }

  setMinMaxTimeStamps = (minMax: MinMax) => {
    this.minMaxTimeStamps.set(minMax);
  };

  onUrlChange = (searchParams: URLSearchParams) => {
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

    this.urlInitialized.set(true);
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
      const mappedTimeGrain = V1TimeGrainToDateTimeUnit[timeGrain];
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
