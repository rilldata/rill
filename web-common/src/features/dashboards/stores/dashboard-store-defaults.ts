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
import { getLocalUserPreferences } from "@rilldata/web-common/features/dashboards/user-preferences";
import { getTimeComparisonParametersForComponent } from "@rilldata/web-common/lib/time/comparisons";
import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
import { ISODurationToTimePreset } from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import type {
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
  metricsExplorer.selectedTimezone = timeZone;
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
        metricsView.defaultComparisonDimension ||
        metricsView.dimensions?.[0]?.name;
      break;

    // if default_comparison is not specified it defaults to time comparison
    case MetricsViewSpecComparisonMode.COMPARISON_MODE_UNSPECIFIED:
    case MetricsViewSpecComparisonMode.COMPARISON_MODE_TIME:
      setDefaultComparisonTimeRange(
        metricsView,
        metricsExplorer,
        fullTimeRange,
      );
      break;
  }
}

function setDefaultComparisonTimeRange(
  metricsView: V1MetricsViewSpec,
  metricsExplorer: MetricsExplorerEntity,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  if (!fullTimeRange) return;

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
  metricsExplorer.showTimeComparison = true;
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
    visibleMeasureKeys: new Set(defaultMeasureNames),
    allMeasuresVisible: true,
    visibleDimensionKeys: new Set(defaultDimNames),
    allDimensionsVisible: true,
    leaderboardMeasureName: defaultMeasureNames[0],
    whereFilter: createAndExpression([]),
    havingFilter: createAndExpression([]),
    dimensionFilterExcludeMode: new Map(),
    leaderboardContextColumn: LeaderboardContextColumn.HIDDEN,
    dashboardSortType: SortType.VALUE,
    sortDirection: SortDirection.DESCENDING,

    showTimeComparison: false,
    dimensionSearchText: "",
    temporaryFilterName: null,
    pinIndex: -1,
    contextColumnWidths: { ...contextColWidthDefaults },
  };
  // set time range related stuff
  setDefaultTimeRange(metricsView, metricsExplorer, fullTimeRange);
  setDefaultComparison(metricsView, metricsExplorer, fullTimeRange);
  return metricsExplorer;
}
