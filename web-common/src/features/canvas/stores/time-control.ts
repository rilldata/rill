import { getComponentMetricsViewFromSpec } from "@rilldata/web-common/features/canvas/components/util";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import {
  calculateComparisonTimeRangePartial,
  calculateTimeRangePartial,
  type ComparisonTimeRangeState,
  type TimeRangeState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { toTimeRangeParam } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { fromTimeRangeUrlParam } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { fromTimeRangesParams } from "@rilldata/web-common/features/dashboards/url-state/convertURLToExplorePreset";
import {
  FromURLParamTimeGrainMap,
  ToURLParamTimeGrainMapMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { type DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
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
  deriveInterval,
  normalizeWeekday,
} from "../../dashboards/time-controls/new-time-controls";
import { type CanvasResponse } from "../selector";
import type { SearchParamsStore } from "./canvas-entity";
import {
  V1TimeGrainToDateTimeUnit,
  V1TimeGrainToDefaultRange,
} from "@rilldata/web-common/lib/time/new-grains";

export type MinMax = {
  min: DateTime<true>;
  max: DateTime<true>;
};

function maybeWritable<T>(value?: T): Writable<T | undefined> {
  return writable(value);
}

// This class is used for global and widget time controls on Canvas
// To avoid methods running in widgets, use isWidgetInstance flag
export class TimeControls {
  isWidgetInstance: boolean;
  selectedComparisonTimeRange: Writable<DashboardTimeControls | undefined>;
  showTimeComparison: Writable<boolean>;
  selectedTimezone: Writable<string>;

  // Inclusive start - exclusive end that accounts for time zone and min time grain
  _allTimeInterval: Readable<Interval<true> | undefined>;
  isReady: Readable<boolean>;
  minTimeGrain: Writable<V1TimeGrain> = writable(
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
  );
  interval: Writable<Interval> = writable(
    Interval.fromDateTimes(DateTime.fromJSDate(new Date(0)), DateTime.now()),
  );
  hasTimeSeries: Readable<boolean>;
  timeRangeStateStore: Writable<TimeRangeState | undefined> =
    writable(undefined);
  comparisonRangeStateStore: Readable<ComparisonTimeRangeState | undefined>;
  _canPan: Readable<{ left: boolean; right: boolean }>;
  latestMetricsView = maybeWritable<string>();

  private minMaxTimeStamps: Readable<MinMax | undefined>;

  constructor(
    private specStore: CanvasSpecResponseStore,
    public searchParamsStore: SearchParamsStore,
    public componentName?: string,
    public canvasName?: string,
  ) {
    this.selectedComparisonTimeRange = writable(undefined);
    this.showTimeComparison = writable(false);
    this.selectedTimezone = writable("UTC");
    this.isWidgetInstance = Boolean(componentName);

    this.minMaxTimeStamps = this.deriveMinMaxTimestamps(
      runtime,
      this.specStore,
    );

    this._allTimeInterval = derived(
      [this.minMaxTimeStamps, this.selectedTimezone, this.minTimeGrain],
      ([minMax, selectedTimezone, minTimeGrain]) => {
        return createAllTimeRange(
          minMax?.min,
          minMax?.max,
          selectedTimezone,
          minTimeGrain,
        );
      },
    );

    this.hasTimeSeries = derived(specStore, (spec) => {
      let metricsViews = spec?.data?.metricsViews || {};

      const metricsViewName = getComponentMetricsViewFromSpec(
        componentName,
        spec?.data,
      );
      if (metricsViewName && metricsViews[metricsViewName]) {
        metricsViews = {
          [metricsViewName]: metricsViews[metricsViewName],
        };
      }
      return Object.keys(metricsViews).some((metricView) => {
        const metricsViewSpec = metricsViews[metricView]?.state?.validSpec;
        return Boolean(metricsViewSpec?.timeDimension);
      });
    });

    const listener = derived(
      [
        specStore,
        this.minMaxTimeStamps,
        this.searchParamsStore,
        this.minTimeGrain,
      ],
      async ([spec, minMaxTimeStamps, searchParams, minTimeGrain]) => {
        if (!spec?.data || !minMaxTimeStamps) {
          return undefined;
        }

        if (!searchParams.toString() && !this.isWidgetInstance) {
          this.processSpec(spec.data);
        }

        const {
          timeRange,
          selectedComparisonTimeRange,
          showTimeComparison,
          timeZone,
          // grain: urlGrain,  // Not supported on the Canvas surface at the moment - bgh
        } = parseSearchParams(searchParams);

        // Component does not have local time range
        if (this.isWidgetInstance && !timeRange) return;

        // TODO: figure out a better way of handling this property
        // when it's not consistent across all metrics views - bgh
        const firstMetricsViewName = Object.keys(spec.data.metricsViews)?.[0];
        const firstDayOfWeekOfFirstMetricsView =
          spec.data.metricsViews[firstMetricsViewName]?.state?.validSpec
            ?.firstDayOfWeek;

        if (!firstMetricsViewName) return;

        Settings.defaultWeekSettings = {
          firstDay: normalizeWeekday(firstDayOfWeekOfFirstMetricsView),
          weekend: [6, 7],
          minimalDays: 4,
        };

        const { defaultPreset, timeRanges } = spec.data?.canvas || {};

        const finalRange: string =
          timeRange ||
          defaultPreset?.timeRange ||
          (timeRanges?.[0] as string | undefined) ||
          "PT24H";

        const result = await deriveInterval(
          finalRange,
          firstMetricsViewName,
          timeZone,
        );

        if (result.error || !result.interval.start || !result.interval.end)
          return;

        const selectedTimeRange: DashboardTimeControls = {
          name: finalRange,
          start: result.interval.start?.toJSDate() ?? new Date(),
          end: result.interval.end?.toJSDate() ?? new Date(),
          interval: result.grain,
        };

        this.interval.set(result.interval);

        this.selectedTimezone.set(timeZone);

        if (selectedComparisonTimeRange) {
          this.selectedComparisonTimeRange.set(selectedComparisonTimeRange);
        }

        this.showTimeComparison.set(showTimeComparison);

        const defaultTimeRange = isoDurationToFullTimeRange(
          defaultPreset?.timeRange,
          minMaxTimeStamps.min.toJSDate(),
          minMaxTimeStamps.max.toJSDate(),
          timeZone,
        );

        const timeRangeState = calculateTimeRangePartial(
          {
            start: minMaxTimeStamps.min.toJSDate(),
            end: minMaxTimeStamps.max.toJSDate(),
          },
          selectedTimeRange,
          undefined, // scrub not present in canvas yet
          timeZone,
          defaultTimeRange,
          minTimeGrain,
        );

        if (!timeRangeState) return undefined;

        this.timeRangeStateStore.set(timeRangeState);
      },
    );

    listener.subscribe(() => {});

    this.comparisonRangeStateStore = derived(
      [
        specStore,
        this.minMaxTimeStamps,
        this.selectedComparisonTimeRange,
        this.selectedTimezone,
        this.showTimeComparison,
        this.timeRangeStateStore,
      ],
      ([
        spec,
        minMaxTimeStamps,
        selectedComparisonTimeRange,
        selectedTimezone,
        showTimeComparison,
        timeRangeState,
      ]) => {
        if (!spec?.data || !timeRangeState || !minMaxTimeStamps)
          return undefined;
        const timeRanges = spec.data?.canvas?.timeRanges;
        return calculateComparisonTimeRangePartial(
          timeRanges,
          {
            start: minMaxTimeStamps.min.toJSDate(),
            end: minMaxTimeStamps.max.toJSDate(),
          },
          selectedComparisonTimeRange,
          selectedTimezone,
          undefined, // scrub not present in canvas yet
          showTimeComparison,
          timeRangeState,
        );
      },
    );

    if (!componentName) {
      this.specStore.subscribe((spec) => {
        if (!spec?.data) return;
        this.processSpec(spec.data);
      });
    }

    this._canPan = derived(
      [this.interval, this.minMaxTimeStamps],
      ([interval, minMaxTimeStamps]) => {
        if (!interval || !interval.start || !interval.end || !minMaxTimeStamps)
          return { left: false, right: false };

        const canPanLeft = interval?.start > minMaxTimeStamps.min;
        const canPanRight = interval?.end < minMaxTimeStamps.max;

        return { left: canPanLeft, right: canPanRight };
      },
    );
  }

  processSpec = (spec: CanvasResponse) => {
    const defaultPreset = spec?.canvas?.defaultPreset;
    const timeRanges = spec?.canvas?.timeRanges;

    const minTimeGrain = deriveMinTimeGrain(this.componentName, spec);

    this.minTimeGrain.set(minTimeGrain);

    const initialRange =
      defaultPreset?.timeRange ??
      timeRanges?.[0].range ??
      V1TimeGrainToDefaultRange[minTimeGrain];

    const didSet = this.set.range(initialRange, true, true);

    if (
      didSet &&
      defaultPreset?.comparisonMode ===
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
    ) {
      this.set.comparison("rill-PP", true, true);
    }
  };

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
    zone: (timeZone: string, checkIfSet = false, replaceState = false) => {
      return this.searchParamsStore.set(
        ExploreStateURLParams.TimeZone,
        timeZone,
        checkIfSet,
        replaceState,
      );
    },
    range: (range: string, checkIfSet = false, replaceState = false) => {
      return this.searchParamsStore.set(
        ExploreStateURLParams.TimeRange,
        range,
        checkIfSet,
        replaceState,
      );
    },
    grain: (
      timeGrain: V1TimeGrain,
      checkIfSet = false,
      replaceState = false,
    ) => {
      const mappedTimeGrain = ToURLParamTimeGrainMapMap[timeGrain];
      if (mappedTimeGrain) {
        return this.searchParamsStore.set(
          ExploreStateURLParams.TimeGrain,
          mappedTimeGrain,
          checkIfSet,
          replaceState,
        );
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
          const comparisonStore = get(this.comparisonRangeStateStore);
          const selectedComparisonTimeRange =
            comparisonStore?.selectedComparisonTimeRange;

          if (!selectedComparisonTimeRange) return;

          return this.searchParamsStore.set(
            ExploreStateURLParams.ComparisonTimeRange,
            toTimeRangeParam(selectedComparisonTimeRange),
            checkIfSet,
            replaceState,
          );
        } else if (typeof range === "string") {
          return this.searchParamsStore.set(
            ExploreStateURLParams.ComparisonTimeRange,
            range,
            checkIfSet,
            replaceState,
          );
        }
      } else {
        return this.searchParamsStore.set(
          ExploreStateURLParams.ComparisonTimeRange,
          undefined,
          checkIfSet,
          replaceState,
        );
      }
    },
  };

  clearAll = () => {
    this.searchParamsStore.set(ExploreStateURLParams.TimeRange, undefined);
    this.searchParamsStore.set(ExploreStateURLParams.TimeGrain, undefined);
    this.searchParamsStore.set(
      ExploreStateURLParams.ComparisonTimeRange,
      undefined,
    );
    this.searchParamsStore.set(ExploreStateURLParams.TimeZone, undefined);
  };

  setSelectedComparisonRange = (comparisonTimeRange: DashboardTimeControls) => {
    this.selectedComparisonTimeRange.set(comparisonTimeRange);
  };
}

