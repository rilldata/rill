import { getSnoozeValueFromAlertSpec } from "@rilldata/web-common/features/alerts/delivery-tab/snooze.ts";
import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import {
  getEmptyMeasureFilterEntry,
  mapExprToMeasureFilter,
  type MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils.ts";
import { getExistingScheduleFormValues } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
import {
  type V1AlertSpec,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationRequest,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export type AlertFormValuesSubset = Pick<
  AlertFormValues,
  | "metricsViewName"
  | "measure"
  | "splitByDimension"
  | "criteria"
  | "criteriaOperation"
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

export function getExistingAlertInitialFormValues(
  alertSpec: V1AlertSpec,
  metricsViewName: string,
): AlertFormValues {
  const queryArgsJson = JSON.parse(
    (alertSpec.resolverProperties?.query_args_json ??
      alertSpec.queryArgsJson) as string,
  ) as V1MetricsViewAggregationRequest;

  const exploreName = getExploreName(
    alertSpec.annotations?.web_open_path ?? "",
  );

  return {
    name: alertSpec.displayName as string,
    exploreName: exploreName ?? metricsViewName,
    snooze: getSnoozeValueFromAlertSpec(alertSpec),
    evaluationInterval: alertSpec.intervalsIsoDuration ?? "",
    refreshWhenDataRefreshes: !alertSpec.refreshSchedule?.cron,
    ...getExistingScheduleFormValues(alertSpec.refreshSchedule),
    ...extractAlertNotification(alertSpec),
    ...extractAlertFormValues(queryArgsJson),
  };
}
