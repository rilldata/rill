import * as m from "@rilldata/web-common/paraglide/messages.js";

export function getAlertIntervalOptions() {
  return [
    {
      label: m.interval_none(),
      value: "",
    },
    {
      label: m.interval_hour(),
      value: "PT1H",
    },
    {
      label: m.interval_day(),
      value: "P1D",
    },
    {
      label: m.interval_week(),
      value: "P1W",
    },
  ];
}
