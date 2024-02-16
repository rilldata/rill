import { SnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze";
import type { V1AlertSpec } from "@rilldata/web-common/runtime-client";
import { Duration } from "luxon";

export function humaniseAlertSnoozeOption(
  alert: V1AlertSpec | undefined,
): string {
  if (!alert?.emailRenotify || !alert.emailRenotifyAfterSeconds)
    return SnoozeOptions[0].label;
  const preset = SnoozeOptions.find(
    (o) => o.value === alert.emailRenotifyAfterSeconds + "",
  );
  if (preset) return preset.label;
  return (
    "Rest of " +
    Duration.fromMillis(alert.emailRenotifyAfterSeconds * 1000).toHuman()
  );
}
