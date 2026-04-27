// Visible label for the synthetic remainder slice.
export const OTHER_VALUE = "Other";

// Domain-values keys used to ferry derived values from the provider to the
// spec generator via `ChartDataResult.domainValues`.
export const TOTAL_DOMAIN_KEY = "total";
export const OTHER_VALUE_DOMAIN_KEY = "__other_value";

// Synthetic field added by the percent-of-total transform.
export const PERCENT_OF_TOTAL_FIELD = "__percent_of_total";

// Tooltip column title for the percent-of-total entry.
export const PERCENT_OF_TOTAL_TITLE = "% of total";

export type LabelsFormat = "percent" | "value";
export const DEFAULT_LABELS_FORMAT: LabelsFormat = "percent";
export const DEFAULT_LABELS_THRESHOLD = 5;

export type LabelsConfig = {
  show?: boolean;
  format?: LabelsFormat;
  // Hide labels for slices below this percent of total.
  threshold?: number;
};
