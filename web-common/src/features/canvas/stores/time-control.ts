import { getComponentMetricsViewFromSpec } from "@rilldata/web-common/features/canvas/components/util";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import {
  calculateComparisonTimeRangePartial,
  calculateTimeRangePartial,
  getComparisonTimeRange,
  getTimeGrain,
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
import {
  TimeRangePreset,
  type DashboardTimeControls,
  type TimeRange,
} from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewTimeRange,
  V1ExploreComparisonMode,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import {
  runtime,
  type Runtime,
} from "@rilldata/web-common/runtime-client/runtime-store";
import { Settings } from "luxon";
import {
  derived,
  get,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import { normalizeWeekday } from "../../dashboards/time-controls/new-time-controls";
import type { SearchParamsStore } from "./canvas-entity";
import Timer from "@rilldata/web-common/components/icons/Timer.svelte";

type AllTimeRange = TimeRange & { isFetching: boolean };

let lastAllTimeRange: AllTimeRange | undefined;

export class TimeControls {
  /**
   * Writables which can be updated by the user
   */
  selectedTimeRange: Writable<DashboardTimeControls | undefined>;
  selectedComparisonTimeRange: Writable<DashboardTimeControls | undefined>;
  showTimeComparison: Writable<boolean>;
  selectedTimezone: Writable<string>;

  /**
   * Derived stores based on writables and spec
   */
  allTimeRange: Readable<AllTimeRange>;
  isReady: Readable<boolean>;
  minTimeGrain: Readable<V1TimeGrain>;
  hasTimeSeries: Readable<boolean>;
  timeRangeStateStore: Readable<TimeRangeState | undefined>;
  comparisonRangeStateStore: Readable<ComparisonTimeRangeState | undefined>;

  private isInitialStateSet: boolean = false;
  private specStore: CanvasSpecResponseStore;

  constructor(
    specStore: CanvasSpecResponseStore,
    public searchParamsStore: SearchParamsStore,
    public componentName?: string,
  ) {
    this.allTimeRange = this.combinedTimeRangeSummaryStore(runtime, specStore);
    this.selectedTimeRange = writable(undefined);
    this.selectedComparisonTimeRange = writable(undefined);
    this.showTimeComparison = writable(false);
    this.selectedTimezone = writable("UTC");
    this.componentName = componentName;

    this.specStore = specStore;

    this.minTimeGrain = derived(specStore, (spec) => {
      let metricsViews = spec?.data?.metricsViews || {};

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

          return isGrainBigger(min, timeGrain) ? timeGrain : min;
        },
        V1TimeGrain.TIME_GRAIN_YEAR, // Use max time grain as starting point
      );

      return minTimeGrain;
    });

    this.hasTimeSeries = derived(specStore, (spec) => {
      let metricsViews = spec?.data?.metricsViews || {};

      const metricsViewName = getComponentMetricsViewFromSpec(
        componentName,
        spec,
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

    this.timeRangeStateStore = derived(
      [
        specStore,
        this.allTimeRange,
        this.selectedTimeRange,
        this.selectedTimezone,
        this.minTimeGrain,
      ],
      ([
        spec,
        allTimeRange,
        selectedTimeRange,
        selectedTimezone,
        minTimeGrain,
      ]) => {
        if (!spec?.data || !selectedTimeRange) {
          return undefined;
        }

        // TODO: figure out a better way of handling this property
        // when it's not consistent across all metrics views - bgh
        const firstMetricsView = Object.values(spec.data.metricsViews)?.[0];
        const firstDayOfWeekOfFirstMetricsView =
          firstMetricsView?.state?.validSpec?.firstDayOfWeek;

        Settings.defaultWeekSettings = {
          firstDay: normalizeWeekday(firstDayOfWeekOfFirstMetricsView),
          weekend: [6, 7],
          minimalDays: 4,
        };

        const { defaultPreset } = spec.data?.canvas || {};
        const defaultTimeRange = isoDurationToFullTimeRange(
          defaultPreset?.timeRange,
          allTimeRange.start,
          allTimeRange.end,
          selectedTimezone,
        );

        const timeRangeState = calculateTimeRangePartial(
          allTimeRange,
          selectedTimeRange,
          undefined, // scrub not present in canvas yet
          selectedTimezone,
          defaultTimeRange,
          minTimeGrain,
        );
        if (!timeRangeState) return undefined;
        return { ...timeRangeState };
      },
    );

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
        if (!spec?.data || !timeRangeState) return undefined;
        const timeRanges = spec.data?.canvas?.timeRanges;
        return calculateComparisonTimeRangePartial(
          timeRanges,
          allTimeRange,
          selectedComparisonTimeRange,
          selectedTimezone,
          undefined, // scrub not present in canvas yet
          showTimeComparison,
          timeRangeState,
        );
      },
    );

    this.searchParamsStore.subscribe((searchParams) => {
      const string = searchParams.toString();

      if (!string) return;

      this.setTimeFiltersFromText(string);
    });

    // this.setInitialState();
  }

  // setInitialState = () => {
  //   const defaultStore = derived(
  //     [this.allTimeRange, this.specStore],
  //     ([allTimeRange, spec]) => {
  //       if (!spec?.data || allTimeRange.isFetching || this.isInitialStateSet) {
  //         return;
  //       }

  //       const isLocalComponentControl = Boolean(this.componentName);

  //       const selectedTimezone = get(this.selectedTimezone);
  //       const comparisonTimeRange = get(this.selectedComparisonTimeRange);

  //       const { defaultPreset } = spec.data?.canvas || {};
  //       const timeRanges = spec?.data?.canvas?.timeRanges;

  //       const defaultTimeRange = isoDurationToFullTimeRange(
  //         defaultPreset?.timeRange,
  //         allTimeRange.start,
  //         allTimeRange.end,
  //         selectedTimezone,
  //       );

  //       const newTimeRange: DashboardTimeControls = {
  //         name: defaultTimeRange.name,
  //         start: defaultTimeRange.start,
  //         end: defaultTimeRange.end,
  //       };

  //       newTimeRange.interval = getTimeGrain(
  //         undefined,
  //         newTimeRange,
  //         V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
  //       );

  //       const newComparisonRange = getComparisonTimeRange(
  //         timeRanges,
  //         allTimeRange,
  //         newTimeRange,
  //         comparisonTimeRange,
  //       );

  //       this.selectedComparisonTimeRange.set(newComparisonRange);

  //       if (
  //         defaultPreset?.comparisonMode ===
  //           V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME &&
  //         !isLocalComponentControl
  //       ) {
  //         this.showTimeComparison.set(true);
  //       }

  //       this.selectedTimeRange.set(newTimeRange);
  //       this.isInitialStateSet = true;
  //     },
  //   );

  //   // Subscribe to ensure the derived code runs
  //   defaultStore.subscribe(() => {});
  // };

  combinedTimeRangeSummaryStore = (
    runtime: Writable<Runtime>,
    specStore: CanvasSpecResponseStore,
  ): Readable<AllTimeRange> => {
    return derived([runtime, specStore], ([r, spec], set) => {
      const metricsViews = spec?.data?.metricsViews || {};

      const metricsViewName = getComponentMetricsViewFromSpec(
        this.componentName,
        spec,
      );

      const metricsReferred = metricsViewName
        ? [metricsViewName]
        : Object.keys(metricsViews);

      if (
        !metricsReferred.length ||
        (metricsViewName && !metricsViews[metricsViewName])
      ) {
        return set({
          name: TimeRangePreset.ALL_TIME,
          start: new Date(0),
          end: new Date(),
          isFetching: false,
        });
      }

      const timeRangeQueries = [...metricsReferred].map((metricView) => {
        return createQueryServiceMetricsViewTimeRange(
          r.instanceId,
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
        let start = new Date();
        let end = new Date(0);
        timeRanges.forEach((timeRange) => {
          const metricsStart = timeRange.data?.timeRangeSummary?.min;
          const metricsEnd = timeRange.data?.timeRangeSummary?.max;
          if (metricsStart) {
            const metricsStartDate = new Date(metricsStart);
            start = new Date(
              Math.min(start.getTime(), metricsStartDate.getTime()),
            );
          }
          if (metricsEnd) {
            const metricsEndDate = new Date(metricsEnd);
            end = new Date(Math.max(end.getTime(), metricsEndDate.getTime()));
          }
        });
        if (start.getTime() >= end.getTime()) {
          start = new Date(0);
          end = new Date();
        }
        let newTimeRange: AllTimeRange = {
          name: TimeRangePreset.ALL_TIME,
          start,
          end,
          isFetching,
        };
        if (isFetching && lastAllTimeRange) {
          newTimeRange = { ...lastAllTimeRange, isFetching };
        }
        const noChange =
          lastAllTimeRange &&
          lastAllTimeRange.start.getTime() === newTimeRange.start.getTime() &&
          lastAllTimeRange.end.getTime() === newTimeRange.end.getTime();

        /**
         * TODO: We want to avoid updating the store when there is no change
         * to avoid any downstream updates which depend on allTimeRange
         * Returning without setting any value leads to undefined value in the store.
         */
        if (noChange) {
          querySet(lastAllTimeRange);
          return;
        }

        if (!isFetching) {
          lastAllTimeRange = newTimeRange;
        }
        querySet(newTimeRange);
      }).subscribe(set);
    });
  };

  set = {
    zone: (timeZone: string) => {
      this.searchParamsStore.set(ExploreStateURLParams.TimeZone, timeZone);
    },
    range: (range: string) => {
      this.searchParamsStore.set(ExploreStateURLParams.TimeRange, range);
    },
    grain: (timeGrain: V1TimeGrain) => {
      const mappedTimeGrain = ToURLParamTimeGrainMapMap[timeGrain];
      if (mappedTimeGrain) {
        this.searchParamsStore.set(
          ExploreStateURLParams.TimeGrain,
          mappedTimeGrain,
        );
      }
    },
    comparison: (range: boolean | string) => {
      const showTimeComparison = Boolean(range);
      this.showTimeComparison.set(showTimeComparison);
      if (showTimeComparison) {
        if (range === true) {
          const comparisonStore = get(this.comparisonRangeStateStore);
          const selectedComparisonTimeRange =
            comparisonStore?.selectedComparisonTimeRange;
          if (!selectedComparisonTimeRange) return;
          this.searchParamsStore.set(
            ExploreStateURLParams.ComparisonTimeRange,
            toTimeRangeParam(selectedComparisonTimeRange),
          );
        } else if (typeof range === "string") {
          this.searchParamsStore.set(
            ExploreStateURLParams.ComparisonTimeRange,
            range,
          );
        }
      } else {
        this.searchParamsStore.set(
          ExploreStateURLParams.ComparisonTimeRange,
          undefined,
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

  displayTimeComparison = (showTimeComparison: boolean) => {
    this.showTimeComparison.set(showTimeComparison);
  };

  setTimeFiltersFromText = (timeFilter: string) => {
    const {
      selectedTimeRange,
      selectedComparisonTimeRange,
      showTimeComparison,
      timeZone,
    } = getTimeRangeFromText(timeFilter);

    console.log({
      cn: this.componentName,
      selectedTimeRange,
      selectedComparisonTimeRange,
      showTimeComparison,
      timeZone,
    });

    if (!selectedTimeRange) {
      const specStore = get(this.specStore);

      const { defaultPreset } = specStore.data?.canvas || {};
      const timeRanges = specStore?.data?.canvas?.timeRanges;

      this.selectedTimeRange.set({
        name:
          (defaultPreset?.timeRange as TimeRangePreset) ||
          timeRanges?.[0] ||
          TimeRangePreset.LAST_24_HOURS,
      } as DashboardTimeControls);
    } else {
      this.selectedTimeRange.set(selectedTimeRange);
    }

    this.selectedTimezone.set(timeZone ?? "UTC");

    if (selectedComparisonTimeRange) {
      this.selectedComparisonTimeRange.set(selectedComparisonTimeRange);
    }

    this.showTimeComparison.set(showTimeComparison);
  };
}

export function getTimeRangeFromText(timeFilter: string) {
  const urlParams = new URLSearchParams(timeFilter);
  const { preset, errors } = fromTimeRangesParams(urlParams, new Map());

  if (errors?.length) {
    console.warn(errors);
    return {
      selectedTimeRange: undefined,
      selectedComparisonTimeRange: undefined,
      showTimeComparison: false,
      timeZone: "UTC",
    };
  }
  let selectedTimeRange: DashboardTimeControls | undefined;
  let selectedComparisonTimeRange: DashboardTimeControls | undefined;
  let showTimeComparison = false;
  let timeZone: string = "UTC";

  if (preset.timeRange) {
    selectedTimeRange = fromTimeRangeUrlParam(preset.timeRange);
  }

  if (preset.timeGrain && selectedTimeRange) {
    selectedTimeRange.interval = FromURLParamTimeGrainMap[preset.timeGrain];
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
    selectedTimeRange,
    selectedComparisonTimeRange,
    showTimeComparison,
    timeZone,
  };
}
