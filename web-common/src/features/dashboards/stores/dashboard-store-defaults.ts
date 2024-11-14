import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  contextColWidthDefaults,
  type MetricsExplorerEntity,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getPersistentDashboardState } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { getLocalUserPreferences } from "@rilldata/web-common/features/dashboards/user-preferences";
import { getTimeComparisonParametersForComponent } from "@rilldata/web-common/lib/time/comparisons";
import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
import { ISODurationToTimePreset } from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  V1ExplorePreset,
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import { V1ExploreComparisonMode } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export function setDefaultTimeRange(
  explorePreset: V1ExplorePreset | undefined,
  metricsExplorer: MetricsExplorerEntity,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  // This function implementation mirrors some code in the metricsExplorer.init() function
  if (
    !fullTimeRange ||
    !fullTimeRange.timeRangeSummary?.min ||
    !fullTimeRange.timeRangeSummary?.max
  )
    return;
  const timeZone = get(getLocalUserPreferences()).timeZone;
  const fullTimeStart = new Date(fullTimeRange.timeRangeSummary.min);
  const fullTimeEnd = new Date(fullTimeRange.timeRangeSummary.max);
  const timeRange = isoDurationToFullTimeRange(
    explorePreset?.timeRange,
    fullTimeStart,
    fullTimeEnd,
    timeZone,
  );

  const timeGrain = getDefaultTimeGrain(timeRange.start, timeRange.end);
  metricsExplorer.selectedTimeRange = {
    ...timeRange,
    interval: timeGrain.grain,
  };
  metricsExplorer.selectedTimezone = timeZone ?? "UTC";
  // TODO: refactor all sub methods and call setSelectedScrubRange here
  metricsExplorer.selectedScrubRange = undefined;
  metricsExplorer.lastDefinedScrubRange = undefined;
}

function setDefaultComparison(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  metricsExplorer: MetricsExplorerEntity,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  switch (exploreSpec?.defaultPreset?.comparisonMode) {
    case V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION:
      metricsExplorer.selectedComparisonDimension =
        normaliseName(
          exploreSpec?.defaultPreset?.comparisonDimension,
          metricsViewSpec.dimensions,
        ) || exploreSpec.dimensions?.[0];
      break;

    case V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME:
      setDefaultComparisonTimeRange(
        exploreSpec?.defaultPreset,
        metricsExplorer,
        fullTimeRange,
      );
      break;

    // if default_comparison is not specified it defaults to no comparison
    case V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_UNSPECIFIED:
  }
}

function setDefaultComparisonTimeRange(
  explorePreset: V1ExplorePreset | undefined,
  metricsExplorer: MetricsExplorerEntity,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  if (
    !fullTimeRange ||
    !fullTimeRange.timeRangeSummary?.min ||
    !fullTimeRange.timeRangeSummary?.max
  )
    return;
  metricsExplorer.showTimeComparison = true;

  const preset = ISODurationToTimePreset(explorePreset?.timeRange, true);
  if (!preset) return;
  const comparisonOption = DEFAULT_TIME_RANGES[preset]
    ?.defaultComparison as TimeComparisonOption;
  if (!comparisonOption) return;

  const fullTimeStart = new Date(fullTimeRange.timeRangeSummary.min);
  const fullTimeEnd = new Date(fullTimeRange.timeRangeSummary.max);
  const comparisonRange = getTimeComparisonParametersForComponent(
    comparisonOption,
    fullTimeStart,
    fullTimeEnd,
    metricsExplorer.selectedTimeRange?.start,
    metricsExplorer.selectedTimeRange?.end,
  );
  if (
    !comparisonRange.isComparisonRangeAvailable ||
    !comparisonRange.start ||
    !comparisonRange.end
  )
    return;

  metricsExplorer.selectedComparisonTimeRange = {
    name: comparisonOption,
    start: comparisonRange.start,
    end: comparisonRange.end,
  };
  metricsExplorer.leaderboardContextColumn =
    LeaderboardContextColumn.DELTA_PERCENT;
}