export function parseSearchParams(urlParams: URLSearchParams) {
  const { preset } = fromTimeRangesParams(urlParams, new Map());

  let timeRange: string | undefined;
  let selectedComparisonTimeRange: DashboardTimeControls | undefined;
  let showTimeComparison = false;
  let timeZone: string = "UTC";
  let grain: V1TimeGrain = V1TimeGrain.TIME_GRAIN_UNSPECIFIED;

  if (preset.timeRange) {
    timeRange = preset.timeRange;
  }

  if (preset.timeGrain && timeRange) {
    grain = FromURLParamTimeGrainMap[preset.timeGrain];
  }

  if (preset.compareTimeRange) {
    selectedComparisonTimeRange = fromTimeRangeUrlParam(
      preset.compareTimeRange,
    );
    showTimeComparison = true;
  } else if (
    preset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
  ) {
    showTimeComparison = true;
  }

  if (
    preset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE
  ) {
    // unset all comparison setting if mode is none
    selectedComparisonTimeRange = undefined;
    showTimeComparison = false;
  }
  if (preset.timezone) {
    timeZone = preset.timezone;
  }

  return {
    timeRange,
    selectedComparisonTimeRange,
    showTimeComparison,
    timeZone,
    grain,
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

export function createAllTimeInterval(
  min: DateTime | Date | undefined,
  max: DateTime | Date | undefined,
  timezone: string | undefined,
  minTimeGrain: V1TimeGrain | undefined,
): Interval<true> | undefined {
  if (!min || !max || !timezone || !minTimeGrain) return undefined;

  const start = min instanceof Date ? DateTime.fromJSDate(min) : min;
  const end = max instanceof Date ? DateTime.fromJSDate(max) : max;

  const interval = Interval.fromDateTimes(
    start.setZone(timezone),
    end
      .setZone(timezone)
      .plus({ [V1TimeGrainToDateTimeUnit[minTimeGrain]]: 1 }),
  );

  if (!interval.isValid) return undefined;

  return interval;
}

export function createInterval(
  min: DateTime | Date | undefined,
  max: DateTime | Date | undefined,
  timezone: string | undefined,
): Interval<true> | undefined {
  if (!min || !max || !timezone) return undefined;

  const start = min instanceof Date ? DateTime.fromJSDate(min) : min;
  const end = max instanceof Date ? DateTime.fromJSDate(max) : max;

  const interval = Interval.fromDateTimes(
    start.setZone(timezone),
    end.setZone(timezone),
  );

  if (!interval.isValid) return undefined;

  return interval;
}
