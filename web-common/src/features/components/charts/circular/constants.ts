// Visible label for the synthetic remainder slice.
export const OTHER_VALUE = "Other";

// Domain-values keys used to ferry derived values from the provider to the
// spec generator via `ChartDataResult.domainValues`.
export const TOTAL_DOMAIN_KEY = "total";
export const OTHER_VALUE_DOMAIN_KEY = "__other_value";

// Synthetic field added by the percent-of-total transform.
export const PERCENT_OF_TOTAL_FIELD = "__percent_of_total";

// Synthetic field used to pin the "Other" slice to the end of the arc order
// regardless of how rows sort by measure or category.
export const OTHER_SORT_KEY_FIELD = "__other_sort_key";

// Tooltip column title for the percent-of-total entry.
export const PERCENT_OF_TOTAL_TITLE = "% of total";
