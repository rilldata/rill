import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
import {
  V1ExploreComparisonMode,
  V1TimeGrain,
  type V1ExploreTimeRange,
} from "@rilldata/web-common/runtime-client";
import { Settings } from "luxon";
import { derived, get, writable, type Readable } from "svelte/store";
import { normalizeWeekday } from "../../dashboards/time-controls/new-time-controls";
import { type CanvasResponse } from "../selector";
import type { CanvasEntity, SearchParamsStore } from "./canvas-entity";
import { maybeWritable } from "@rilldata/web-common/lib/store-utils";
import { TimeState, type MinMax } from "./time-state";

export class TimeManager {
  minMaxTimeStamps: Readable<MinMax | undefined>;
  largestMinTimeGrain: Readable<V1TimeGrain>;

  hasTimeSeriesMap = writable<Map<string, boolean>>(new Map());
  minTimeGrainMap = writable<Map<string, V1TimeGrain>>(new Map());
  hasTimeSeriesStore: Readable<boolean>;

  defaultTimeRangeStore = maybeWritable<string>();
  defaultComparisonRangeStore = maybeWritable<string>();
  timeRangeOptionsStore = writable<V1ExploreTimeRange[]>([]);
  availableTimeZonesStore = writable<string[]>([]);
  allowCustomRangeStore = writable<boolean>(true);
  specInitialized = false;
  state: TimeState;

  constructor(
    public searchParamsStore: SearchParamsStore,
    public parent: CanvasEntity,
  ) {
    this.largestMinTimeGrain = derived(this.minTimeGrainMap, (map) => {
      let largest: V1TimeGrain = V1TimeGrain.TIME_GRAIN_UNSPECIFIED;

      for (const grain of map.values()) {
        if (isGrainBigger(grain, largest)) {
          largest = grain;
        }
      }
      return largest;
    });

    this.hasTimeSeriesStore = derived(
      [this.hasTimeSeriesMap],
      ([$hasTimeSeriesMap]) => {
        const values = Array.from($hasTimeSeriesMap.values());

        return values.some((v) => v);
      },
    );

    this.state = new TimeState(this.searchParamsStore, this.parent, this);

    this.minMaxTimeStamps = this.state.minMaxTimeStamps;
  }

  createLocalTimeState = (
    componentName: string,
    searchParamsStore: SearchParamsStore,
    mericsViewName: string,
  ) => {
    return new TimeState(
      searchParamsStore,
      this.parent,
      this,
      componentName,
      mericsViewName,
    );
  };

  onSpecChange = (spec: CanvasResponse) => {
    this.specInitialized = true;
    const defaultPreset = spec?.canvas?.defaultPreset;

    const ranges = this.checkAndSetTimeRangeOptions(spec);
    this.checkAndSetDefaultTimeRange(spec, ranges);
    this.checkAndSetAvailableTimeZones(spec);
    this.checkIfHasTimeSeries(spec);
    this.checkAndSetFirstDayOfWeek(spec);

    if (
      defaultPreset?.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
    ) {
      this.defaultComparisonRangeStore.set("rill-PP");
    } else {
      this.defaultComparisonRangeStore.set(undefined);
    }
  };

  checkAndSetFirstDayOfWeek(response: CanvasResponse) {
    // TODO: figure out a better way of handling this property
    // when it's not consistent across all metrics views - bgh
    const firstMetricsViewWithTimeSeries = Object.entries(
      response.metricsViews || {},
    ).find(([, mv]) => Boolean(mv?.state?.validSpec?.timeDimension))?.[0];

    if (!firstMetricsViewWithTimeSeries) return;

    const firstMetricsViewName = firstMetricsViewWithTimeSeries;

    const firstDayOfWeekOfFirstMetricsView =
      response.metricsViews[firstMetricsViewName]?.state?.validSpec
        ?.firstDayOfWeek;

    Settings.defaultWeekSettings = {
      firstDay: normalizeWeekday(firstDayOfWeekOfFirstMetricsView),
      weekend: [6, 7],
      minimalDays: 4,
    };
  }

  checkIfHasTimeSeries(response: CanvasResponse) {
    const metricsViews = response.metricsViews || {};

    const timeGrainMap = new Map<string, V1TimeGrain>();
    const hasTimeSeriesMap = new Map<string, boolean>();

    Object.entries(metricsViews).forEach(([mvName, mv]) => {
      const metricsViewSpec = mv?.state?.validSpec;
      const timeGrain = metricsViewSpec?.smallestTimeGrain;
      timeGrainMap.set(mvName, timeGrain || V1TimeGrain.TIME_GRAIN_UNSPECIFIED);
      const hasTimeDimension = Boolean(metricsViewSpec?.timeDimension);
      hasTimeSeriesMap.set(mvName, hasTimeDimension);
    });

    this.minTimeGrainMap.set(timeGrainMap);
    this.hasTimeSeriesMap.set(hasTimeSeriesMap);
  }

  checkAndSetAllowCustomRange = (response: CanvasResponse) => {
    const allowCustomRange = response.canvas?.allowCustomTimeRange ?? true;

    this.allowCustomRangeStore.set(allowCustomRange);

    return allowCustomRange;
  };

  checkAndSetAvailableTimeZones(response: CanvasResponse) {
    const timeZones = response.canvas?.timeZones ?? [];

    this.availableTimeZonesStore.set(timeZones);

    return timeZones;
  }

  checkAndSetDefaultTimeRange(
    response: CanvasResponse,

    timeRanges: V1ExploreTimeRange[],
  ) {
    const defaultTimeRange = response.canvas?.defaultPreset?.timeRange;

    const finalRange = defaultTimeRange ?? timeRanges[0]?.range;

    const currentValue = get(this.defaultTimeRangeStore);
    if (currentValue !== finalRange && finalRange) {
      this.defaultTimeRangeStore.set(finalRange);
    }
  }

  checkAndSetTimeRangeOptions(response: CanvasResponse) {
    const timeRangeOptions = response.canvas?.timeRanges ?? [];

    this.timeRangeOptionsStore.set(timeRangeOptions);

    return timeRangeOptions;
  }
}
