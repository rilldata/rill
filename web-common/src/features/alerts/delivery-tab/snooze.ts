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
    value: HOUR + "",
    label: "Rest of the hour",
  },
  {
    value: DAY + "",
    label: "Rest of the day",
  },
  {
    value: WEEK + "",
    label: "Rest of the week",
  },
  {
    value: MONTH + "",
    label: "Rest of the month",
  },
];
