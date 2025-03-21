export enum DimensionFilterMode {
  Select = "Select",
  Contains = "Contains",
  InList = "InList",
}

export const DimensionFilterModeOptions = [
  {
    value: DimensionFilterMode.Select,
    label: "Select",
    description: "Manually select values for this filter",
  },
  {
    value: DimensionFilterMode.Contains,
    label: "Contains",
    description: "Create a dynamic filter based on a search term",
  },
  {
    value: DimensionFilterMode.InList,
    label: "In List",
    description: "Create a filter based on a list of values",
  },
];
