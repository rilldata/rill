import { getSnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze.ts";
import {
  AlertNotifyRowsLimitDefault,
  type AlertFormValues,
} from "@rilldata/web-common/features/alerts/form-utils.ts";
import { getEmptyMeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import { getInitialScheduleFormValues } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
import {
  V1ExportFormat,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

// showEmbedOwnerEmail is true if the form shows the field for the email of the embed user who creates the alert.
// See showEmbedOwnerEmail in AlertFormDataWrapper.
export function getNewAlertInitialFormValues(
  metricsViewName: string,
  exploreName: string,
  exploreState: Partial<ExploreState>,
  userEmail: string,
  showEmbedOwnerEmail = false,
): AlertFormValues {
  // Use comparison dimension only when in TDD view (where it's visually relevant).
  // Otherwise use the expanded dimension table dimension.
  const dimension = exploreState.tdd?.expandedMeasureName
    ? (exploreState.selectedComparisonDimension ?? "")
    : (exploreState.selectedDimensionName ?? "");

  // Ephemeral measures can back an alert: the saved query embeds the
  // expression, so the name resolves without the dashboard state.
  const measure =
    exploreState.tdd?.expandedMeasureName ||
    exploreState.leaderboardSortByMeasureName ||
    "";
  const criteriaMeasure = exploreState.leaderboardSortByMeasureName ?? "";

  return {
    name: "",
    measure,
    splitByDimension: dimension,
    notifyRowsLimit: String(AlertNotifyRowsLimitDefault),
    evaluationInterval: "",
    criteria: [
      {
        ...getEmptyMeasureFilterEntry(),
        measure: criteriaMeasure,
      },
    ],
    criteriaOperation: V1Operation.OPERATION_AND,
    snooze: getSnoozeOptions()[0].value, // Defaults to `Off`

    refreshWhenDataRefreshes: true,
    ...getInitialScheduleFormValues(),
    enableSlackNotification: false,
    slackChannels: [""],
    slackUsers: userEmail ? [userEmail, ""] : [""],
    enableEmailNotification: true,
    // The creator's email is in its own field if it's shown, so it isn't also added to the recipients.
    emailRecipients: showEmbedOwnerEmail || !userEmail ? [""] : [userEmail, ""],
    // In an embed, only the creator's email gets an open link to the embedder's app.
    // The embed's `user_email` is used if set; otherwise the creator can enter it.
    embedOwnerEmail: showEmbedOwnerEmail ? userEmail : "",
    exportAttach: false,
    exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV,
    exportLimit: "",
    exportIncludeHeader: false,

    metricsViewName,
    exploreName: exploreName,
    ephemeralMeasures: exploreState.ephemeralMeasures,
  };
}
