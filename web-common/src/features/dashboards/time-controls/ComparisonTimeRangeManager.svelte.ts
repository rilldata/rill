import { TimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeRangeManager.svelte.ts";
import {
  RillIsoInterval,
  RillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { type Interval } from "luxon";
import {
  getAvailableComparisonsForTimeRange,
  getComparisonInterval,
} from "@rilldata/web-common/lib/time/comparisons";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { copySubsetParams } from "@rilldata/web-common/lib/url-utils.ts";
import { getAdjustedInterval } from "@rilldata/web-common/lib/time/ranges";

type ComparisonTimeRangeOption = {
  name: TimeComparisonOption;
  key: number;
  interval: Interval<true>;
};

const ComparisonTimeRangeParams = new Set([
  ExploreStateURLParams.ComparisonTimeRange,
]);

export class ComparisonTimeRangeManager {
  public comparisonTimeRange = $state<string | undefined>(undefined);
  public showComparison = $state<boolean>(false);
  public interval = $state<Interval | undefined>(undefined);
  public adjustedInterval = $state<Interval | undefined>(undefined);

  public comparisonTimeRangeOptions: ComparisonTimeRangeOption[];

  public parsedTime: RillTime | undefined;

  public curStateParams = $state(new URLSearchParams());
  public curSetParams = $state(new URLSearchParams());

  public constructor(
    private readonly yamlConfigProvider: YAMLConfigProvider,
    private readonly timeRangeManager: TimeRangeManager,
    private readonly allowCustomTimeRange: boolean,
  ) {
    this.comparisonTimeRangeOptions = $derived(
      this.getComparisonTimeRangeOptions(),
    );

    this.parsedTime = $derived.by(() => {
      if (!this.comparisonTimeRange) return undefined;
      try {
        return parseRillTime(this.comparisonTimeRange);
      } catch {
        return undefined;
      }
    });
  }

  public createListener() {
    $effect(() => {
      const newParams = new URLSearchParams();
      this.applyFilterToParams(newParams);
      if (newParams.toString() === this.curStateParams.toString()) return;
      this.curStateParams = newParams;
    });
  }

  public setUrlParams(searchParams: URLSearchParams) {
    this.curSetParams = copySubsetParams(
      searchParams,
      ComparisonTimeRangeParams,
    );

    if (searchParams.has(ExploreStateURLParams.ComparisonTimeRange)) {
      this.showComparison = true;
      void this.onSelectComparisonRange(
        searchParams.get(ExploreStateURLParams.ComparisonTimeRange)!,
      );
    } else {
      this.showComparison = false;
      this.comparisonTimeRange = undefined;
    }
  }

  public applyFilterToParams(searchParams: URLSearchParams) {
    if (this.showComparison && this.comparisonTimeRange) {
      searchParams.set(
        ExploreStateURLParams.ComparisonTimeRange,
        this.comparisonTimeRange,
      );
    } else {
      searchParams.delete(ExploreStateURLParams.ComparisonTimeRange);
    }
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
          this.timeRangeManager.interval,
          range,
          this.timeRangeManager.timeZone,
        );
        this.adjustedInterval = this.interval
          ? getAdjustedInterval(
              this.interval,
              this.timeRangeManager.timeGrain,
              this.timeRangeManager.timeZone,
            )
          : undefined;
      }
    } catch {
      return undefined;
    }
  }

  public onToggleShowComparison() {
    this.showComparison = !this.showComparison;
  }

  private getComparisonTimeRangeOptions() {
    // Type-safety
    if (
      !this.timeRangeManager.minDate ||
      !this.timeRangeManager.maxDate ||
      !this.timeRangeManager.timeRange ||
      !this.timeRangeManager.interval?.isValid ||
      !this.timeRangeManager.interval.start ||
      !this.timeRangeManager.interval.end
    )
      return [];

    let allOptions: TimeComparisonOption[];

    const timeRange = this.yamlConfigProvider.timeRanges?.find(
      (tr) => tr.range === this.timeRangeManager.timeRange,
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
      this.timeRangeManager.minDate.toJSDate(),
      this.timeRangeManager.maxDate.toJSDate(),
      this.timeRangeManager.interval.start.toJSDate(),
      this.timeRangeManager.interval.end.toJSDate(),
      allOptions,
      this.timeRangeManager.timeZone,
    );

    return timeComparisonOptions
      .map((co, i) => {
        const comparisonTimeRange = getComparisonInterval(
          this.timeRangeManager.interval,
          co,
          this.timeRangeManager.timeZone,
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
