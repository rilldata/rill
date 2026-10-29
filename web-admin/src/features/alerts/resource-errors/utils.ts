import type { V1AlertOptions } from "@rilldata/web-admin/client";
import { getSnoozeValueFromAlertSpec } from "@rilldata/web-common/features/alerts/delivery-tab/snooze";
import type { V1AlertSpec } from "@rilldata/web-common/runtime-client";

export const RESOURCE_STATUS_RESOLVER = "resource_status";
export const RESOURCE_ERRORS_OPEN_PATH = "/-/status/resources?status=error";
export const DEFAULT_CHECK_CRON = "*/15 * * * *";

export type ResourceErrorAlertFormValues = {
  name: string;
  // Watched resources as "kind/name"; empty means all resources.
  resources: string[];
  refreshCron: string;
  emailRecipients: string[];
  slackChannels: string[];
  slackUsers: string[];
  snooze: string;
};

export function isResourceErrorAlert(spec: V1AlertSpec | undefined): boolean {
  return spec?.annotations?.admin_system === "true";
}

export function resourceKey(kind: string, name: string): string {
  return `${kind}/${name}`;
}

export function getAlertOptionsFromFormValues(
  values: ResourceErrorAlertFormValues,
): V1AlertOptions {
  const snoozeSeconds = Number(values.snooze) || 0;
  return {
    displayName: values.name.trim(),
    refreshCron: values.refreshCron,
    resolver: RESOURCE_STATUS_RESOLVER,
    resolverProperties: values.resources.length
      ? { where_error: true, where_resource: values.resources }
      : { where_error: true },
    emailRecipients: values.emailRecipients,
    slackChannels: values.slackChannels.filter(Boolean),
    slackUsers: values.slackUsers.filter(Boolean),
    renotify: snoozeSeconds > 0,
    renotifyAfterSeconds: snoozeSeconds,
    webOpenPath: RESOURCE_ERRORS_OPEN_PATH,
  };
}

export function getFormValuesFromAlertSpec(
  spec: V1AlertSpec,
): ResourceErrorAlertFormValues {
  const notifierProp = (connector: string, prop: string) => [
    ...((spec.notifiers?.find((n) => n.connector === connector)?.properties?.[
      prop
    ] as string[] | undefined) ?? []),
  ];

  return {
    name: spec.displayName ?? "",
    resources: [
      ...((spec.resolverProperties?.where_resource as string[] | undefined) ??
        []),
    ],
    refreshCron: spec.refreshSchedule?.cron || DEFAULT_CHECK_CRON,
    emailRecipients: notifierProp("email", "recipients"),
    slackChannels: notifierProp("slack", "channels"),
    slackUsers: notifierProp("slack", "users"),
    snooze: getSnoozeValueFromAlertSpec(spec),
  };
}
