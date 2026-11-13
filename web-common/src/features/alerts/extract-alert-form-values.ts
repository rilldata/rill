import { getSnoozeValueFromAlertSpec } from "@rilldata/web-common/features/alerts/delivery-tab/snooze.ts";
import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import { ephemeralDefsFromRequestMeasures } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import {
  getEmptyMeasureFilterEntry,
  mapExprToMeasureFilter,
  type MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils.ts";
import { getExistingScheduleFormValues } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import {
  type V1AlertSpec,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationRequest,
  V1ExportFormat,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export type AlertFormValuesSubset = Pick<
  AlertFormValues,
  | "metricsViewName"
  | "measure"
  | "splitByDimension"
  | "criteria"
  | "criteriaOperation"
  | "ephemeralMeasures"
>;

export function extractAlertFormValues(
  queryArgs: V1MetricsViewAggregationRequest,
): AlertFormValuesSubset {
  if (!queryArgs) return {} as AlertFormValuesSubset;

  const measures = queryArgs.measures as V1MetricsViewAggregationMeasure[];
  const dimensions =
    queryArgs.dimensions as V1MetricsViewAggregationDimension[];

  return {
    measure: measures[0]?.name ?? "",
    splitByDimension: dimensions[0]?.name ?? "",

    criteria: (queryArgs.having?.cond?.exprs?.map(
      mapExprToMeasureFilter,
    ) as MeasureFilterEntry[]) ?? [getEmptyMeasureFilterEntry()],
    criteriaOperation: queryArgs.having?.cond?.op ?? V1Operation.OPERATION_AND,

    // These are not part of the form, but are used to track the state of the form
    metricsViewName: queryArgs.metricsView as string,
    // An edited alert carries its ephemeral definitions in the saved query.
    ephemeralMeasures: ephemeralDefsFromRequestMeasures(measures),
  };
}

export type AlertNotificationValues = Pick<
  AlertFormValues,
  | "enableSlackNotification"
  | "slackChannels"
  | "slackUsers"
  | "enableEmailNotification"
  | "emailRecipients"
>;

export function extractAlertNotification(
  alertSpec: V1AlertSpec,
): AlertNotificationValues {
  const slackNotifier = alertSpec.notifiers?.find(
    (n) => n.connector === "slack",
  );
  const slackChannels = slackNotifier?.properties?.channels
    ? [...(slackNotifier.properties.channels as string[])]
    : [];
  slackChannels.push("");
  const slackUsers = slackNotifier?.properties?.users
    ? [...(slackNotifier.properties.users as string[])]
    : [];
  slackUsers.push("");

  const emailNotifier = alertSpec.notifiers?.find(
    (n) => n.connector === "email",
  );
  const emailRecipients = emailNotifier?.properties?.recipients
    ? [...(emailNotifier.properties.recipients as string[])]
    : [];
  emailRecipients.push("");

  return {
    enableSlackNotification: !!slackNotifier,
    slackChannels,
    slackUsers,

    enableEmailNotification: !!emailNotifier,
    emailRecipients,
  };
}

// userEmail is the email of the current user, and showEmbedOwnerEmail is true if the form shows the field for the creator's email.
// See showEmbedOwnerEmail in AlertFormDataWrapper.
export function getExistingAlertInitialFormValues(
  alertSpec: V1AlertSpec,
  metricsViewName: string,
  userEmail = "",
  showEmbedOwnerEmail = false,
): AlertFormValues {
  const queryArgsJson = JSON.parse(
    (alertSpec.resolverProperties?.query_args_json ??
      alertSpec.queryArgsJson) as string,
  ) as V1MetricsViewAggregationRequest;

  const exploreName = getExploreName(
    alertSpec.annotations?.web_open_path ?? "",
  );

  // The creator's saved email is in its own field rather than in the recipients,
  // since the admin service always keeps it as a recipient (the creator can't unsubscribe).
  const savedOwnerEmail = alertSpec.annotations?.admin_embed_owner_email ?? "";
  const notification = extractAlertNotification(alertSpec);
  if (savedOwnerEmail) {
    notification.emailRecipients = notification.emailRecipients.filter(
      (r) => r.toLowerCase() !== savedOwnerEmail.toLowerCase(),
    );
  }

  return {
    name: alertSpec.displayName as string,
    exploreName: exploreName ?? metricsViewName,
    snooze: getSnoozeValueFromAlertSpec(alertSpec),
    evaluationInterval: alertSpec.intervalsIsoDuration ?? "",
    refreshWhenDataRefreshes: !alertSpec.refreshSchedule?.cron,
    ...getExistingScheduleFormValues(alertSpec.refreshSchedule),
    ...notification,
    // In an embed, the embed's `user_email` takes precedence over the saved email, the same way the admin service applies it.
    // Outside of an embed, userEmail is the editor's email, so only the saved email applies.
    embedOwnerEmail: showEmbedOwnerEmail
      ? (EmbedStore.isEmbedded() ? userEmail : "") || savedOwnerEmail
      : "",
    exportAttach: alertSpec.exportAttach ?? false,
    // Alerts that don't attach an export have no format, so default to CSV for when attaching is enabled
    exportFormat:
      alertSpec.exportFormat &&
      alertSpec.exportFormat !== V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED
        ? alertSpec.exportFormat
        : V1ExportFormat.EXPORT_FORMAT_CSV,
    exportLimit:
      !alertSpec.exportLimit || alertSpec.exportLimit === "0"
        ? ""
        : alertSpec.exportLimit,
    exportIncludeHeader: alertSpec.exportIncludeHeader ?? false,
    ...extractAlertFormValues(queryArgsJson),
  };
}
