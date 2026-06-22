import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

export enum DimensionFilterMode {
  Select = "Select",
  Contains = "Contains",
  InList = "InList",
}

// Built as a function (not a module-level constant) so the labels resolve in
// the active locale each time they are read, rather than freezing at import.
export function getDimensionFilterModeOptions() {
  return [
    {
      value: DimensionFilterMode.Select,
      label: m.dashboards_filters_mode_select(),
      description: m.dashboards_filters_mode_select_desc(),
    },
    {
      value: DimensionFilterMode.Contains,
      label: m.dashboards_filters_mode_contains(),
      description: m.dashboards_filters_mode_contains_desc(),
    },
    {
      value: DimensionFilterMode.InList,
      label: m.dashboards_filters_mode_in_list(),
      description: m.dashboards_filters_mode_in_list_desc(),
    },
  ];
}