export function getDefaultMetricsExplorerEntity(
  name: string,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
): MetricsExplorerEntity {
  const defaultMeasureNames =
    explore?.defaultPreset?.measures ?? explore?.measures ?? [];

  const defaultDimNames =
    explore?.defaultPreset?.dimensions ?? explore?.dimensions ?? [];

  const metricsExplorer: MetricsExplorerEntity = {
    name,
    visibleMeasureKeys: new Set(
      defaultMeasureNames
        .map((dm) => normaliseName(dm, metricsView.measures))
        .filter((dm) => !!dm) as string[],
    ),
    allMeasuresVisible: defaultMeasureNames.length === explore.measures?.length,
    visibleDimensionKeys: new Set(
      defaultDimNames
        .map((dd) => normaliseName(dd, metricsView.dimensions))
        .filter((dd) => !!dd) as string[],
    ),
    allDimensionsVisible: defaultDimNames.length === explore.dimensions?.length,
    leaderboardMeasureName: defaultMeasureNames[0],
    whereFilter: createAndExpression([]),
    havingFilter: createAndExpression([]),
    dimensionThresholdFilters: [],
    dimensionFilterExcludeMode: new Map(),
    leaderboardContextColumn: LeaderboardContextColumn.HIDDEN,
    dashboardSortType: SortType.VALUE,
    sortDirection: SortDirection.DESCENDING,
    selectedTimezone: "UTC",
    selectedTimeRange: undefined,

    activePage: DashboardState_ActivePage.DEFAULT,
    selectedComparisonDimension: undefined,
    selectedDimensionName: undefined,

    showTimeComparison: false,
    temporaryFilterName: null,
    tdd: {
      chartType: TDDChart.DEFAULT,
      expandedMeasureName: "",
      pinIndex: -1,
    },
    pivot: {
      active: false,
      rows: {
        dimension: [],
      },
      columns: {
        dimension: [],
        measure: [],
      },
      rowJoinType: "nest",
      expanded: {},
      sorting: [],
      rowPage: 1,
      enableComparison: true,
      columnPage: 1,
      activeCell: null,
    },
    canvas: {
      active: false,
    },
    contextColumnWidths: { ...contextColWidthDefaults },
  };
  // set time range related stuff
  setDefaultTimeRange(explore?.defaultPreset, metricsExplorer, fullTimeRange);
  setDefaultComparison(metricsView, explore, metricsExplorer, fullTimeRange);
  return metricsExplorer;
}

export function restorePersistedDashboardState(
  metricsExplorer: MetricsExplorerEntity,
) {
  const persistedState = getPersistentDashboardState();
  if (persistedState.visibleMeasures) {
    metricsExplorer.allMeasuresVisible =
      persistedState.visibleMeasures.length ===
      metricsExplorer.visibleMeasureKeys.size; // TODO: check values
    metricsExplorer.visibleMeasureKeys = new Set(
      persistedState.visibleMeasures,
    );
  }
  if (persistedState.visibleDimensions) {
    metricsExplorer.allDimensionsVisible =
      persistedState.visibleDimensions.length ===
      metricsExplorer.visibleDimensionKeys.size; // TODO: check values
    metricsExplorer.visibleDimensionKeys = new Set(
      persistedState.visibleDimensions,
    );
  }
  if (persistedState.leaderboardMeasureName) {
    metricsExplorer.leaderboardMeasureName =
      persistedState.leaderboardMeasureName;
  }
  if (persistedState.dashboardSortType) {
    metricsExplorer.dashboardSortType = persistedState.dashboardSortType;
  }
  if (persistedState.sortDirection) {
    metricsExplorer.sortDirection = persistedState.sortDirection;
  }
  return metricsExplorer;
}

function normaliseName(
  name: string | undefined,
  entities:
    | MetricsViewSpecMeasureV2[]
    | MetricsViewSpecDimensionV2[]
    | undefined,
): string | undefined {
  return entities?.find((e) => e.name?.toLowerCase() === name?.toLowerCase())
    ?.name;
}
