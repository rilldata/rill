import type { V1Notifier } from "@rilldata/web-common/runtime-client";

export type EmailNotifierProperties = {
  recipients: string[];
};
export type SlackNotifierProperties = {
  users: string[];
  channels: string[];
  webhooks: string[];
};

type NotifierPropsMap = {
  email: EmailNotifierProperties;
  slack: SlackNotifierProperties;
};

export function extractNotifier<Notifier extends keyof NotifierPropsMap>(
  notifiers: V1Notifier[] | undefined,
  name: Notifier,
): NotifierPropsMap[Notifier] | undefined {
  if (!notifiers) return undefined;
  const notifier = notifiers.find((n) => n.connector === name);
  if (!notifier?.properties) return undefined;

  switch (name) {
    case "email":
      return {
        recipients: notifier.properties.recipients ?? [],
      } as NotifierPropsMap[Notifier];

    case "slack":
      return {
        users: notifier.properties.recipients ?? [],
        channels: notifier.properties.recipients ?? [],
        webhooks: notifier.properties.recipients ?? [],
      } as NotifierPropsMap[Notifier];
  }

  return notifier.properties as NotifierPropsMap[Notifier];
}
