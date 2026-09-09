import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { copySubsetParams } from "@rilldata/web-common/lib/url-utils.ts";
import type { UrlParamsStore } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DEFAULT_TIMEZONE } from "@rilldata/web-common/lib/time/config.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { invalidationForMetricsViewData } from "@rilldata/web-common/runtime-client/invalidation.ts";
import {
  constructAsOfString,
  constructNewString,
  deriveInterval,
} from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import {
  allowedGrainsForInterval,
  DateTimeUnitToV1TimeGrain,
  getGrainOrder,
  V1TimeGrainToDateTimeUnit,
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
  RillTime,
  RillTimeLabel,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";
import { DateTime, type Interval } from "luxon";
import { getTruncationGrain } from "@rilldata/web-common/lib/time/rill-time-grains.ts";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types.ts";
import {
  getAvailableComparisonsForTimeRange,
  getComparisonInterval,
} from "@rilldata/web-common/lib/time/comparisons";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { getDefaultTimeRange } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state.ts";

type ComparisonTimeRangeOption = {
  name: TimeComparisonOption;
  key: number;
  interval: Interval<true>;
};

const TimeFilterParams = new Set([
  ExploreStateURLParams.TimeRange,
  ExploreStateURLParams.TimeGrain,
  ExploreStateURLParams.TimeZone,
  ExploreStateURLParams.TimeDimension,
  ExploreStateURLParams.ComparisonTimeRange,
]);

export class TimeFilterManager implements UrlParamsStore {
  // State set directly from URL/controls.
  public timeRange = $state<string | undefined>(undefined);
  public timeGrain = $state<V1TimeGrain | undefined>(undefined);
  public timeZone = $state<string>(DEFAULT_TIMEZONE);
  public timeDimension = $state<string | undefined>(undefined);
  public comparisonTimeRange = $state<string | undefined>(undefined);
  public showComparison = $state<boolean>(false);

  // Computed values
  public minDate: DateTime<true> | undefined;
  public maxDate: DateTime<true> | undefined;
  public interval = $state<Interval | undefined>(undefined);
  public adjustedInterval = $state<Interval | undefined>(undefined);
  public comparisonInterval = $state<Interval | undefined>(undefined);
  public adjustedComparisonInterval = $state<Interval | undefined>(undefined);

  // RillTime related values
  public parsedTime: RillTime | undefined;
  public truncationGrain: V1TimeGrain | undefined;
  public ref: RillTimeLabel | string | undefined;
  public snapToEnd: boolean;
  public parsedComparisonTime: RillTime | undefined;

  public comparisonTimeRangeOptions: ComparisonTimeRangeOption[];

  public curParams = $state(new URLSearchParams());

  public hasTimeSeries: boolean;
  public ready: boolean;

  // Temporary lock in explore. Once we move whereFilter out of explore, we can remove this.
  public updating = false;
  private timeRangeReady = $state(false);

  public constructor(
    private readonly runtimeClient: RuntimeClient,
    private readonly metricsViewsProvider: MetricsViewsProvider,
    private readonly yamlConfigProvider: YAMLConfigProvider,
    private readonly allowCustomTimeRange: boolean,
  ) {
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

    this.parsedComparisonTime = $derived.by(() => {
      if (!this.comparisonTimeRange) return undefined;
      try {
        return parseRillTime(this.comparisonTimeRange);
      } catch {
        return undefined;
      }
    });

    this.comparisonTimeRangeOptions = $derived(
      this.getComparisonTimeRangeOptions(),
    );

    this.hasTimeSeries = $derived(
      Boolean(metricsViewsProvider.timeRangeSummary),
    );

    this.ready = $derived.by(() => {
      if (!metricsViewsProvider.ready) return false;
      for (const metricsView of metricsViewsProvider.metricsViewNames) {
        const spec = metricsViewsProvider.specs[metricsView];
        if (!spec) return false;
        if (!spec.timeDimension) continue;
        if (!metricsViewsProvider.timeRangeSummaries[metricsView]) return false;
      }

      // A dashboard without a time dimension has no time range to resolve,
      // so waiting for one would never let the consumers of this become ready.
      if (!this.hasTimeSeries) return true;

      return this.timeRangeReady;
    });
  }

  public setUrlParams(urlParams: URLSearchParams) {
    this.curParams = copySubsetParams(urlParams, TimeFilterParams);

    this.timeGrain =
      DateTimeUnitToV1TimeGrain[
        urlParams.get(ExploreStateURLParams.TimeGrain)!
      ];

    this.timeZone =
      urlParams.get(ExploreStateURLParams.TimeZone) ??
      this.yamlConfigProvider.defaultTimeZone;

    this.timeDimension =
      urlParams.get(ExploreStateURLParams.TimeDimension) ?? undefined;

    if (urlParams.has(ExploreStateURLParams.TimeRange)) {
      void this.onSelectRange(
        urlParams.get(ExploreStateURLParams.TimeRange)!,
        true,
      );
    } else {
      let defaultTimeRange = this.yamlConfigProvider.defaultTimeRange;
      if (!defaultTimeRange) {
        defaultTimeRange = getDefaultTimeRange(
          this.metricsViewsProvider.smallestTimeGrain,
          this.metricsViewsProvider.timeRangeSummary,
        );
      }
      if (defaultTimeRange) {
        void this.onSelectRange(defaultTimeRange, true);
      } else {
        this.timeRange = undefined;
        this.interval = undefined;
      }
    }

    if (urlParams.has(ExploreStateURLParams.ComparisonTimeRange)) {
      this.showComparison = true;
      void this.onSelectComparisonRange(
        urlParams.get(ExploreStateURLParams.ComparisonTimeRange)!,
      );
    } else {
      this.showComparison = false;
      this.comparisonTimeRange = undefined;
    }
  }

  public applyFilterToParams(urlParams: URLSearchParams) {
    if (this.timeRange) {
      urlParams.set(ExploreStateURLParams.TimeRange, this.timeRange);
    } else {
      urlParams.delete(ExploreStateURLParams.TimeRange);
    }

    const mappedGrain = this.timeGrain
      ? V1TimeGrainToDateTimeUnit[this.timeGrain]
      : undefined;
    if (mappedGrain) {
      urlParams.set(ExploreStateURLParams.TimeGrain, mappedGrain);
    } else {
      urlParams.delete(ExploreStateURLParams.TimeGrain);
    }

    if (this.timeZone !== DEFAULT_TIMEZONE) {
      urlParams.set(ExploreStateURLParams.TimeZone, this.timeZone);
    } else {
      urlParams.delete(ExploreStateURLParams.TimeZone);
    }

    if (this.timeDimension) {
      urlParams.set(ExploreStateURLParams.TimeDimension, this.timeDimension);
    } else {
      urlParams.delete(ExploreStateURLParams.TimeDimension);
    }

    if (this.showComparison && this.comparisonTimeRange) {
      urlParams.set(
        ExploreStateURLParams.ComparisonTimeRange,
        this.comparisonTimeRange,
      );
    } else {
      urlParams.delete(ExploreStateURLParams.ComparisonTimeRange);
    }
  }

  // Mutation methods used by different UI controls

  public onSelectRange(range: string, ignoreSnap?: boolean) {
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
          this.metricsViewsProvider.smallestGrainOrder;
        const newAsOfString = constructAsOfString(
          this.ref ?? RillTimeLabel.Latest,
          ignoreSnap
            ? undefined
            : this.truncationGrain
              ? isTruncationGrainAllowed
                ? this.truncationGrain
                : parsed.rangeGrain
              : (this.metricsViewsProvider.smallestTimeGrain ??
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

  public onSelectGrain(grain: V1TimeGrain | undefined) {
    if (!this.timeRange) return;

    const newString = constructNewString({
      currentString: this.timeRange,
      truncationGrain: grain === this.truncationGrain ? undefined : grain,
      snapToEnd: grain === this.truncationGrain ? false : this.snapToEnd,
      ref: this.ref,
    });

    return this.applyTimeRange(newString);
  }

  public onSelectZone(tz: string) {
    this.timeZone = tz;
    if (!this.timeRange || !this.parsedTime) return;

    if (this.parsedTime.interval instanceof RillIsoInterval) {
      // TODO
    } else {
      void this.applyTimeRange(this.timeRange, tz);
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
    if (this.timeRange) void this.applyTimeRange(this.timeRange);
  }

  public onSelectComparisonRange(range: string) {
    // TODO: reassign when primary time range changes.

    this.comparisonTimeRange = range;
    if (!this.showComparison) {
      this.interval = undefined;
      return;
    }

    try {
      const parsed = parseRillTime(range);
      if (parsed.interval instanceof RillIsoInterval) {
        // TODO
      } else {
        this.interval = getComparisonInterval(
          this.interval,
          range,
          this.timeZone,
        );
        this.adjustedInterval = this.interval
          ? getAdjustedInterval(this.interval, this.timeGrain, this.timeZone)
          : undefined;
      }
    } catch {
      return undefined;
    }
  }

  public onToggleShowComparison() {
    this.showComparison = !this.showComparison;
  }

  private async applyTimeRange(newTimeRange: string, tz = this.timeZone) {
    // The runtime resolves the range against the metrics views, so their names are all this needs.
    // The time range summary can still be loading at this point;
    // waiting for it here would drop the range the dashboard loaded with.
    if (
      !this.metricsViewsProvider.metricsViewNames.length ||
      this.timeRange === newTimeRange
    ) {
      return;
    }

    // This should be returned by the API, but it is not yet implemented
    const includesTimeZoneOffset = newTimeRange.includes("tz");

    if (includesTimeZoneOffset) {
      const timeZone = newTimeRange.match(/tz (.*)/)?.[1];

      if (timeZone) this.timeZone = timeZone;
    }

    await queryClient.cancelQueries({
      predicate: (query) =>
        this.metricsViewsProvider.metricsViewNames.some((mvName) =>
          invalidationForMetricsViewData(query, mvName),
        ),
    });

    const promises = this.metricsViewsProvider.metricsViewNames.map(
      (mvName) => {
        return deriveInterval(
          newTimeRange,
          this.runtimeClient,
          mvName,
          tz ?? "UTC",
          this.timeDimension,
          // executionTime, // TODO
        );
      },
    );
    const intervals = await Promise.all(promises);
    let latestInterval: Interval<true> | undefined = undefined;
    let smallestGrain: V1TimeGrain | undefined = undefined;
    intervals.forEach(({ interval, grain }) => {
      if (
        interval?.isValid &&
        interval.end &&
        (!latestInterval || latestInterval.end < interval.end)
      ) {
        latestInterval = interval;
      }

      if (
        grain &&
        (!smallestGrain ||
          V1TimeGrainToOrder[grain] < V1TimeGrainToOrder[smallestGrain])
      ) {
        smallestGrain = grain;
      }
    });
    if (!latestInterval) return;

    const allowedGrains = allowedGrainsForInterval(
      latestInterval,
      this.metricsViewsProvider.smallestTimeGrain ??
        V1TimeGrain.TIME_GRAIN_MINUTE,
    );

    const finalGrain =
      this.timeGrain && allowedGrains.includes(this.timeGrain)
        ? this.timeGrain
        : smallestGrain && allowedGrains.includes(smallestGrain)
          ? smallestGrain
          : allowedGrains[0];

    this.interval = latestInterval;
    this.adjustedInterval = getAdjustedInterval(
      latestInterval,
      finalGrain,
      this.timeZone,
    );
    this.timeRange = newTimeRange;
    this.timeGrain = finalGrain;
    this.timeRangeReady = true;
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
      if (this.allowCustomTimeRange)
        allOptions.push(TimeComparisonOption.CUSTOM);
    } else {
      allOptions = [...Object.values(TimeComparisonOption)];
      if (!this.allowCustomTimeRange) {
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

    return timeComparisonOptions
      .map((co, i) => {
        const comparisonTimeRange = getComparisonInterval(
          this.interval,
          co,
          this.timeZone,
        );

        if (!comparisonTimeRange) return undefined;
        return <ComparisonTimeRangeOption>{
          name: co,
          key: i,
          interval: comparisonTimeRange,
        };
      })
      .filter(Boolean) as ComparisonTimeRangeOption[];
  }
}
