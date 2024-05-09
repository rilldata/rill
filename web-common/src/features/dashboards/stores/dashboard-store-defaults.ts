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
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import { MetricsViewSpecComparisonMode } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export function setDefaultTimeRange(
  metricsView: V1MetricsViewSpec,
  metricsExplorer: MetricsExplorerEntity,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  // This function implementation mirrors some code in the metricsExplorer.init() function
  if (!fullTimeRange) return;
  const timeZone = get(getLocalUserPreferences()).timeZone;
  const fullTimeStart = new Date(fullTimeRange.timeRangeSummary.min);
  const fullTimeEnd = new Date(fullTimeRange.timeRangeSummary.max);
  const timeRange = isoDurationToFullTimeRange(
    metricsView.defaultTimeRange,
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
  metricsView: V1MetricsViewSpec,
  metricsExplorer: MetricsExplorerEntity,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  switch (metricsView.defaultComparisonMode) {
    case MetricsViewSpecComparisonMode.COMPARISON_MODE_DIMENSION:
      metricsExplorer.selectedComparisonDimension =
        normaliseName(
          metricsView.defaultComparisonDimension,
          metricsView.dimensions,
        ) || metricsView.dimensions?.[0]?.name;
      break;

    case MetricsViewSpecComparisonMode.COMPARISON_MODE_TIME:
      setDefaultComparisonTimeRange(
        metricsView,
        metricsExplorer,
        fullTimeRange,
      );
      break;

    // if default_comparison is not specified it defaults to no comparison
    case MetricsViewSpecComparisonMode.COMPARISON_MODE_UNSPECIFIED:
  }
}

function setDefaultComparisonTimeRange(
  metricsView: V1MetricsViewSpec,
  metricsExplorer: MetricsExplorerEntity,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  if (!fullTimeRange) return;
  metricsExplorer.showTimeComparison = true;

  const preset = ISODurationToTimePreset(metricsView.defaultTimeRange, true);
  const comparisonOption = DEFAULT_TIME_RANGES[preset]
    ?.defaultComparison as TimeComparisonOption;
  if (!comparisonOption) return;

  const fullTimeStart = new Date(fullTimeRange.timeRangeSummary.min);
  const fullTimeEnd = new Date(fullTimeRange.timeRangeSummary.max);
  const comparisonRange = getTimeComparisonParametersForComponent(
    comparisonOption,
    fullTimeStart,
    fullTimeEnd,
    metricsExplorer.selectedTimeRange.start,
    metricsExplorer.selectedTimeRange.end,
  );
  if (!comparisonRange.isComparisonRangeAvailable) return;

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
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
): MetricsExplorerEntity {
  // CAST SAFETY: safe b/c (1) measure.name is a string if defined,
  // and (2) we filter out undefined values
  const defaultMeasureNames = (metricsView?.measures
    ?.map((measure) => measure?.name)
    .filter((name) => name !== undefined) ?? []) as string[];

  // CAST SAFETY: safe b/c (1) measure.name is a string if defined,
  // and (2) we filter out undefined values
  const defaultDimNames = (metricsView?.dimensions
    ?.map((dim) => dim.name)
    .filter((name) => name !== undefined) ?? []) as string[];

  const metricsExplorer: MetricsExplorerEntity = {
    name,
    visibleMeasureKeys: metricsView.defaultMeasures?.length
      ? new Set(
          metricsView.defaultMeasures
            .map((dm) => normaliseName(dm, metricsView.measures))
            .filter((dm) => !!dm) as string[],
        )
      : new Set(defaultMeasureNames),
    allMeasuresVisible:
      !metricsView.defaultMeasures?.length ||
      metricsView.defaultMeasures?.length === defaultMeasureNames.length,
    visibleDimensionKeys: metricsView.defaultDimensions?.length
      ? new Set(
          metricsView.defaultDimensions
            .map((dd) => normaliseName(dd, metricsView.dimensions))
            .filter((dd) => !!dd) as string[],
        )
      : new Set(defaultDimNames),
    allDimensionsVisible:
      !metricsView.defaultDimensions?.length ||
      metricsView.defaultDimensions?.length === defaultDimNames.length,
    leaderboardMeasureName: defaultMeasureNames[0],
    whereFilter: createAndExpression([]),
    dimensionThresholdFilters: [],
    dimensionFilterExcludeMode: new Map(),
    leaderboardContextColumn: LeaderboardContextColumn.HIDDEN,
    dashboardSortType: SortType.VALUE,
    sortDirection: SortDirection.DESCENDING,
    selectedTimezone: "UTC",

    activePage: DashboardState_ActivePage.DEFAULT,

    showTimeComparison: false,
    dimensionSearchText: "",
    temporaryFilterName: null,
    tdd: {
      chartType: TDDChart.DEFAULT,
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
      columnPage: 1,
    },
    contextColumnWidths: { ...contextColWidthDefaults },
  };
  // set time range related stuff
  setDefaultTimeRange(metricsView, metricsExplorer, fullTimeRange);
  setDefaultComparison(metricsView, metricsExplorer, fullTimeRange);
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
