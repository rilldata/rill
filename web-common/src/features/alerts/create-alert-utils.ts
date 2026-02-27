import type { V1User } from "@rilldata/web-admin/client";
import { SnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze.ts";
import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils.ts";
import { getEmptyMeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import { ExploreMetricsViewMetadata } from "@rilldata/web-common/features/dashboards/stores/ExploreMetricsViewMetadata.ts";
import { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import { getInitialScheduleFormValues } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
import { V1Operation } from "@rilldata/web-common/runtime-client";

export function getNewAlertInitialFormValues(
  metricsViewName: string,
  exploreName: string,
  exploreState: Partial<ExploreState>,
  user: V1User | undefined,
): AlertFormValues {
  // Use comparison dimension only when in TDD view (where it's visually relevant).
  // Otherwise use the expanded dimension table dimension.
  const dimension = exploreState.tdd?.expandedMeasureName
    ? (exploreState.selectedComparisonDimension ?? "")
    : (exploreState.selectedDimensionName ?? "");

  return {
    name: "",
    measure:
      exploreState.tdd?.expandedMeasureName ||
      exploreState.leaderboardSortByMeasureName ||
      "",
    splitByDimension: dimension,
    evaluationInterval: "",
    criteria: [
      {
        ...getEmptyMeasureFilterEntry(),
        measure: exploreState.leaderboardSortByMeasureName ?? "",
      },
    ],
    criteriaOperation: V1Operation.OPERATION_AND,
    snooze: SnoozeOptions[0].value, // Defaults to `Off`

    refreshWhenDataRefreshes: true,
    ...getInitialScheduleFormValues(),
    enableSlackNotification: false,
    slackChannels: [""],
    slackUsers: [user?.email ?? "", ""],
    enableEmailNotification: true,
    emailRecipients: [user?.email ?? "", ""],

    metricsViewName,
    exploreName: exploreName,
  };
}

export function getNewAlertInitialFiltersFormValues(
  instanceId: string,
  metricsViewName: string,
  exploreName: string,
  exploreState: Partial<ExploreState>,
) {
  const metricsViewMetadata = new ExploreMetricsViewMetadata(
    instanceId,
    metricsViewName,
    exploreName,
  );
  const filters = new Filters(metricsViewMetadata, {
    whereFilter: exploreState.whereFilter ?? createAndExpression([]),
    dimensionsWithInlistFilter: exploreState.dimensionsWithInlistFilter ?? [],
    dimensionThresholdFilters: exploreState.dimensionThresholdFilters ?? [],
    dimensionFilterExcludeMode:
      exploreState.dimensionFilterExcludeMode ?? new Map<string, boolean>(),
  });
  const timeControls = new TimeControls(metricsViewMetadata, {
    selectedTimeRange: exploreState.selectedTimeRange,
    selectedComparisonTimeRange: exploreState.selectedComparisonTimeRange,
    showTimeComparison: exploreState.showTimeComparison ?? false,
    selectedTimezone: exploreState.selectedTimezone ?? "UTC",
  });
  return { filters, timeControls };
}
