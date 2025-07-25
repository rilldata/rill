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

import type { CanvasResponse } from "../selector";

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
  minTimeGrain: Writable<V1TimeGrain> = writable(
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
  );
  hasTimeSeries: Readable<boolean>;
  timeRangeStateStore: Readable<TimeRangeState | undefined>;
  comparisonRangeStateStore: Readable<ComparisonTimeRangeState | undefined>;

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
      this.setTimeFiltersFromText(searchParams.toString());
    });

    if (!componentName) {
      this.specStore.subscribe((spec) => {
        if (!spec?.data) return;
        const defaultPreset = spec.data?.canvas?.defaultPreset;
        const timeRanges = spec?.data?.canvas?.timeRanges;

        const minTimeGrain = deriveMinTimeGrain(this.componentName, spec?.data);

        this.minTimeGrain.set(minTimeGrain);

        const defaultRange = defaultPreset?.timeRange;

        const allTimeRange = get(this.allTimeRange);
        const selectedTimezone = get(this.selectedTimezone);

        const initialRange = isoDurationToFullTimeRange(
          defaultPreset?.timeRange,
          allTimeRange.start,
          allTimeRange.end,
          selectedTimezone,
        );

        const didSet = this.set.range(
          initialRange.name ?? fallbackInitialRanges[minTimeGrain] ?? "PT24H",
          true,
        );

        const newComparisonRange = getComparisonTimeRange(
          timeRanges,
          get(this.allTimeRange),
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
      });
    }
  }

  combinedTimeRangeSummaryStore = (
    runtime: Writable<Runtime>,
    specStore: CanvasSpecResponseStore,
  ): Readable<AllTimeRange> => {
    return derived([runtime, specStore], ([r, spec], set) => {
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
        if (start.getTime() > end.getTime()) {
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

  setTimeFiltersFromText = (timeFilter: string) => {
    const {
      selectedTimeRange,
      selectedComparisonTimeRange,
      showTimeComparison,
      timeZone,
    } = getTimeRangeFromText(timeFilter);

    if (!selectedTimeRange) {
      const specStore = get(this.specStore);

      const { defaultPreset } = specStore.data?.canvas || {};
      const timeRanges = specStore?.data?.canvas?.timeRanges;

      const finalRange =
        (defaultPreset?.timeRange as TimeRangePreset) || timeRanges?.[0];

      if (finalRange) {
        this.selectedTimeRange.set({
          name: finalRange,
        } as DashboardTimeControls);
      }
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
