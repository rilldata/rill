import { AlertIntervalOptions } from "@rilldata/web-common/features/alerts/data-tab/intervals";
import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import type { V1AlertSpec } from "@rilldata/web-common/runtime-client";

export function humaniseAlertRunDuration(alert: V1AlertSpec | undefined) {
  if (!alert?.intervalsIsoDuration) return "None";
  const preset = AlertIntervalOptions.find(
    (o) => o.value === alert.intervalsIsoDuration,
  );
  if (preset) return preset.label;
  return humaniseISODuration(alert.intervalsIsoDuration);
}
