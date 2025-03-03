import { getExploreName } from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
import {
  getDayOfMonthFromCronExpression,
  getDayOfWeekFromCronExpression,
  getFrequencyFromCronExpression,
  getNextQuarterHour,
  getTimeIn24FormatFromDateTime,
  getTimeOfDayFromCronExpression,
  getTodaysDayOfWeek,
  ReportFrequency,
} from "@rilldata/web-common/features/scheduled-reports/time-utils";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import {
  V1ExportFormat,
  type V1Notifier,
  type V1Query,
  type V1ReportSpec,
} from "@rilldata/web-common/runtime-client";

export type ReportValues = ReturnType<typeof getInitialValues>;

export function getQueryNameFromQuery(query: V1Query) {
  if (query.metricsViewAggregationRequest) {
    return "MetricsViewAggregation";
  } else {
    throw new Error(
      "Currently, only `MetricsViewAggregation` queries can be scheduled through the UI",
    );
  }
}

export function getQueryArgsFromQuery(query: V1Query) {
  if (query.metricsViewAggregationRequest) {
    return query.metricsViewAggregationRequest;
  } else {
    throw new Error(
      "Currently, only `MetricsViewAggregation` queries can be scheduled through the UI",
    );
  }
}

export function getInitialValues(
  reportSpec: V1ReportSpec | undefined,
  userEmail: string | undefined,
) {
  return {
    title: reportSpec?.displayName ?? "",
    frequency: reportSpec
      ? getFrequencyFromCronExpression(
          reportSpec.refreshSchedule?.cron as string,
        )
      : ReportFrequency.Weekly,
    dayOfWeek: reportSpec
      ? getDayOfWeekFromCronExpression(
          reportSpec.refreshSchedule?.cron as string,
        )
      : getTodaysDayOfWeek(),
    dayOfMonth: reportSpec
      ? getDayOfMonthFromCronExpression(
          reportSpec.refreshSchedule?.cron as string,
        )
      : 1, // We only support first of day right now
    timeOfDay: reportSpec
      ? getTimeOfDayFromCronExpression(
          reportSpec.refreshSchedule?.cron as string,
        )
      : getTimeIn24FormatFromDateTime(getNextQuarterHour()),
    timeZone: reportSpec?.refreshSchedule?.timeZone ?? getLocalIANA(),
    exportFormat: reportSpec
      ? (reportSpec?.exportFormat ?? V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED)
      : V1ExportFormat.EXPORT_FORMAT_CSV,
    exportLimit: reportSpec
      ? reportSpec.exportLimit === "0"
        ? ""
        : reportSpec.exportLimit
      : "",
    ...extractNotificationV2(reportSpec?.notifiers, userEmail, !!reportSpec),
  };
}

export function getDashboardNameFromReport(
  reportSpec: V1ReportSpec | undefined,
): string | null {
  if (!reportSpec?.queryArgsJson) return null;
  if (reportSpec.annotations?.web_open_path)
    return getExploreName(reportSpec.annotations.web_open_path);

  const queryArgsJson = JSON.parse(reportSpec.queryArgsJson);
  return (
    queryArgsJson?.metrics_view_name ??
    queryArgsJson?.metricsViewName ??
    queryArgsJson?.metrics_view ??
    queryArgsJson?.metricsView ??
    null
  );
}

export function extractNotificationV2(
  notifiers: V1Notifier[] | undefined,
  userEmail: string | undefined,
  isEdit: boolean,
) {
  const slackNotifier = notifiers?.find((n) => n.connector === "slack");
  const slackChannels = mapAndAddEmptyEntry(
    slackNotifier?.properties?.channels as string[],
  );
  const slackUsers = mapAndAddEmptyEntry(
    slackNotifier?.properties?.users as string[],
  );

  const emailNotifier = notifiers?.find((n) => n.connector === "email");
  const emailRecipients = mapAndAddEmptyEntry(
    emailNotifier?.properties?.recipients as string[],
  );

  if (userEmail && !isEdit) {
    slackUsers.unshift(userEmail);
    emailRecipients.unshift(userEmail);
  }

  return {
    enableSlackNotification: isEdit ? !!slackNotifier : false,
    slackChannels,
    slackUsers,

    enableEmailNotification: isEdit ? !!emailNotifier : true,
    emailRecipients,
  };
}

function mapAndAddEmptyEntry(entries: string[] | undefined) {
  const finalEntries = entries ? [...entries] : [];
  finalEntries.push("");
  return finalEntries;
}
