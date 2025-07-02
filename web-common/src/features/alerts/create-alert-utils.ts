import type { V1User } from "@rilldata/web-admin/client";
import { SnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze.ts";
import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils.ts";
import { getEmptyMeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import {
  mapSelectedComparisonTimeRangeToV1TimeRange,
  mapSelectedTimeRangeToV1TimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import {
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  V1Operation,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";

export function getNewAlertInitialFormValues(
  metricsViewName: string,
  exploreName: string,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
  exploreState: Partial<ExploreState>,
  user: V1User | undefined,
): AlertFormValues {
  const dimension =
    exploreState.selectedComparisonDimension ||
    exploreState.selectedDimensionName ||
    "";

  const timeControlsState = getTimeControlState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
    exploreState,
  )!; // we have a check on time range beforehand. So timeControlsState cannot be undefined here.

  const timeRange = mapSelectedTimeRangeToV1TimeRange(
    timeControlsState,
    exploreState.selectedTimezone ?? "",
    exploreSpec,
  );
  const comparisonTimeRange = mapSelectedComparisonTimeRangeToV1TimeRange(
    timeControlsState,
    timeRange,
  );

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

    enableSlackNotification: false,
    slackChannels: [""],
    slackUsers: [user?.email ?? "", ""],
    enableEmailNotification: true,
    emailRecipients: [user?.email ?? "", ""],

    metricsViewName,
    exploreName: exploreName,
    whereFilter: exploreState.whereFilter ?? createAndExpression([]),
    dimensionsWithInlistFilter: exploreState.dimensionsWithInlistFilter ?? [],
    dimensionThresholdFilters: exploreState.dimensionThresholdFilters ?? [],
    timeRange: timeRange
      ? {
          ...timeRange,
          end: timeControlsState.timeEnd,
        }
      : {},
    comparisonTimeRange: comparisonTimeRange
      ? {
          ...comparisonTimeRange,
          end: timeControlsState.timeEnd,
        }
      : undefined,
  };
}
