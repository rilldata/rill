import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { type Interval, DateTime } from "luxon";
import {
  RillIsoInterval,
  RillPeriodToGrainInterval,
  RillTime,
  RillTimeLabel,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  overrideRillTimeRef,
  parseRillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { getTruncationGrain } from "@rilldata/web-common/lib/time/rill-time-grains.ts";
import {
  allowedGrainsForInterval,
  getGrainOrder,
  V1TimeGrainToOrder,
} from "@rilldata/web-common/lib/time/new-grains.ts";
import {
  constructAsOfString,
  constructNewString,
  deriveInterval,
} from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { invalidationForMetricsViewData } from "@rilldata/web-common/runtime-client/invalidation.ts";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export class TimeRangeManager {
  public timeRange = $state<string | undefined>(undefined);
  public timeGrain = $state<V1TimeGrain | undefined>(undefined);
  public timeZone = $state<string>("UTC");
  public timeDimension = $state<string | undefined>(undefined);

  public minDate: DateTime<true> | undefined;
  public maxDate: DateTime<true> | undefined;

  public interval: Interval | undefined;
  public parsedTime: RillTime | undefined;
  public truncationGrain: V1TimeGrain | undefined;
  public ref: RillTimeLabel | string | undefined;
  public snapToEnd: boolean;

  public constructor(
    private readonly runtimeClient: RuntimeClient,
    private readonly metricsViewsProvider: MetricsViewsProvider,
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
  }

  public onSelectRange = (range: string, ignoreSnap?: boolean) => {
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
  };

  public onSelectGrain = (grain: V1TimeGrain | undefined) => {
    if (!this.timeRange) return;

    const newString = constructNewString({
      currentString: this.timeRange,
      truncationGrain: grain === this.truncationGrain ? undefined : grain,
      snapToEnd: grain === this.truncationGrain ? false : this.snapToEnd,
      ref: this.ref,
    });

    return this.applyTimeRange(newString);
  };

  public onSelectZone = (tz: string) => {
    this.timeZone = tz;
    if (!this.timeRange || !this.parsedTime) return;

    if (this.parsedTime.interval instanceof RillIsoInterval) {
      // TODO
    } else {
      void this.applyTimeRange(this.timeRange, tz);
    }
  };

  public onSelectAsOfOption = (
    ref: RillTimeLabel | string | undefined,
    inclusive: boolean,
  ) => {
    if (!this.timeRange) return;
    const newString = constructNewString({
      currentString: this.timeRange,
      truncationGrain: this.truncationGrain,
      snapToEnd: ref === "watermark" ? false : inclusive,
      ref,
    });

    return this.applyTimeRange(newString);
  };

  public onSelectTimeDimension = (timeDimension: string) => {
    this.timeDimension = timeDimension;
    if (this.timeRange) void this.applyTimeRange(this.timeRange);
  };

  private async applyTimeRange(newTimeRange: string, tz = this.timeZone) {
    // If we don't have a valid time range, early return
    if (!this.metricsViewsProvider.timeRangeSummary?.max) return;

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
    this.timeGrain = finalGrain;
  }
}
