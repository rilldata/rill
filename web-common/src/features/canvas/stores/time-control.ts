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

type AllTimeRange = TimeRange & { isFetching: boolean };

let lastAllTimeRange: AllTimeRange | undefined;

// This class is used for global and widget time controls on Canvas
// To avoid methods running in widgets, use isWidgetInstance flag
export class TimeControls {
  isWidgetInstance: boolean;
  selectedComparisonTimeRange: Writable<DashboardTimeControls | undefined>;
  showTimeComparison: Writable<boolean>;
  selectedTimezone: Writable<string>;

  allTimeRange: Readable<AllTimeRange>;
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
    this.isWidgetInstance = Boolean(componentName);

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
      [specStore, this.allTimeRange, this.searchParamsStore, this.minTimeGrain],
      async ([spec, allTimeRange, searchParams, minTimeGrain]) => {
        if (!spec?.data || !allTimeRange) {
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
          allTimeRange.start,
          allTimeRange.end,
          timeZone,
        );

        const timeRangeState = calculateTimeRangePartial(
          allTimeRange,
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

    if (!componentName) {
      this.specStore.subscribe((spec) => {
        if (!spec?.data) return;
        this.processSpec(spec.data);
      });
    }

    this._canPan = derived(
      [this.interval, this.allTimeRange],
      ([interval, allTimeRange]) => {
        if (!interval || !interval.start || !interval.end)
          return { left: false, right: false };

        const canPanLeft =
          interval?.start > DateTime.fromJSDate(allTimeRange.start);
        const canPanRight =
          interval?.end < DateTime.fromJSDate(allTimeRange.end);

        return { left: canPanLeft, right: canPanRight };
      },
    );
  }

  processSpec = (spec: CanvasResponse) => {
    const defaultPreset = spec?.canvas?.defaultPreset;
    const timeRanges = spec?.canvas?.timeRanges;

    const minTimeGrain = deriveMinTimeGrain(this.componentName, spec);

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
      this.set.comparison(newComparisonRange.name, true, true);
    }
  };

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
