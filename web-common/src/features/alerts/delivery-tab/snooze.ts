import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { V1AlertSpec } from "../../../runtime-client";

const SEC = 1;
const MIN = 60 * SEC;
const HOUR = 60 * MIN;
const DAY = 24 * HOUR;
const WEEK = 7 * DAY;
const MONTH = 30 * DAY;

export function getSnoozeOptions() {
  return [
    {
      value: "0",
      label: m.snooze_off(),
    },
    {
      value: HOUR.toString(),
      label: "1 hour",
    },
    {
      value: DAY.toString(),
      label: "1 day",
    },
    {
      value: WEEK.toString(),
      label: "1 week",
    },
    {
      value: MONTH.toString(),
      label: "1 month",
    },
  ];
}

export function getSnoozeValueFromAlertSpec(alertSpec: V1AlertSpec): string {
  return alertSpec?.renotifyAfterSeconds?.toString() || "0";
}
