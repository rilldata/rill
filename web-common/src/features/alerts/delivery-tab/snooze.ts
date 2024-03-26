import type { V1AlertSpec } from "../../../runtime-client";

const SEC = 1;
const MIN = 60 * SEC;
const HOUR = 60 * MIN;
const DAY = 24 * HOUR;
const WEEK = 7 * DAY;
const MONTH = 30 * DAY;

export const SnoozeOptions = [
  {
    value: "0",
    label: "Off",
  },
  {
    value: HOUR.toString(),
    label: "Rest of the hour",
  },
  {
    value: DAY.toString(),
    label: "Rest of the day",
  },
  {
    value: WEEK.toString(),
    label: "Rest of the week",
  },
  {
    value: MONTH.toString(),
    label: "Rest of the month",
  },
];

export function getSnoozeValueFromAlertSpec(alertSpec: V1AlertSpec): string {
  return alertSpec?.notifySpec?.renotifyAfterSeconds?.toString() || "0";
}
