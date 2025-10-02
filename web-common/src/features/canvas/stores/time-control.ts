import { getComponentMetricsViewFromSpec } from "@rilldata/web-common/features/canvas/components/util";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { fromTimeRangesParams } from "@rilldata/web-common/features/dashboards/url-state/convertURLToExplorePreset";
import { ToURLParamTimeGrainMapMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
import {
  createQueryServiceMetricsViewTimeRange,
  V1ExploreComparisonMode,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import {
  runtime,
  type Runtime,
} from "@rilldata/web-common/runtime-client/runtime-store";
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
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import { getDefaultTimeRange } from "../../dashboards/stores/get-rill-default-explore-state";
import {
  DateTimeUnitToV1TimeGrain,
  getSmallestGrain,
  GrainAliasToV1TimeGrain,
  isGrainAllowed,
  V1TimeGrainToDateTimeUnit,
  type TimeGrainAlias,
} from "@rilldata/web-common/lib/time/new-grains";
import { type RillTime } from "../../dashboards/url-state/time-ranges/RillTime";
import { parseRillTime } from "../../dashboards/url-state/time-ranges/parser";

export type MinMax = {
  min: DateTime<true>;
  max: DateTime<true>;
};

function maybeWritable<T>(value?: T): Writable<T | undefined> {
  return writable(value);
}

export class TimeControls {
  _urlRange = maybeWritable<string>();
  _urlGrain = maybeWritable<V1TimeGrain>();

  _defaultRange = maybeWritable<string>();

  _range: Readable<string | undefined>;
  _grain: Readable<V1TimeGrain | undefined>;
  _zone: Writable<string> = writable("UTC");
  _interval: Readable<Interval<true> | undefined>;
  _isIntervalAllowed: Readable<boolean>;

  _showTimeComparison = writable(false);
  _comparisonRange = maybeWritable<string>();
  _comparisonInterval: Readable<Interval<true> | undefined>;

  minMaxTimeStamps: Readable<MinMax | undefined>;
  isReady: Readable<boolean>;
  hasTimeSeries = writable(false);
  minTimeGrain = maybeWritable<V1TimeGrain>();

  timeRangeOptions = writable<string[]>([]);
  latestMetricsView = maybeWritable<string>();

  firstPass = true;

  private parsedRange: Readable<RillTime | undefined>;
  private parsedRangePrecision: Readable<V1TimeGrain | undefined>;

  constructor(
    private specStore: CanvasSpecResponseStore,
    public searchParamsCallback: (
      key: string,
      value: string | undefined,
      checkIfSet?: boolean,
    ) => boolean,
    public componentName?: string,
    public canvasName?: string,
  ) {
    this.minMaxTimeStamps = this.deriveMinMaxTimestamps(
      runtime,
      this.specStore,
    );

    this._range = derived(
      [
        this._urlRange,
        this.minTimeGrain,
        this.minMaxTimeStamps,
        this._defaultRange,
        this.timeRangeOptions,
      ],
      ([
        urlRange,
        minTimeGrain,
        minMaxTimeStamps,
        defaultRange,
        timeRangeOptions,
      ]) => {
        if (urlRange) return urlRange;

        if (defaultRange) return defaultRange;

        if (timeRangeOptions.length > 0) return timeRangeOptions[0];

        if (!minMaxTimeStamps) return undefined;

        const derivedDefault = getDefaultTimeRange(minTimeGrain, {
          max: minMaxTimeStamps.max.toISO(),
          min: minMaxTimeStamps.min.toISO(),
        });

        if (derivedDefault) return derivedDefault;

        return ALL_TIME_RANGE_ALIAS;
      },
    );

    this.parsedRange = derived(this._range, (range) => {
      if (!range) return undefined;

      try {
        const parsed = parseRillTime(range);
        return parsed;
      } catch {
        return undefined;
      }
    });

    this.parsedRangePrecision = derived(this.parsedRange, (parsedRange) => {
      if (!parsedRange) return undefined;

      const asOfSnap = parsedRange.asOfLabel?.snap;

      const asOfSnapV1Grain =
        GrainAliasToV1TimeGrain[asOfSnap as TimeGrainAlias];
      const rangeV1Grain = parsedRange.rangeGrain;
      const intervalV1Grain = parsedRange.interval.getGrain();

      return getSmallestGrain([asOfSnapV1Grain, rangeV1Grain, intervalV1Grain]);
    });

    this._grain = derived(
      [this._urlGrain, this.minTimeGrain, this.parsedRangePrecision],
      ([urlGrain, minTimeGrain, parsedRangePrecision]) => {
        if (urlGrain && isGrainAllowed(urlGrain, minTimeGrain)) {
          return urlGrain;
        } else if (
          parsedRangePrecision &&
          isGrainAllowed(parsedRangePrecision, minTimeGrain)
        ) {
          return parsedRangePrecision;
        } else {
          return minTimeGrain;
        }
      },
    );

    this._interval = derived<
      [
        typeof this._range,
        typeof this.latestMetricsView,
        typeof this._zone,
        typeof this.minTimeGrain,
      ],
      Interval<true> | undefined
    >(
      [this._range, this.latestMetricsView, this._zone, this.minTimeGrain],
      ([range, latestMetricsView, zone], set) => {
        if (!latestMetricsView || !range || !zone) {
          set(undefined);
          return;
        }

        deriveInterval(range, latestMetricsView, zone)
          .then((result) => {
            if (result.interval.isValid) {
              set(result.interval as Interval<true>);
            } else {
              set(undefined);
            }
          })
          .catch(() => {
            set(undefined);
          });
      },
      undefined,
    );

    this._isIntervalAllowed = derived(
      [this._interval, this.minTimeGrain],
      ([interval, minTimeGrain]) => {
        if (!interval || !minTimeGrain) return false;

        const minDateTimeUnit = V1TimeGrainToDateTimeUnit[minTimeGrain];

        return (
          interval.start.startOf(minDateTimeUnit).equals(interval.start) &&
          interval.end.startOf(minDateTimeUnit).equals(interval.end)
        );
      },
    );

    this._comparisonInterval = derived(
      [this._interval, this._comparisonRange, this._range, this._zone],
      ([interval, comparisonRange, range, zone]) => {
        if (range === ALL_TIME_RANGE_ALIAS) return undefined;
        return getComparisonInterval(interval, comparisonRange, zone);
      },
    );
  }

  onSpecChange = (response: CanvasResponse) => {
    this.checkAndSetDefaultComparisonMode(response);
    this.checkIfHasTimeSeries(response);
    this.checkAndSetMinTimeGrain(response);
    this.checkAndSetFirstDayOfWeek(response);
    this.checkAndSetDefaultTimeRange(response);
    this.checkAndSetTimeRangeOptions(response);

    this.firstPass = false;
  };

  processUrl = (searchParams: URLSearchParams) => {
    const { range, comparisonRange, zone, grain } =
      parseSearchParams(searchParams);

    const defaultComparisonRange = getDefaultComparisonRange(range);

    return {
      range,
      grain,
      zone,
      comparisonRange: comparisonRange ?? defaultComparisonRange,
      showTimeComparison: !!comparisonRange,
    };
  };

  onUrlChange = (searchParams: URLSearchParams) => {
    const { range, comparisonRange, zone, grain, showTimeComparison } =
      this.processUrl(searchParams);

    // Component without local time range
    if (this.componentName && !range) return;

    if (range) this._urlRange.set(range);
    if (grain) this._urlGrain.set(grain);
    if (zone) this._zone.set(zone);

    this._comparisonRange.set(comparisonRange);
    this._showTimeComparison.set(showTimeComparison);
  };

  checkIfHasTimeSeries(response: CanvasResponse) {
    const currentValue = get(this.hasTimeSeries);
    let metricsViews = response.metricsViews || {};

    const metricsViewName = getComponentMetricsViewFromSpec(
      this.componentName,
      response,
    );
    if (metricsViewName && metricsViews[metricsViewName]) {
      metricsViews = {
        [metricsViewName]: metricsViews[metricsViewName],
      };
    }
    const newValue = Object.keys(metricsViews).some((metricView) => {
      const metricsViewSpec = metricsViews[metricView]?.state?.validSpec;
      return Boolean(metricsViewSpec?.timeDimension);
    });

    if (currentValue !== newValue) {
      this.hasTimeSeries.set(newValue);
    }
  }

  checkAndSetMinTimeGrain(response: CanvasResponse) {
    const currentValue = get(this.minTimeGrain);
    const minTimeGrain = deriveMinTimeGrain(this.componentName, response);

    if (currentValue !== minTimeGrain) {
      this.minTimeGrain.set(minTimeGrain);
    }

    return minTimeGrain;
  }

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

  checkAndSetDefaultTimeRange(response: CanvasResponse) {
    const defaultTimeRange = response.canvas?.defaultPreset?.timeRange;

    if (defaultTimeRange) {
      this._defaultRange.set(defaultTimeRange);
    }
  }

  checkAndSetDefaultComparisonMode(response: CanvasResponse) {
    const comparisonMode = response.canvas?.defaultPreset?.comparisonMode;

    const currentTimeRange = get(this._range);

    if (!comparisonMode || currentTimeRange) return;

    if (
      comparisonMode === V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
    ) {
      this.set.comparison(true, true);
    }
  }

  checkAndSetTimeRangeOptions(response: CanvasResponse) {
    const timeRangeOptions = response.canvas?.timeRanges ?? [];

    const ranges = timeRangeOptions.map(({ range }) => range).filter(Boolean);

    this.timeRangeOptions.set(ranges as string[]);
  }

  deriveMinMaxTimestamps = (
    runtime: Writable<Runtime>,
    specStore: CanvasSpecResponseStore,
  ): Readable<{ min: DateTime; max: DateTime } | undefined> => {
    let cachedDerived: { min: DateTime; max: DateTime } | undefined = undefined;

    return derived([runtime, specStore], ([{ instanceId }, spec], set) => {
      const metricsViews = spec?.data?.metricsViews || {};

      const metricsViewName = getComponentMetricsViewFromSpec(
        this.componentName,
        spec?.data,
      );

      const metricsReferred = metricsViewName
        ? [metricsViewName]
        : Object.keys(metricsViews);

      if (
        !metricsReferred.length ||
        (metricsViewName && !metricsViews[metricsViewName])
      ) {
        return set(undefined);
      }

      const timeRangeQueries = [...metricsReferred].map((metricView) => {
        return createQueryServiceMetricsViewTimeRange(
          instanceId,
          metricView,
          {},
          {
            query: {
              enabled:
                !!metricsViews[metricView]?.state?.validSpec?.timeDimension,
              staleTime: Infinity,
              gcTime: Infinity,
            },
          },
          queryClient,
        );
      });

      return derived(timeRangeQueries, (timeRanges, querySet) => {
        const isFetching = timeRanges.some((q) => q.isFetching);
        let latestMetricsView: string | undefined = undefined;

        if (isFetching) {
          querySet(cachedDerived);
          return;
        }

        let min: DateTime | undefined = undefined;
        let max: DateTime | undefined = undefined;

        timeRanges.forEach((timeRange, i) => {
          const summary = timeRange.data?.timeRangeSummary || {};
          const metricsViewName = metricsReferred[i];

          if (summary.min) {
            const minDateTime = DateTime.fromISO(summary.min);
            min = min ? DateTime.min(min, minDateTime) : minDateTime;
          }

          if (summary.max) {
            const maxDateTime = DateTime.fromISO(summary.max);
            if (!max || maxDateTime > max) {
              max = maxDateTime;
              latestMetricsView = metricsViewName;
            }
          }
        });

        if (min && max) {
          const minMax = { min, max };

          cachedDerived = minMax;
          if (latestMetricsView) {
            this.latestMetricsView.set(latestMetricsView);
          }
          querySet(minMax);
        } else {
          querySet(cachedDerived);
          return;
        }
      }).subscribe(set);
    });
  };

  set = {
    zone: (timeZone: string, checkIfSet = false) => {
      return this.searchParamsCallback(
        ExploreStateURLParams.TimeZone,
        timeZone,
        checkIfSet,
      );
    },
    range: (range: string, checkIfSet = false) => {
      return this.searchParamsCallback(
        ExploreStateURLParams.TimeRange,
        range,
        checkIfSet,
      );
    },
    grain: (timeGrain: V1TimeGrain, checkIfSet = false) => {
      const mappedTimeGrain = ToURLParamTimeGrainMapMap[timeGrain];
      if (mappedTimeGrain) {
        return this.searchParamsCallback(
          ExploreStateURLParams.TimeGrain,
          mappedTimeGrain,
          checkIfSet,
        );
      }
    },
    comparison: (range: boolean | string, checkIfSet = false) => {
      const showTimeComparison = Boolean(range);

      if (showTimeComparison) {
        if (range === true) {
          const comparisonRange = get(this._comparisonRange);

          if (!comparisonRange) return;

          return this.searchParamsCallback(
            ExploreStateURLParams.ComparisonTimeRange,
            comparisonRange,
            checkIfSet,
          );
        } else if (typeof range === "string") {
          return this.searchParamsCallback(
            ExploreStateURLParams.ComparisonTimeRange,
            range,
            checkIfSet,
          );
        }
      } else {
        return this.searchParamsCallback(
          ExploreStateURLParams.ComparisonTimeRange,
          undefined,
          checkIfSet,
        );
      }
    },
  };

  clearAll = () => {
    this.searchParamsCallback(ExploreStateURLParams.TimeRange, undefined);
    this.searchParamsCallback(ExploreStateURLParams.TimeGrain, undefined);
    this.searchParamsCallback(
      ExploreStateURLParams.ComparisonTimeRange,
      undefined,
    );
    this.searchParamsCallback(ExploreStateURLParams.TimeZone, undefined);
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

function deriveMinTimeGrain(
  componentName: string | undefined,
  spec: CanvasResponse | undefined,
): V1TimeGrain {
  let metricsViews = spec?.metricsViews || {};

  if (componentName) {
    const metricsViewName = getComponentMetricsViewFromSpec(
      componentName,
      spec,
    );

    if (metricsViewName && metricsViews[metricsViewName]) {
      metricsViews = {
        [metricsViewName]: metricsViews[metricsViewName],
      };
    }
  }

  const minTimeGrain = Object.values(metricsViews).reduce<V1TimeGrain>(
    (min: V1TimeGrain, metricsView) => {
      const timeGrain = metricsView?.state?.validSpec?.smallestTimeGrain;

      if (!timeGrain || timeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) {
        return min;
      }

      if (min === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) {
        return timeGrain;
      }

      return isGrainBigger(timeGrain, min) ? timeGrain : min;
    },
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
  );

  return minTimeGrain;
}

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

function getDefaultComparisonRange(
  range: string | undefined,
): string | undefined {
  if (range === ALL_TIME_RANGE_ALIAS) return undefined;
  if (!range) return "rill-PP";

  const RANGE_PATTERNS = [
    {
      patterns: ["rill-TD", "rill-PDC", "DTD", "-1d/d to ref/d"],
      option: TimeComparisonOption.DAY,
    },
    {
      patterns: ["rill-WTD", "rill-PWC", "WTD", "-1w/w to ref/w"],
      option: TimeComparisonOption.WEEK,
    },
    {
      patterns: ["rill-MTD", "rill-PMC", "MTD", "-1M/M to ref/M"],
      option: TimeComparisonOption.MONTH,
    },
    {
      patterns: ["rill-QTD", "rill-PQC", "QTD", "-1q/q to ref/q"],
      option: TimeComparisonOption.QUARTER,
    },
    {
      patterns: ["rill-YTD", "rill-PYC", "YTD", "-1y/y to ref/y"],
      option: TimeComparisonOption.YEAR,
    },
  ];

  const lowerRange = range.toLowerCase();

  for (const { patterns, option } of RANGE_PATTERNS) {
    if (
      patterns.some(
        (pattern) =>
          range.startsWith(pattern) ||
          lowerRange.startsWith(pattern.toLowerCase()),
      )
    ) {
      return option;
    }
  }

  return TimeComparisonOption.CONTIGUOUS;
}
