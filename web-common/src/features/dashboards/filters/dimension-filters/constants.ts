import * as m from "@rilldata/web-common/paraglide/messages.js";

export enum DimensionFilterMode {
  Select = "Select",
  Contains = "Contains",
  InList = "InList",
}

// Labels/descriptions use lazy getters so they resolve in the active locale at
// access time (render) rather than freezing to the locale active when this
// module loaded.
export const DimensionFilterModeOptions = [
  {
    value: DimensionFilterMode.Select,
    get label() {
      return m.filter_mode_select();
    },
    get description() {
      return m.filter_mode_select_description();
    },
  },
  {
    value: DimensionFilterMode.Contains,
    get label() {
      return m.filter_mode_contains();
    },
    get description() {
      return m.filter_mode_contains_description();
    },
  },
  {
    value: DimensionFilterMode.InList,
    get label() {
      return m.filter_mode_in_list();
    },
    get description() {
      return m.filter_mode_in_list_description();
    },
  },
];
