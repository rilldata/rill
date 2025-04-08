import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";
import {
  getLocalIANA,
  getUTCIANA,
} from "@rilldata/web-common/lib/time/timezone";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { DateTime, IANAZone, Interval } from "luxon";

export function getBlankExploreState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  return <Partial<MetricsExplorerEntity>>{
    activePage: DashboardState_ActivePage.DEFAULT,

    whereFilter: createAndExpression([]),
    dimensionThresholdFilters: [],
    dimensionsWithInlistFilter: [],

    ...getBlankExploreTimeState(metricsViewSpec, exploreSpec, fullTimeRange),
    ...getBlankExploreViewState(exploreSpec),

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

function getBlankExploreTimeState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
): Partial<MetricsExplorerEntity> {
  if (
    !fullTimeRange?.timeRangeSummary?.min ||
    !fullTimeRange?.timeRangeSummary?.max
  ) {
    return {};
  }

  const timeRangeName = getDefaultTimeRange(
    metricsViewSpec.smallestTimeGrain,
    fullTimeRange,
  );

  const timeZone = getDefaultTimeZone(exploreSpec);

  return {
    selectedTimeRange: {
      name: timeRangeName,
    } as DashboardTimeControls,
    selectedTimezone: timeZone,

    showTimeComparison: false,
    selectedComparisonTimeRange: undefined,

    selectedScrubRange: undefined,
    lastDefinedScrubRange: undefined,
  };
}

function getDefaultTimeRange(
  smallestTimeGrain: V1TimeGrain | undefined,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  if (
    !fullTimeRange?.timeRangeSummary?.min ||
    !fullTimeRange?.timeRangeSummary?.max
  ) {
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
      DateTime.fromISO(fullTimeRange?.timeRangeSummary?.min),
      DateTime.fromISO(fullTimeRange?.timeRangeSummary?.max),
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

function getDefaultTimeZone(explore: V1ExploreSpec) {
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

function getBlankExploreViewState(
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
