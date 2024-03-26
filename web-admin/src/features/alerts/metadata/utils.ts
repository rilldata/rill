import { AlertIntervalOptions } from "@rilldata/web-common/features/alerts/delivery-tab/intervals";
import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { SnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze";
import type { V1AlertSpec } from "@rilldata/web-common/runtime-client";
import { Duration } from "luxon";

export function humaniseAlertRunDuration(alert: V1AlertSpec | undefined) {
  if (!alert?.intervalsIsoDuration) return "None";
  const preset = AlertIntervalOptions.find(
    (o) => o.value === alert.intervalsIsoDuration,
  );
  if (preset) return preset.label;
  if (alert.intervalsIsoDuration in DEFAULT_TIME_RANGES) {
    return DEFAULT_TIME_RANGES[alert.intervalsIsoDuration].label;
  }
  return humaniseISODuration(alert.intervalsIsoDuration);
}

export function humaniseAlertSnoozeOption(
  alert: V1AlertSpec | undefined,
): string {
  if (!alert?.notifySpec.renotify || !alert?.notifySpec.renotifyAfterSeconds)
    return SnoozeOptions[0].label;
  const preset = SnoozeOptions.find(
    (o) => o.value === alert.notifySpec.renotifyAfterSeconds + "",
  );
  if (preset) return preset.label;
  return (
    "Rest of " +
    Duration.fromMillis(alert.notifySpec.renotifyAfterSeconds * 1000).toHuman()
  );
}
