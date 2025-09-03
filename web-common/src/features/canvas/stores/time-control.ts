import { getComponentMetricsViewFromSpec } from "@rilldata/web-common/features/canvas/components/util";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import {
  calculateComparisonTimeRangePartial,
  calculateTimeRangePartial,
  getComparisonTimeRange,
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
  type V1MetricsView,
  type V1ResolveCanvasResponse,
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

export class TimeControls {
  /**
   * Writables which can be updated by the user
   */
  selectedComparisonTimeRange: Writable<DashboardTimeControls | undefined>;
  showTimeComparison: Writable<boolean>;
  selectedTimezone: Writable<string>;

  /**
   * Derived stores based on writables and spec
   */
  allTimeRange: Readable<Interval<true> | undefined>;
  isReady: Readable<boolean>;
  minTimeGrain: Writable<V1TimeGrain> = writable(
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
  );
  interval: Writable<Interval> = writable(
    Interval.fromDateTimes(DateTime.fromJSDate(new Date(0)), DateTime.now()),
  );
  hasTimeSeries = writable(false);
  timeRangeStateStore: Writable<TimeRangeState | undefined> =
    writable(undefined);
  timeRangeOptions = writable<string[]>([]);
  comparisonRangeStateStore: Readable<ComparisonTimeRangeState | undefined>;
  defaultTimeRange: Writable<string | undefined> = writable(undefined);

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

    if (currentValue === minTimeGrain) return;

    this.minTimeGrain.set(minTimeGrain);
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
    const currentValue = get(this.defaultTimeRange);
    const defaultTimeRange = response.canvas?.defaultPreset?.timeRange;

    if (currentValue === defaultTimeRange) return;

    this.defaultTimeRange.set(defaultTimeRange);
  }

  checkAndSetTimeRangeOptions(response: CanvasResponse) {
    // const currentValue = get(this.timeRangeOptions);
    const timeRangeOptions = response.canvas?.timeRanges ?? [];

    const ranges = timeRangeOptions.map(({ range }) => range).filter(Boolean);

    this.timeRangeOptions.set(ranges as string[]);
  }

  primaryMetricsView: Writable<string | undefined> = writable(undefined);

  checkAndSetPrimaryMetricsView(response: CanvasResponse) {
    const currentValue = get(this.primaryMetricsView);
    const firstMetricsViewName = Object.keys(response.metricsViews)?.[0];

    if (currentValue === firstMetricsViewName) return;

    this.primaryMetricsView.set(firstMetricsViewName);
  }

  constructor(
    private specStore: CanvasSpecResponseStore,
    public searchParamsStore: SearchParamsStore,
    public componentName?: string,
    public canvasName?: string,
  ) {
    this.allTimeRange = this.combinedTimeRangeSummaryStore(runtime, specStore);
    this.selectedComparisonTimeRange = writable(undefined);
    this.showTimeComparison = writable(false);
    this.selectedTimezone = writable("UTC");
    this.componentName = componentName;

    this.specStore.subscribe((spec) => {
      const response = spec?.data;
      if (!response) return;

      this.checkIfHasTimeSeries(response);
      this.checkAndSetMinTimeGrain(response);
      this.checkAndSetFirstDayOfWeek(response);
      this.checkAndSetPrimaryMetricsView(response);
      this.checkAndSetDefaultTimeRange(response);
      this.checkAndSetTimeRangeOptions(response);
    });

    // this.hasTimeSeries = derived(specStore, (spec) => {
    //   let metricsViews = response.metricsViews || {};

    //   const metricsViewName = getComponentMetricsViewFromSpec(
    //     componentName,
    //     spec?.data,
    //   );
    //   if (metricsViewName && metricsViews[metricsViewName]) {
    //     metricsViews = {
    //       [metricsViewName]: metricsViews[metricsViewName],
    //     };
    //   }
    //   return Object.keys(metricsViews).some((metricView) => {
    //     const metricsViewSpec = metricsViews[metricView]?.state?.validSpec;
    //     return Boolean(metricsViewSpec?.timeDimension);
    //   });
    // });

    const listener = derived(
      [
        this.allTimeRange,
        this.searchParamsStore,
        this.minTimeGrain,
        this.primaryMetricsView,
        this.defaultTimeRange,
        this.timeRangeOptions,
      ],
      async ([
        allTimeRange,
        searchParams,
        minTimeGrain,
        primaryMetricsView,
        defaultTimeRange,
        timeRangeOptions,
      ]) => {
        if (!allTimeRange || !primaryMetricsView) {
          return undefined;
        }

        const {
          timeRange,
          selectedComparisonTimeRange,
          showTimeComparison,
          timeZone,
          grain: urlGrain,
        } = parseSearchParams(searchParams);

        // Component does not have local time range
        if (this.componentName && !timeRange) return;

        // // TODO: figure out a better way of handling this property
        // // when it's not consistent across all metrics views - bgh
        // const firstMetricsViewName = Object.keys(spec.data.metricsViews)?.[0];
        // const firstDayOfWeekOfFirstMetricsView =
        //   spec.data.metricsViews[firstMetricsViewName]?.state?.validSpec
        //     ?.firstDayOfWeek;

        // if (!firstMetricsViewName) return;

        // Settings.defaultWeekSettings = {
        //   firstDay: normalizeWeekday(firstDayOfWeekOfFirstMetricsView),
        //   weekend: [6, 7],
        //   minimalDays: 4,
        // };

        // const { defaultPreset, timeRanges } = spec.data?.canvas || {};

        const finalRange: string =
          timeRange || defaultTimeRange || timeRangeOptions[0] || "PT24H";

        let interval: Interval;
        let grain: V1TimeGrain | undefined;

        // Handle "inf" (All Time) case specially
        if (finalRange === "inf") {
          interval = allTimeRange;
          grain = undefined;
        } else {
          const result = await deriveInterval(
            finalRange,
            allTimeRange,
            primaryMetricsView,
            timeZone,
          );

          if (result.error || !result.interval.start || !result.interval.end)
            return;
          interval = result.interval;
          grain = result.grain;
        }

        const selectedTimeRange: DashboardTimeControls = {
          name: finalRange,
          start: interval.start?.toJSDate() ?? new Date(),
          end: interval.end?.toJSDate() ?? new Date(),
          interval: urlGrain || grain,
        };

        this.interval.set(interval);

        this.selectedTimezone.set(timeZone);

        if (selectedComparisonTimeRange) {
          this.selectedComparisonTimeRange.set(selectedComparisonTimeRange);
        }

        this.showTimeComparison.set(showTimeComparison);

        const derivedDefaultRange = isoDurationToFullTimeRange(
          defaultTimeRange,
          allTimeRange.start.toJSDate(),
          allTimeRange.end.toJSDate(),
          timeZone,
        );

        const timeRangeState = calculateTimeRangePartial(
          {
            start: allTimeRange.start.toJSDate(),
            end: allTimeRange.end.toJSDate(),
          },
          selectedTimeRange,
          undefined, // scrub not present in canvas yet
          timeZone,
          derivedDefaultRange,
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
        this.allTimeRange,
        this.selectedComparisonTimeRange,
        this.selectedTimezone,
        this.showTimeComparison,
        this.timeRangeStateStore,
      ],
      ([
        spec,
        allTimeRange,
        selectedComparisonTimeRange,
        selectedTimezone,
        showTimeComparison,
        timeRangeState,
      ]) => {
        if (!spec?.data || !timeRangeState || !allTimeRange) return undefined;
        const timeRanges = spec.data?.canvas?.timeRanges;
        return calculateComparisonTimeRangePartial(
          timeRanges,
          {
            start: allTimeRange.start.toJSDate(),
            end: allTimeRange.end.toJSDate(),
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
      const test = derived(
        [
          this.specStore,
          this.selectedTimezone,
          this.allTimeRange,
          this.minTimeGrain,
        ],
        ([spec, selectedTimezone, allTimeRange, minTimeGrain]) => {
          if (!spec?.data) return;
          const defaultPreset = spec.data?.canvas?.defaultPreset;
          const timeRanges = spec?.data?.canvas?.timeRanges;

          // const minTimeGrain = deriveMinTimeGrain(
          //   this.componentName,
          //   spec?.data,
          // );

          // this.minTimeGrain.set(minTimeGrain);

          const defaultRange = defaultPreset?.timeRange;

          if (!allTimeRange) return;

          const initialRange = isoDurationToFullTimeRange(
            defaultPreset?.timeRange,
            allTimeRange?.start.toJSDate(),
            allTimeRange?.end.toJSDate(),
            selectedTimezone,
          );

          const didSet = this.set.range(
            initialRange.name ?? fallbackInitialRanges[minTimeGrain] ?? "PT24H",
            true,
          );

          const newComparisonRange = getComparisonTimeRange(
            timeRanges,
            {
              start: allTimeRange.start.toJSDate(),
              end: allTimeRange.end.toJSDate(),
            },
            { name: defaultRange } as DashboardTimeControls,
            undefined,
          );

          if (
            newComparisonRange?.name &&
            didSet &&
            defaultPreset?.comparisonMode ===
              V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
          ) {
            this.set.comparison(newComparisonRange.name, true);
          }
        },
      );

      test.subscribe(() => {});
    }
  }

  combinedTimeRangeSummaryStore = (
    runtime: Writable<Runtime>,
    specStore: CanvasSpecResponseStore,
  ): Readable<Interval<true> | undefined> => {
    let cachedInterval: Interval<true> | undefined = undefined;

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

        if (isFetching) {
          querySet(cachedInterval);
          return;
        }

        let start: DateTime | undefined;
        let end: DateTime | undefined;

        timeRanges.forEach((timeRange) => {
          const { min, max } = timeRange.data?.timeRangeSummary || {};

          if (min) {
            const minDateTime = DateTime.fromISO(min);
            start = start ? DateTime.min(start, minDateTime) : minDateTime;
          }

          if (max) {
            const maxDateTime = DateTime.fromISO(max);
            end = end ? DateTime.max(end, maxDateTime) : maxDateTime;
          }
        });

        if (start && end) {
          const interval = Interval.fromDateTimes(start, end);
          if (interval.isValid) {
            cachedInterval = interval;
            querySet(interval);
            return;
          } else {
            querySet(cachedInterval);
            return;
          }
        } else {
          querySet(cachedInterval);
          return;
        }
      }).subscribe(set);
    });
  };

  set = {
    zone: (timeZone: string, checkIfSet = false) => {
      return this.searchParamsStore.set(
        ExploreStateURLParams.TimeZone,
        timeZone,
        checkIfSet,
      );
    },
    range: (range: string, checkIfSet = false) => {
      return this.searchParamsStore.set(
        ExploreStateURLParams.TimeRange,
        range,
        checkIfSet,
      );
    },
    grain: (timeGrain: V1TimeGrain, checkIfSet = false) => {
      const mappedTimeGrain = ToURLParamTimeGrainMapMap[timeGrain];
      if (mappedTimeGrain) {
        return this.searchParamsStore.set(
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
          const comparisonStore = get(this.comparisonRangeStateStore);
          const selectedComparisonTimeRange =
            comparisonStore?.selectedComparisonTimeRange;

          if (!selectedComparisonTimeRange) return;

          return this.searchParamsStore.set(
            ExploreStateURLParams.ComparisonTimeRange,
            toTimeRangeParam(selectedComparisonTimeRange),
            checkIfSet,
          );
        } else if (typeof range === "string") {
          return this.searchParamsStore.set(
            ExploreStateURLParams.ComparisonTimeRange,
            range,
            checkIfSet,
          );
        }
      } else {
        return this.searchParamsStore.set(
          ExploreStateURLParams.ComparisonTimeRange,
          undefined,
          checkIfSet,
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

const fallbackInitialRanges = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "PT24H",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "PT24H",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "PT24H",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "PT24H",
  [V1TimeGrain.TIME_GRAIN_DAY]: "P7D",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "P4W",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "P3M",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "P3Q",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "P5Y",
};

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
