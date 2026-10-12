import type { V1Notifier } from "@rilldata/web-common/runtime-client";

export type EmailNotifierProperties = {
  recipients: string[];
};
export type SlackNotifierProperties = {
  users: string[];
  channels: string[];
  webhooks: string[];
};
export type WebhookNotifierProperties = {
  urls: string[];
};

type NotifierPropsMap = {
  email: EmailNotifierProperties;
  slack: SlackNotifierProperties;
  webhook: WebhookNotifierProperties;
};

export function extractNotifier<Notifier extends keyof NotifierPropsMap>(
  notifiers: V1Notifier[] | undefined,
  name: Notifier,
): NotifierPropsMap[Notifier] | undefined {
  if (!notifiers) return undefined;
  const notifier = notifiers.find((n) => n.connector === name);
  if (!notifier?.properties) return undefined;
  return notifier.properties as NotifierPropsMap[Notifier];
}
