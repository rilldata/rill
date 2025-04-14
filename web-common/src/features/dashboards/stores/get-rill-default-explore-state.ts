import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import { ToURLParamTimeGrainMapMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  getLocalIANA,
  getUTCIANA,
} from "@rilldata/web-common/lib/time/timezone";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  V1TimeGrain,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { DateTime, IANAZone, Interval } from "luxon";
import type { MetricsExplorerEntity } from "./metrics-explorer-entity";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { createAndExpression } from "./filter-utils";
import { TDDChart } from "../time-dimension-details/types";

export function getRillDefaultExploreState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  return <Partial<MetricsExplorerEntity>>{
    activePage: DashboardState_ActivePage.DEFAULT,

    whereFilter: createAndExpression([]),
    dimensionThresholdFilters: [],
    dimensionsWithInlistFilter: [],

    ...getRillDefaultExploreTimeState(
      metricsViewSpec,
      exploreSpec,
      timeRangeSummary,
    ),
    ...getRillDefaultExploreViewState(exploreSpec),

    tdd: {
      expandedMeasureName: "",
      chartType: TDDChart.DEFAULT,
      pinIndex: -1,
    },

    pivot: {
      active: false,
      rows: [],
      columns: [],
      sorting: [],
      expanded: {},
      columnPage: 1,
      rowPage: 1,
      enableComparison: true,
      activeCell: null,
      tableMode: "nest",
    },
  };
}

function getRillDefaultExploreTimeState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
): Partial<MetricsExplorerEntity> {
  if (!timeRangeSummary?.min || !timeRangeSummary?.max) {
    return {};
  }

  const timeRangeName = getDefaultTimeRange(
    metricsViewSpec.smallestTimeGrain,
    timeRangeSummary,
  );

  const timeZone = getDefaultTimeZone(exploreSpec);

  return {
    selectedTimeRange: {
      name: timeRangeName,
      interval: timeRangeName
        ? getGrainForRange(timeRangeName, timeZone, timeRangeSummary)
        : undefined,
    } as DashboardTimeControls,
    selectedTimezone: timeZone,

    showTimeComparison: false,
    selectedComparisonTimeRange: undefined,

    selectedScrubRange: undefined,
    lastDefinedScrubRange: undefined,
  };
}

function getRillDefaultExploreViewState(
  exploreSpec: V1ExploreSpec,
): Partial<MetricsExplorerEntity> {
  return {
    visibleMeasures: exploreSpec.measures ?? [],
    allMeasuresVisible: true,

    visibleDimensions: exploreSpec.dimensions ?? [],
    allDimensionsVisible: true,

    leaderboardSortByMeasureName: exploreSpec.measures?.[0],
    dashboardSortType: DashboardState_LeaderboardSortType.VALUE,
    sortDirection: DashboardState_LeaderboardSortDirection.DESCENDING,

    leaderboardMeasureCount: 1,

    selectedDimensionName: "",
  };
}

export function getDefaultTimeRange(
  smallestTimeGrain: V1TimeGrain | undefined,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  if (!timeRangeSummary?.min || !timeRangeSummary?.max) {
    return undefined;
  }

  if (
    smallestTimeGrain &&
    smallestTimeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED
  ) {
    switch (smallestTimeGrain) {
      case V1TimeGrain.TIME_GRAIN_SECOND:
      case V1TimeGrain.TIME_GRAIN_MINUTE:
        return TimeRangePreset.LAST_SIX_HOURS;
      case V1TimeGrain.TIME_GRAIN_HOUR:
        return TimeRangePreset.LAST_24_HOURS;
      case V1TimeGrain.TIME_GRAIN_DAY:
        return TimeRangePreset.LAST_7_DAYS;
      case V1TimeGrain.TIME_GRAIN_WEEK:
        return TimeRangePreset.LAST_4_WEEKS;
      case V1TimeGrain.TIME_GRAIN_MONTH:
        return TimeRangePreset.LAST_3_MONTHS;
      case V1TimeGrain.TIME_GRAIN_YEAR:
        return "P2Y";
      default:
        return TimeRangePreset.LAST_7_DAYS;
    }
  } else {
    const dayCount = Interval.fromDateTimes(
      DateTime.fromISO(timeRangeSummary?.min),
      DateTime.fromISO(timeRangeSummary?.max),
    )
      .toDuration()
      .as("days");

    let preset: TimeRangePreset = TimeRangePreset.LAST_12_MONTHS;

    if (dayCount <= 2) {
      preset = TimeRangePreset.LAST_SIX_HOURS;
    } else if (dayCount <= 14) {
      preset = TimeRangePreset.LAST_7_DAYS;
    } else if (dayCount <= 60) {
      preset = TimeRangePreset.LAST_4_WEEKS;
    } else if (dayCount <= 180) {
      preset = TimeRangePreset.QUARTER_TO_DATE;
    }

    return preset;
  }
}

export function getGrainForRange(
  timeRangeName: string,
  timezone: string | undefined,
  timeRangeSummary: V1TimeRangeSummary,
) {
  const fullTimeStart = new Date(timeRangeSummary.min!);
  const fullTimeEnd = new Date(timeRangeSummary.max!);
  const timeRange = isoDurationToFullTimeRange(
    timeRangeName,
    fullTimeStart,
    fullTimeEnd,
    timezone,
  );

  return getDefaultTimeGrain(timeRange.start, timeRange.end);
}

export function getDefaultTimeZone(explore: V1ExploreSpec) {
  const preference = explore.timeZones?.[0] ?? DEFAULT_TIMEZONES[0];

  if (preference === "Local") {
    return getLocalIANA();
  } else {
    try {
      const zone = new IANAZone(preference);

      if (zone.isValid) {
        return preference;
      } else {
        throw new Error("Invalid timezone");
      }
    } catch {
      return getUTCIANA();
    }
  }
}
