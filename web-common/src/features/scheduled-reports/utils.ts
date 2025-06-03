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

export type ReportValues = ReturnType<typeof getNewReportInitialFormValues>;

export function getQueryNameFromQuery(query: V1Query) {
  if (query.metricsViewAggregationRequest) {
    return "MetricsViewAggregation";
  } else {
    throw new Error(
      "Currently, only `MetricsViewAggregation` queries can be scheduled through the UI",
    );
  }
}

export function getQueryArgsJsonFromQuery(query: V1Query): string {
  if (query.metricsViewAggregationRequest) {
    return JSON.stringify(query.metricsViewAggregationRequest);
  } else {
    throw new Error(
      "Currently, only `MetricsViewAggregation` queries can be scheduled through the UI",
    );
  }
}

export function getNewReportInitialFormValues(userEmail: string | undefined) {
  return {
    title: "",
    frequency: ReportFrequency.Weekly,
    dayOfWeek: getTodaysDayOfWeek(),
    dayOfMonth: 1,
    timeOfDay: getTimeIn24FormatFromDateTime(getNextQuarterHour()),
    timeZone: getLocalIANA(),
    exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV as V1ExportFormat,
    exportLimit: "",
    exportIncludeHeader: false,
    ...extractNotification(undefined, userEmail, false),
  };
}

export function getExistingReportInitialFormValues(
  reportSpec: V1ReportSpec,
  userEmail: string | undefined,
) {
  return {
    title: reportSpec.displayName ?? "",
    frequency: getFrequencyFromCronExpression(
      reportSpec.refreshSchedule?.cron as string,
    ),
    dayOfWeek: getDayOfWeekFromCronExpression(
      reportSpec.refreshSchedule?.cron as string,
    ),
    dayOfMonth: getDayOfMonthFromCronExpression(
      reportSpec.refreshSchedule?.cron as string,
    ),
    timeOfDay: getTimeOfDayFromCronExpression(
      reportSpec.refreshSchedule?.cron as string,
    ),
    timeZone: reportSpec.refreshSchedule?.timeZone ?? getLocalIANA(),
    exportFormat:
      reportSpec?.exportFormat ?? V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED,
    exportLimit: reportSpec.exportLimit === "0" ? "" : reportSpec.exportLimit,
    exportIncludeHeader: reportSpec.exportIncludeHeader ?? false,
    ...extractNotification(reportSpec.notifiers, userEmail, true),
  };
}

export function getDashboardNameFromReport(reportSpec: V1ReportSpec): string {
  if (reportSpec.annotations?.explore) return reportSpec.annotations.explore;

  if (reportSpec.annotations?.web_open_path)
    return getExploreName(reportSpec.annotations.web_open_path);

  const queryArgsJson = JSON.parse(reportSpec.queryArgsJson!);

  return (
    queryArgsJson?.metrics_view_name ??
    queryArgsJson?.metricsViewName ??
    queryArgsJson?.metrics_view ??
    queryArgsJson?.metricsView ??
    ""
  );
}

function extractNotification(
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
