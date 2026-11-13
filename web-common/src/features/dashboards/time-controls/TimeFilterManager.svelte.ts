import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { RillDefaultConfigProvider } from "@rilldata/web-common/features/dashboards/providers/RillDefaultConfigProvider.svelte.ts";
import type { UrlParamsStore } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";
import {
  V1TimeGrain,
  type V1MetricsViewTimeRangesResponse,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { DEFAULT_TIMEZONE } from "@rilldata/web-common/lib/time/config.ts";
import {
  constructAsOfString,
  constructNewString,
  INHERIT_TIME_RANGE_ALIAS,
} from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import {
  type DerivedInterval,
  getTimeRangesQueryOptions,
  toDerivedInterval,
} from "@rilldata/web-common/features/dashboards/time-controls/rill-time-ranges.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { arrayOrderedEquals } from "@rilldata/web-common/lib/arrayUtils.ts";
import { createQueries } from "@tanstack/svelte-query";
import { untrack } from "svelte";
import {
  allowedGrainsForInterval,
  TokenToV1TimeGrain,
  getGrainOrder,
  MinSupportedDateTimeUnit,
  MinSupportedGrain,
  V1TimeGrainToToken,
  V1TimeGrainToOrder,
} from "@rilldata/web-common/lib/time/new-grains.ts";
import { getAdjustedInterval } from "@rilldata/web-common/lib/time/ranges";
import {
  overrideRillTimeRef,
  parseRillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import {
  RillIsoInterval,
  RillPeriodToGrainInterval,
  type RillTime,
  RillTimeLabel,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";
import { DateTime, Interval } from "luxon";
import { getTruncationGrain } from "@rilldata/web-common/lib/time/rill-time-grains.ts";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  getAvailableComparisonsForTimeRange,
  getComparisonInterval,
} from "@rilldata/web-common/lib/time/comparisons";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { measureSelection } from "@rilldata/web-common/features/dashboards/time-series/measure-selection/measure-selection.ts";
import { getDurationObjectFromMS } from "@rilldata/web-common/lib/time/transforms";
import { getOrderedStartEndDateTime } from "@rilldata/web-common/features/dashboards/time-series/utils.ts";
import { toStore } from "svelte/store";
import { UrlParamsChangeTracker } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";
import type { TimeFiltersConfig } from "@rilldata/web-common/features/dashboards/time-controls/time-filters-config.ts";
import { getValidatedTimeGrain } from "@rilldata/web-common/lib/time/grains";

type ComparisonTimeRangeOption = {
  name: TimeComparisonOption;
  key: number;
};

const TimeFilterParams = new Set([
  ExploreStateURLParams.TimeRange,
  ExploreStateURLParams.TimeGrain,
  ExploreStateURLParams.TimeZone,
  ExploreStateURLParams.TimeDimension,
  ExploreStateURLParams.ComparisonTimeRange,
  ExploreStateURLParams.HighlightedTimeRange,
]);

export const DEFAULT_INHERIT_URL_PARAMS = [
  [ExploreStateURLParams.TimeRange, INHERIT_TIME_RANGE_ALIAS],
  [ExploreStateURLParams.ComparisonTimeRange, INHERIT_TIME_RANGE_ALIAS],
];

type ScrubRange = {
  start: DateTime;
  end: DateTime;
  isScrubbing: boolean;
};

export type TimeControlState = {
  timeRange: string | undefined;
  timeGrain: V1TimeGrain | undefined;
  timeZone: string;
  interval: Interval<true> | undefined;
  apiTimeRange: V1TimeRange;

  comparisonTimeRange: string | undefined;
  showComparison: boolean;
  // Only set while the comparison is shown.
  comparisonInterval: Interval<true> | undefined;
  apiComparisonTimeRange: V1TimeRange | undefined;

  // Undefined until the metrics view specs resolve. Consumers that gate queries on a time range
  // need to tell "this dashboard has no time dimension" apart from "we do not know yet".
  hasTimeSeries: boolean | undefined;
  ready: boolean;
};

type TimeFilterManagerEvents = {
  ready: void;
  rangeResolved: void;
};

export class TimeFilterManager implements UrlParamsStore {
  // State set directly from URL/controls.
  /**
   * Exact value of time range as set from setUrlParams. It can have values like `inherit`.
   */
  public urlTimeRange = $state<string | undefined>(undefined);
  /**
   * Time range whose interval is being resolved, after `inherit` is replaced with the parent's resolved range.
   * `timeRange` catches up with it once the interval resolves.
   */
  public requestedTimeRange: string | undefined;
  /**
   * Resolved time range, taking into account the parent time range when `inherit` is set.
   * Follows `requestedTimeRange` once its interval has resolved.
   */
  public timeRange: string | undefined;
  /**
   * Exact value of time grain as set from setUrlParams.
   */
  public urlTimeGrain = $state<V1TimeGrain | undefined>(undefined);
  /**
   * Resolved time grain. Always has a value, if url param is undefined, it is based on time range's interval.
   */
  public timeGrain: V1TimeGrain;
  /**
   * Time zone in effect: the parent's while the time range is inherited, otherwise `ownTimeZone`.
   */
  public timeZone: string;
  public timeDimension = $state<string | undefined>(undefined);
  /**
   * Exact value of comparison time range as set from setUrlParams. It can have values like `inherit`.
   */
  public urlComparisonTimeRange = $state<string | undefined>(undefined);
  /**
   * Resolved comparison time range, taking into account the parent comparison time range when `inherit` is set.
   */
  public comparisonTimeRange: string | undefined;
  /**
   * Whether the comparison is shown: the parent's while the comparison range is inherited, otherwise `ownShowComparison`.
   */
  public showComparison: boolean;
  /**
   * Whether the time range follows the parent's, i.e. `urlTimeRange` is `inherit`.
   */
  public inheritsTimeRange: boolean;
  /**
   * Whether the comparison follows the parent's, i.e. `urlComparisonTimeRange` is `inherit`.
   */
  public inheritsComparison: boolean;
  /**
   * Whether a child sets a grain of its own instead of following the parent's.
   */
  public hasOwnTimeGrain: boolean;
  public lastDefinedScrubInterval = $state<Interval<true> | undefined>(
    undefined,
  );
  public scrubInterval = $state<ScrubRange | undefined>(undefined);

  // Selected on this instance, from the url or the controls.
  // A child following its parent reads the parent's values instead, see `timeZone` and `showComparison`.
  private ownTimeZone = $state<string>(DEFAULT_TIMEZONE);
  private ownShowComparison = $state<boolean>(false);

  // When set, relative time ranges are anchored at this time instead of now/latest.
  // Used by scheduled report exports to render data as of the report's execution time.
  // A child always uses its parent's.
  private executionTime = $state<string | undefined>(undefined);

  // Values for the params the url leaves out.
  private readonly rillDefaultConfigProvider: RillDefaultConfigProvider;

  // Computed values
  public minDate: DateTime<true> | undefined;
  public maxDate: DateTime<true> | undefined;
  public interval: Interval<true> | undefined;
  public adjustedInterval: Interval<true> | undefined;
  public comparisonInterval: Interval<true> | undefined;
  public adjustedComparisonInterval: Interval<true> | undefined;
  public canPanLeft: boolean;
  public canPanRight: boolean;
  // For convenience
  public timeStart: string | undefined;
  public timeEnd: string | undefined;
  public apiTimeRange: V1TimeRange;
  public comparisonTimeStart: string | undefined;
  public comparisonTimeEnd: string | undefined;
  public apiComparisonTimeRange: V1TimeRange | undefined;

  // RillTime related values
  public parsedTime: RillTime | undefined;
  public truncationGrain: V1TimeGrain | undefined;
  public ref: RillTimeLabel | string | undefined;
  public snapToEnd: boolean;
  public parsedComparisonTime: RillTime | undefined;

  // Options for controls
  public aggregationOptions: V1TimeGrain[];
  public comparisonTimeRangeOptions: ComparisonTimeRangeOption[];

  public hasTimeSeries: boolean | undefined;
  public ready = $state(false);

  private events = new EventEmitter<TimeFilterManagerEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  public paramKeys = TimeFilterParams;
  public readonly storeSync: UrlParamsChangeTracker;

  /** Unsubscribes from the time ranges queries. */
  public readonly cleanup: () => void;
  // Responses behind the current interval, used to skip notifications that bring no new data.
  private committedResponses: V1MetricsViewTimeRangesResponse[] = [];
  // Intervals last resolved by the runtime, one per metrics view.
  // `timeRange`, `interval`, `adjustedInterval` and `timeGrain` derive from it,
  // so they keep their values while the next range is being resolved, and on error.
  private resolved = $state.raw<
    | {
        timeRange: string;
        timeZone: string;
        intervals: DerivedInterval[];
      }
    | undefined
  >(undefined);
  // Set when the range being resolved came from the controls rather than the url.
  private notifyChangeOnCommit = false;

  public constructor(
    private readonly runtimeClient: RuntimeClient,
    private readonly metricsViewsProvider: MetricsViewsProvider,
    private readonly yamlConfigProvider: YAMLConfigProvider,
    public readonly config: TimeFiltersConfig = {},
    public readonly parent?: TimeFilterManager,
  ) {
    this.storeSync = new UrlParamsChangeTracker(this);
    this.rillDefaultConfigProvider = new RillDefaultConfigProvider(
      metricsViewsProvider,
      yamlConfigProvider,
    );

    // An inherited range, zone and comparison follow the parent through these derived inputs,
    // so a change on the parent re-resolves the child's interval through its own query.
    this.inheritsTimeRange = $derived(
      !!this.parent && this.urlTimeRange === INHERIT_TIME_RANGE_ALIAS,
    );
    this.inheritsComparison = $derived(
      !!this.parent && this.urlComparisonTimeRange === INHERIT_TIME_RANGE_ALIAS,
    );
    this.hasOwnTimeGrain = $derived(!!this.parent && !!this.urlTimeGrain);
    this.requestedTimeRange = $derived(
      this.inheritsTimeRange ? this.parent?.timeRange : this.urlTimeRange,
    );
    this.timeZone = $derived(
      this.inheritsTimeRange && this.parent
        ? this.parent.timeZone
        : this.ownTimeZone,
    );
    this.showComparison = $derived(
      this.inheritsComparison && this.parent
        ? this.parent.showComparison
        : this.ownShowComparison,
    );

    // Subscribing keeps the intervals in sync with the cache:
    // an input change fetches, and invalidating a metrics view (when it reconciles) refetches.
    const timeRangesQuery = createQueries(
      { queries: toStore(() => this.getTimeRangesQueryOptions()) },
      queryClient,
    );
    this.cleanup = timeRangesQuery.subscribe(() =>
      this.commitResolvedIntervals(false),
    );

    this.ready = metricsViewsProvider.ready;
    metricsViewsProvider.on("time-specs-loaded", () => {
      this.ready = true;
      this.events.emit("ready");
    });

    this.timeRange = $derived(this.resolved?.timeRange);
    // The latest interval across the metrics views, and the smallest grain their ranges hint at.
    const latest = $derived.by(() => {
      let interval: Interval<true> | undefined = undefined;
      let smallestGrain: V1TimeGrain | undefined = undefined;
      this.resolved?.intervals.forEach(({ interval: candidate, grain }) => {
        if (
          candidate?.isValid &&
          candidate.end &&
          (!interval || interval.end < candidate.end)
        ) {
          interval = candidate;
        }

        if (
          grain &&
          (!smallestGrain ||
            V1TimeGrainToOrder[grain] < V1TimeGrainToOrder[smallestGrain])
        ) {
          smallestGrain = grain;
        }
      });
      return { interval, smallestGrain } as {
        interval: Interval<true> | undefined;
        smallestGrain: V1TimeGrain | undefined;
      };
    });
    this.interval = $derived(latest.interval);
    // TODO: clarify the intended grain when the url grain is not allowed for the new interval (UX follow up).
    //       Today it falls back to the smallest grain the range hints at, then to the smallest allowed grain,
    //       while without a url grain it is validated against the range instead.
    this.timeGrain = $derived.by(() => {
      // Without a grain of its own, a child on an inherited range follows the parent's grain.
      const requestedGrain =
        this.urlTimeGrain ??
        (this.inheritsTimeRange ? this.parent?.timeGrain : undefined);
      if (!this.interval) return requestedGrain ?? MinSupportedGrain;

      const largestMinTimeGrain =
        this.metricsViewsProvider.largestMinTimeGrain ?? MinSupportedGrain;
      if (!requestedGrain) {
        return (
          getValidatedTimeGrain(
            this.interval,
            largestMinTimeGrain,
            undefined,
            this.parsedTime,
            this.metricsViewsProvider.availableTimeGrains,
          ) ?? MinSupportedGrain
        );
      }

      const allowedGrains = allowedGrainsForInterval(
        this.interval,
        largestMinTimeGrain,
        this.metricsViewsProvider.availableTimeGrains,
      );
      if (allowedGrains.includes(requestedGrain)) return requestedGrain;
      if (latest.smallestGrain && allowedGrains.includes(latest.smallestGrain))
        return latest.smallestGrain;
      return allowedGrains[0];
    });
    this.adjustedInterval = $derived(
      this.interval && this.resolved
        ? getAdjustedInterval(
            this.interval,
            this.timeGrain,
            this.resolved.timeZone,
          )
        : undefined,
    );

    this.minDate = $derived.by(() => {
      const minDate = this.metricsViewsProvider.timeRangeSummary?.min
        ? DateTime.fromISO(this.metricsViewsProvider.timeRangeSummary.min)
        : undefined;
      if (!minDate?.isValid) return undefined;
      return minDate;
    });
    this.maxDate = $derived.by(() => {
      const maxDate = this.metricsViewsProvider.timeRangeSummary?.max
        ? DateTime.fromISO(this.metricsViewsProvider.timeRangeSummary.max)
        : undefined;
      if (!maxDate?.isValid) return undefined;
      return maxDate;
    });

    this.canPanLeft = $derived.by(() => {
      if (!this.minDate || !this.interval) return false;
      // Selected start - min > 0
      const diff = this.interval.start.diff(this.minDate);
      return diff.milliseconds > 0;
    });
    this.canPanRight = $derived.by(() => {
      if (!this.maxDate || !this.interval) return false;
      // max - selected end > 0
      const diff = this.maxDate.diff(this.interval.end);
      return diff.milliseconds > 0;
    });

    this.timeStart = $derived.by(() => {
      const start =
        this.lastDefinedScrubInterval?.start ?? this.interval?.start;
      return start?.toJSDate()?.toISOString();
    });
    this.timeEnd = $derived.by(() => {
      const end = this.lastDefinedScrubInterval?.end ?? this.interval?.end;
      return end?.toJSDate()?.toISOString();
    });
    this.apiTimeRange = $derived(<V1TimeRange>{
      start: this.timeStart,
      end: this.timeEnd,
      timeZone: this.timeZone,
      timeDimension: this.timeDimension,
    });

    // TODO: calculate vs scrub
    this.comparisonTimeStart = $derived(
      this.comparisonInterval?.start.toJSDate().toISOString(),
    );
    this.comparisonTimeEnd = $derived(
      this.comparisonInterval?.end.toJSDate().toISOString(),
    );
    this.apiComparisonTimeRange = $derived.by(() => {
      if (!this.comparisonTimeStart || !this.comparisonTimeEnd)
        return undefined;

      return <V1TimeRange>{
        start: this.comparisonTimeStart,
        end: this.comparisonTimeEnd,
        timeZone: this.timeZone,
      };
    });

    this.parsedTime = $derived.by(() => {
      if (!this.timeRange) return undefined;
      try {
        return parseRillTime(this.timeRange);
      } catch {
        return undefined;
      }
    });
    this.truncationGrain = $derived(getTruncationGrain(this.parsedTime));
    this.ref = $derived(
      this.parsedTime?.isOldFormat
        ? RillTimeLabel.Latest
        : this.parsedTime?.asOfLabel?.label,
    );
    this.snapToEnd = $derived(
      this.parsedTime?.isOldFormat
        ? true
        : !!this.parsedTime?.asOfLabel?.offset,
    );

    // The comparison range as selected, with `inherit` replaced by the parent's comparison range.
    const selectedComparisonTimeRange = $derived(
      this.urlComparisonTimeRange === INHERIT_TIME_RANGE_ALIAS
        ? this.parent?.comparisonTimeRange
        : this.urlComparisonTimeRange,
    );
    this.parsedComparisonTime = $derived.by(() => {
      if (!selectedComparisonTimeRange) return undefined;
      try {
        return parseRillTime(selectedComparisonTimeRange);
      } catch {
        return undefined;
      }
    });
    this.comparisonTimeRange = $derived.by(() => {
      // An absolute primary time range has no relative comparison other than the contiguous one,
      // so mark the comparison as contiguous. An absolute comparison window is left as-is:
      // it already names the exact period to compare against.
      if (
        this.parsedTime?.interval instanceof RillIsoInterval &&
        !(this.parsedComparisonTime?.interval instanceof RillIsoInterval)
      ) {
        return TimeComparisonOption.CONTIGUOUS;
      }
      return selectedComparisonTimeRange;
    });
    this.comparisonInterval = $derived.by(() => {
      if (!this.interval || !this.resolved || !selectedComparisonTimeRange)
        return undefined;
      try {
        return getComparisonInterval(
          this.interval,
          selectedComparisonTimeRange,
          this.resolved.timeZone,
        );
      } catch {
        return undefined;
      }
    });
    this.adjustedComparisonInterval = $derived(
      this.comparisonInterval && this.resolved
        ? getAdjustedInterval(
            this.comparisonInterval,
            this.timeGrain,
            this.resolved.timeZone,
          )
        : undefined,
    );

    this.aggregationOptions = $derived(
      allowedGrainsForInterval(
        this.interval,
        this.metricsViewsProvider.largestMinTimeGrain,
        this.metricsViewsProvider.availableTimeGrains,
      ),
    );
    this.comparisonTimeRangeOptions = $derived(
      this.getComparisonTimeRangeOptions(),
    );

    this.hasTimeSeries = $derived.by(() => {
      if (!metricsViewsProvider.ready) return undefined;

      let hasTimeSeries = false;
      for (const metricsView of metricsViewsProvider.metricsViewNames) {
        const spec = metricsViewsProvider.specs[metricsView];
        // Every spec comes from the same ListResources response,
        // so a missing one means they have not resolved and the answer is not knowable yet.
        if (!spec) return undefined;
        if (spec.timeDimension) hasTimeSeries = true;
      }
      return hasTimeSeries;
    });
  }

  public normalizeParams(urlParams: URLSearchParams): URLSearchParams {
    return urlParams;
  }

  public setUrlParams(urlParams: URLSearchParams) {
    if (this.config.log)
      console.log("TimeFilterManager::setUrlParams", urlParams.toString());

    if (!this.config.skipTimeGrain) {
      const urlGrain = urlParams.get(ExploreStateURLParams.TimeGrain);
      this.urlTimeGrain = urlGrain ? TokenToV1TimeGrain[urlGrain] : undefined;
    }

    // Yaml defaults are applied before the url gets here, so a missing param means the rill default.
    // See RillDefaultConfigProvider.
    this.ownTimeZone =
      urlParams.get(ExploreStateURLParams.TimeZone) ??
      this.rillDefaultConfigProvider.defaultTimeZone;

    this.timeDimension =
      urlParams.get(ExploreStateURLParams.TimeDimension) ?? undefined;

    this.ownShowComparison = urlParams.has(
      ExploreStateURLParams.ComparisonTimeRange,
    );
    this.urlComparisonTimeRange =
      urlParams.get(ExploreStateURLParams.ComparisonTimeRange) ??
      TimeComparisonOption.CONTIGUOUS;

    // Apply time range last so that params are taken from url
    if (urlParams.has(ExploreStateURLParams.TimeRange)) {
      this.applyTimeRange(
        urlParams.get(ExploreStateURLParams.TimeRange)!,
        false,
      );
    } else {
      const defaultTimeRange = this.rillDefaultConfigProvider.defaultTimeRange;
      if (defaultTimeRange && !this.config.skipDefaultTimeRange) {
        this.applyTimeRange(defaultTimeRange, false);
      } else {
        this.urlTimeRange = undefined;
        this.resolved = undefined;
      }
    }

    // Apply the highlighted range after the time range, since applying a time range clears it.
    this.scrubInterval = undefined;
    this.lastDefinedScrubInterval = undefined;
    if (urlParams.has(ExploreStateURLParams.HighlightedTimeRange)) {
      try {
        const parsedHighlightRange = parseRillTime(
          urlParams.get(ExploreStateURLParams.HighlightedTimeRange)!,
        );
        if (parsedHighlightRange.interval instanceof RillIsoInterval) {
          const { start, end, interval } =
            parsedHighlightRange.interval.toLuxonTimes();
          this.scrubInterval =
            start && end
              ? {
                  start,
                  end,
                  isScrubbing: false,
                }
              : undefined;
          this.lastDefinedScrubInterval = interval;
        }
      } catch {
        // no-op
      }
    }
  }

  public applyFilterToParams(urlParams: URLSearchParams) {
    if (this.urlTimeRange) {
      urlParams.set(ExploreStateURLParams.TimeRange, this.urlTimeRange);
    } else {
      urlParams.delete(ExploreStateURLParams.TimeRange);
    }

    if (!this.config.skipTimeGrain) {
      const mappedGrain = this.urlTimeGrain
        ? V1TimeGrainToToken[this.urlTimeGrain]
        : undefined;
      if (mappedGrain) {
        urlParams.set(ExploreStateURLParams.TimeGrain, mappedGrain);
      } else {
        urlParams.delete(ExploreStateURLParams.TimeGrain);
      }
    }

    if (this.ownTimeZone !== this.rillDefaultConfigProvider.defaultTimeZone) {
      urlParams.set(ExploreStateURLParams.TimeZone, this.ownTimeZone);
    } else {
      urlParams.delete(ExploreStateURLParams.TimeZone);
    }

    if (this.timeDimension) {
      urlParams.set(ExploreStateURLParams.TimeDimension, this.timeDimension);
    } else {
      urlParams.delete(ExploreStateURLParams.TimeDimension);
    }

    // An inherited comparison is kept whether or not the parent currently shows it.
    if (
      this.urlComparisonTimeRange === INHERIT_TIME_RANGE_ALIAS ||
      (this.ownShowComparison && this.urlComparisonTimeRange)
    ) {
      urlParams.set(
        ExploreStateURLParams.ComparisonTimeRange,
        this.urlComparisonTimeRange,
      );
    } else {
      urlParams.delete(ExploreStateURLParams.ComparisonTimeRange);
    }

    if (this.lastDefinedScrubInterval) {
      const scrubStart = this.lastDefinedScrubInterval.start
        .toJSDate()
        .toISOString();
      const scrubEnd = this.lastDefinedScrubInterval.end
        .toJSDate()
        .toISOString();
      urlParams.set(
        ExploreStateURLParams.HighlightedTimeRange,
        `${scrubStart},${scrubEnd}`,
      );
    } else {
      urlParams.delete(ExploreStateURLParams.HighlightedTimeRange);
    }
  }

  // Mutation methods used by different UI controls

  public onSelectRange(range: string, ignoreSnap?: boolean) {
    if (range === INHERIT_TIME_RANGE_ALIAS) {
      return this.applyTimeRange(range);
    }

    try {
      const parsed = parseRillTime(range);

      const isPeriodToDate =
        parsed.interval instanceof RillPeriodToGrainInterval;

      const rangeGrainOrder =
        getGrainOrder(parsed.rangeGrain) - (isPeriodToDate ? 1 : 0);

      const asOfGrainOrder = getGrainOrder(this.truncationGrain);

      const shouldAppendAsOfString =
        !parsed.asOfLabel && !(parsed.interval instanceof RillIsoInterval);

      if (asOfGrainOrder > rangeGrainOrder && parsed.rangeGrain) {
        this.truncationGrain = parsed.rangeGrain;
      }

      if (shouldAppendAsOfString) {
        const hasAsOfClause = !!this.parsedTime?.asOfLabel;

        const isTruncationGrainAllowed =
          getGrainOrder(this.truncationGrain) >=
          this.metricsViewsProvider.largestMinGrainOrder;
        const newAsOfString = constructAsOfString(
          this.ref ?? RillTimeLabel.Latest,
          ignoreSnap
            ? undefined
            : this.truncationGrain
              ? isTruncationGrainAllowed
                ? this.truncationGrain
                : parsed.rangeGrain
              : (this.metricsViewsProvider.largestMinTimeGrain ??
                V1TimeGrain.TIME_GRAIN_MINUTE),
          hasAsOfClause || this.snapToEnd ? this.snapToEnd : true,
        );

        overrideRillTimeRef(parsed, newAsOfString);
      }

      return this.applyTimeRange(parsed.toString());
    } catch {
      // This function is called in a controlled manner and should not throw
    }
  }

  public onPan(direction: "left" | "right") {
    if (!this.interval) return;

    const currentRangeWidth = this.interval.length("milliseconds");
    const panAmount = getDurationObjectFromMS(currentRangeWidth);

    const newStart = this.interval.start
      .setZone(this.timeZone)
      [direction === "left" ? "plus" : "minus"](panAmount)
      .toISO();
    const newEnd = this.interval.end
      .setZone(this.timeZone)
      [direction === "left" ? "plus" : "minus"](panAmount)
      .toISO();

    this.applyTimeRange(`${newStart} to ${newEnd}`);
  }

  public onSelectGrainEnding(grain: V1TimeGrain | undefined) {
    if (!this.timeRange) return;

    const newString = constructNewString({
      currentString: this.timeRange,
      truncationGrain: grain === this.truncationGrain ? undefined : grain,
      snapToEnd: grain === this.truncationGrain ? false : this.snapToEnd,
      ref: this.ref,
    });

    return this.applyTimeRange(newString);
  }

  public onSelectGrain(grain: V1TimeGrain | undefined) {
    this.urlTimeGrain = grain;
    this.storeSync.stateChanged();
  }

  public onSelectZone(tz: string) {
    this.ownTimeZone = tz;
    if (!this.urlTimeRange || !this.parsedTime) return;

    if (this.parsedTime.interval instanceof RillIsoInterval) {
      // TODO
    } else {
      // Applies the url range rather than the resolved one, so an inherited range stays inherited.
      this.applyTimeRange(this.urlTimeRange);
    }
  }

  public onSelectAsOfOption(
    ref: RillTimeLabel | string | undefined,
    inclusive: boolean,
  ) {
    if (!this.timeRange) return;
    const newString = constructNewString({
      currentString: this.timeRange,
      truncationGrain: this.truncationGrain,
      snapToEnd: ref === "watermark" ? false : inclusive,
      ref,
    });

    return this.applyTimeRange(newString);
  }

  public onSelectTimeDimension(timeDimension: string) {
    this.timeDimension = timeDimension;
    if (this.urlTimeRange) this.applyTimeRange(this.urlTimeRange);
    else this.storeSync.stateChanged();
  }

  public onSelectComparisonRange(range: string) {
    this.urlComparisonTimeRange = range;
    this.ownShowComparison = true;
    this.storeSync.stateChanged();
  }

  public setShowComparison(showComparison: boolean) {
    // Turning an inherited comparison on or off stops following the parent,
    // keeping the comparison range it currently resolves to.
    if (this.urlComparisonTimeRange === INHERIT_TIME_RANGE_ALIAS) {
      this.urlComparisonTimeRange =
        this.comparisonTimeRange ?? TimeComparisonOption.CONTIGUOUS;
    }
    this.ownShowComparison = showComparison;
    this.storeSync.stateChanged();
  }
  public onToggleShowComparison() {
    this.setShowComparison(!this.showComparison);
  }

  public onScrubRange({ start, end, isScrubbing }: ScrubRange) {
    this.scrubInterval = {
      start,
      end,
      isScrubbing,
    };
    if (isScrubbing) return;

    ({ start, end } = getOrderedStartEndDateTime(start, end));
    const newInterval = Interval.fromDateTimes(start, end);
    if (!newInterval.isValid) return;

    this.lastDefinedScrubInterval = newInterval;
    this.storeSync.stateChanged();
  }

  public resetScrubRange(notify = true) {
    this.lastDefinedScrubInterval = undefined;
    this.scrubInterval = undefined;
    if (notify) this.storeSync.stateChanged();
  }

  public getTimeControlStore() {
    return toStore<TimeControlState>(() => ({
      timeRange: this.timeRange,
      timeGrain: this.timeGrain,
      timeZone: this.timeZone,
      interval: this.interval,
      apiTimeRange: this.apiTimeRange,

      comparisonTimeRange: this.comparisonTimeRange,
      showComparison: this.showComparison,
      comparisonInterval: this.showComparison
        ? this.comparisonInterval
        : undefined,
      apiComparisonTimeRange: this.showComparison
        ? this.apiComparisonTimeRange
        : undefined,

      hasTimeSeries: this.hasTimeSeries,
      ready: this.ready,
    }));
  }

  /**
   * Anchors relative ranges at `executionTime`.
   * It is a query input, so a range that has already resolved re-resolves against it.
   */
  public setExecutionTime(executionTime: string | undefined) {
    this.executionTime = executionTime;
  }

  private applyTimeRange(newTimeRange: string, notifyChange = true) {
    // `requestedTimeRange` follows from this, with `inherit` replaced by the parent's range.
    this.urlTimeRange = newTimeRange;

    measureSelection.clear();
    this.resetScrubRange(false);

    // This should be returned by the API, but it is not yet implemented
    const includesTimeZoneOffset = newTimeRange.includes("tz");

    if (includesTimeZoneOffset) {
      const timeZone = newTimeRange.match(/tz (.*)/)?.[1];

      if (timeZone) this.ownTimeZone = timeZone;
    }

    // A range anchored at `now` resolves differently every time,
    // so re-selecting it has to refetch even though the query key is unchanged.
    if (this.requestedTimeRange?.includes("now")) {
      this.getTimeRangesQueryOptions().forEach(({ queryKey }) => {
        void queryClient.invalidateQueries({ queryKey, refetchType: "all" });
      });
    }

    // Change listeners read the resolved interval (explore copies it into its state),
    // so the change is reported from the commit rather than here.
    this.notifyChangeOnCommit = notifyChange;
    // Commit right away when the intervals are cached,
    // otherwise the subscription commits once they resolve.
    this.commitResolvedIntervals(true);
  }

  private getTimeRangesQueryOptions() {
    const expression = this.requestedTimeRange;
    if (!expression) return [];
    return this.metricsViewsProvider.metricsViewNames.map((metricsViewName) =>
      getTimeRangesQueryOptions(this.runtimeClient, {
        metricsViewName,
        expressions: [expression],
        timeZone: this.timeZone,
        timeDimension: this.timeDimension,
        executionTime: this.parent
          ? this.parent.executionTime
          : this.executionTime,
      }),
    );
  }

  /**
   * Commits the intervals for the current inputs once every metrics view has resolved them.
   * Reads the cache for the current query keys instead of the subscription's results,
   * so a response for earlier inputs is never committed for the current ones.
   * Unless `force` is set, it skips responses it has already committed,
   * since the subscription also notifies on changes that do not touch the data.
   */
  private commitResolvedIntervals(force: boolean) {
    // The subscription can call this from inside an effect; reading state here must not make it a dependency.
    untrack(() => {
      const expression = this.requestedTimeRange;
      const queryOptions = this.getTimeRangesQueryOptions();
      if (!expression || !queryOptions.length) return;

      const responses: V1MetricsViewTimeRangesResponse[] = [];
      for (const { queryKey } of queryOptions) {
        const state =
          queryClient.getQueryState<V1MetricsViewTimeRangesResponse>(queryKey);
        // On error the previous interval is kept.
        if (state?.status !== "success" || state.fetchStatus !== "idle") return;
        responses.push(state.data!);
      }
      if (!force && arrayOrderedEquals(responses, this.committedResponses))
        return;
      this.committedResponses = responses;

      const intervals = responses.map((response) =>
        toDerivedInterval(response, expression, this.timeZone),
      );
      // Keep the previous interval when none of the metrics views resolved the range.
      if (!intervals.some(({ interval }) => interval.isValid)) return;
      this.resolved = {
        timeRange: expression,
        timeZone: this.timeZone,
        intervals,
      };

      // All time has nothing before it to compare against.
      if (expression === TimeRangePreset.ALL_TIME)
        this.ownShowComparison = false;

      if (this.config.log)
        console.log(
          "TimeFilterManager::commitResolvedIntervals::dataUpdated",
          expression,
        );
      if (this.notifyChangeOnCommit) {
        this.notifyChangeOnCommit = false;
        this.storeSync.stateChanged();
      }
      this.events.emit("rangeResolved");
    });
  }

  private getComparisonTimeRangeOptions() {
    // Type-safety
    if (
      !this.minDate ||
      !this.maxDate ||
      !this.timeRange ||
      !this.interval?.isValid ||
      !this.interval.start ||
      !this.interval.end
    )
      return [];

    let allOptions: TimeComparisonOption[];

    const timeRange = this.yamlConfigProvider.timeRanges?.find(
      (tr) => tr.range === this.timeRange,
    );
    if (timeRange?.comparisonTimeRanges?.length) {
      allOptions =
        timeRange.comparisonTimeRanges?.map(
          (co) => co.offset as TimeComparisonOption,
        ) ?? [];
      if (this.config.allowCustomTimeRange)
        allOptions.push(TimeComparisonOption.CUSTOM);
    } else {
      allOptions = [...Object.values(TimeComparisonOption)];
      if (!this.config.allowCustomTimeRange) {
        allOptions = allOptions.filter(
          (o) => o !== TimeComparisonOption.CUSTOM,
        );
      }
    }

    const timeComparisonOptions = getAvailableComparisonsForTimeRange(
      this.minDate.toJSDate(),
      this.maxDate.toJSDate(),
      this.interval.start.toJSDate(),
      this.interval.end.toJSDate(),
      allOptions,
      this.timeZone,
    );

    const filteredOptions = timeComparisonOptions
      .map((co, i) => {
        const comparisonTimeRange = getComparisonInterval(
          this.interval,
          co,
          this.timeZone,
        );

        if (
          !comparisonTimeRange ||
          !comparisonTimeRange.length(MinSupportedDateTimeUnit)
        )
          return undefined;
        return <ComparisonTimeRangeOption>{
          name: co,
          key: i,
        };
      })
      .filter(Boolean) as ComparisonTimeRangeOption[];
    if (this.config.allowCustomTimeRange) {
      filteredOptions.push({
        name: TimeComparisonOption.CUSTOM,
        key: filteredOptions.length,
      });
    }
    return filteredOptions;
  }
}
