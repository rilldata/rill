import { getExploreName } from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
import type { AlertNotificationValues } from "@rilldata/web-common/features/alerts/extract-alert-form-values";
import type {
  V1AlertSpec,
  V1Notifier,
  V1ReportSpec,
} from "@rilldata/web-common/runtime-client";

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

export function extractNotification(
  notifiers: V1Notifier[] | undefined,
  userEmail: string | undefined,
  isEdit: boolean,
): AlertNotificationValues {
  const slackNotifier = notifiers?.find((n) => n.connector === "slack");
  const slackChannels = mapAndAddEmptyEntry(
    slackNotifier?.properties?.channels as string[] | undefined,
    "channel",
  );
  const slackUsers = mapAndAddEmptyEntry(
    slackNotifier?.properties?.users as string[] | undefined,
    "email",
  );

  const emailNotifier = notifiers?.find((n) => n.connector === "email");
  const emailRecipients = mapAndAddEmptyEntry(
    emailNotifier?.properties?.recipients as string[] | undefined,
    "email",
  );

  if (userEmail && !isEdit) {
    slackUsers.push({
      email: userEmail,
    });
    emailRecipients.push({
      email: userEmail,
    });
  }

  return {
    enableSlackNotification: isEdit ? !!slackNotifier : true,
    slackChannels,
    slackUsers,

    enableEmailNotification: isEdit ? !!emailNotifier : true,
    emailRecipients,
  };
}

function mapAndAddEmptyEntry<K extends string>(
  entries: string[] | undefined,
  key: K,
) {
  const mappedEntries = entries?.map((e) => ({ [key]: e })) ?? [];
  mappedEntries.push({ [key]: "" });
  return mappedEntries as {
    [KEY in K]: string;
  }[];
}
