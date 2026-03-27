export const OTHER_LABEL = "Other";
export const OTHER_FLAG_FIELD = "__isOther";
const MIN_SLICES_FOR_GROUPING = 6;
const DEFAULT_MAX_OTHER_PERCENT = 0.2;
const DEFAULT_HARD_CAP = 10;
const MAX_OTHER_TOOLTIP_ITEMS = 5;

export interface OtherSliceItem {
  label: string;
  value: number;
}

export interface OtherGroupingResult {
  /** Data rows for Vega-Lite; "Other" row has __isOther: true */
  visibleData: Record<string, unknown>[];
  /** Items grouped into "Other", sorted by value desc; null if no "Other" slice */
  otherItems: OtherSliceItem[] | null;
  /** Total value across ALL items (for percentage calculation) */
  total: number;
}

/**
 * Groups small pie chart slices into an "Other" aggregate.
 *
 * @param data - Raw data rows from the metrics view query
 * @param colorField - The dimension field name used for slice labels
 * @param measureField - The measure field name used for slice values
 * @param options.explicitLimit - If set by the editor in YAML, bypasses dynamic algorithm
 * @param options.showOther - If false, truncate at limit with no "Other" (default: true)
 * @param options.maxOtherPercent - Max fraction of total for "Other" before adding more slices (default: 0.2)
 * @param options.hardCap - Max visible slices before forcing "Other" (default: 10)
 */
export function computeVisibleSlices(
  data: Record<string, unknown>[],
  colorField: string,
  measureField: string,
  options: {
    explicitLimit?: number;
    showOther?: boolean;
    maxOtherPercent?: number;
    hardCap?: number;
  } = {},
): OtherGroupingResult {
  const {
    showOther = true,
    maxOtherPercent = DEFAULT_MAX_OTHER_PERCENT,
    hardCap = DEFAULT_HARD_CAP,
  } = options;

  // Sort by measure value descending
  const sorted = [...data].sort((a, b) => {
    const aVal = Number(a[measureField]) || 0;
    const bVal = Number(b[measureField]) || 0;
    return bVal - aVal;
  });

  const total = sorted.reduce(
    (sum, d) => sum + (Number(d[measureField]) || 0),
    0,
  );

  // If too few items or showOther is false with no explicit limit, return all
  if (sorted.length <= MIN_SLICES_FOR_GROUPING || !showOther) {
    const limit = options.explicitLimit;
    if (!showOther && limit !== undefined && limit < sorted.length) {
      // Truncate without "Other"
      return {
        visibleData: sorted.slice(0, limit),
        otherItems: null,
        total,
      };
    }
    return {
      visibleData: sorted,
      otherItems: null,
      total,
    };
  }

  // Determine how many slices to show
  let visibleCount: number;

  if (options.explicitLimit !== undefined) {
    // Editor set an explicit limit; use it directly
    visibleCount = Math.min(options.explicitLimit, sorted.length);
  } else {
    // Dynamic threshold algorithm
    visibleCount = 0;
    let visibleSum = 0;

    for (const item of sorted) {
      if (visibleCount >= hardCap) break;
      visibleCount++;
      visibleSum += Number(item[measureField]) || 0;
      const remaining = total - visibleSum;
      if (total > 0 && remaining / total <= maxOtherPercent) break;
    }
  }

  // If all items are visible, no "Other" needed
  if (visibleCount >= sorted.length) {
    return {
      visibleData: sorted,
      otherItems: null,
      total,
    };
  }

  const visibleRows = sorted.slice(0, visibleCount);
  const otherRows = sorted.slice(visibleCount);

  const otherItems: OtherSliceItem[] = otherRows.map((row) => ({
    label: String(row[colorField] ?? ""),
    value: Number(row[measureField]) || 0,
  }));

  const otherValue = otherItems.reduce((sum, item) => sum + item.value, 0);

  // Create the "Other" data row
  const otherDataRow: Record<string, unknown> = {
    [colorField]: OTHER_LABEL,
    [measureField]: otherValue,
    [OTHER_FLAG_FIELD]: true,
  };

  return {
    visibleData: [...visibleRows, otherDataRow],
    otherItems,
    total,
  };
}

export { MAX_OTHER_TOOLTIP_ITEMS };
