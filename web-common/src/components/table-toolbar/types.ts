export type SortDirection = "newest" | "oldest";

export type ViewMode = "list" | "grid";

export interface FilterOption {
  value: string;
  label: string;
}

export interface FilterGroup {
  /** Dropdown section header */
  label: string;
  /** Unique key for this filter group */
  key: string;
  /** Available options */
  options: FilterOption[];
  /** Currently selected value(s). String for single-select, string[] for multi-select. */
  selected: string | string[];
  /** Default value; when selected matches defaultValue, no chip is shown */
  defaultValue: string | string[];
  /** Allow multiple selections. Default: false (single-select radio behavior). */
  multiSelect?: boolean;
}
