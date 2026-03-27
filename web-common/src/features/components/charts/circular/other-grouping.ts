import type { V1MetricsViewAggregationResponseDataItem } from "@rilldata/web-common/runtime-client";

export const OTHER_SLICE_LABEL = "Other";
export const OTHER_SLICE_COLOR_LIGHT = "#d1d5db";
export const OTHER_SLICE_COLOR_DARK = "#4b5563";

const MAX_VISIBLE_SLICES = 10;
const OTHER_TARGET_PERCENT = 0.2;
const OTHER_TOOLTIP_TOP_N = 5;

export interface OtherGroupResult {
  visibleData: V1MetricsViewAggregationResponseDataItem[];
  otherItems: V1MetricsViewAggregationResponseDataItem[];
  total: number;
  hasOther: boolean;
}

/**
 * Determines how many slices to show individually vs. group into "Other".
 *
 * Algorithm:
 * - Data is assumed pre-sorted descending by measure.
 * - Walk from the largest slice downward, accumulating their share.
 * - Stop adding individual slices when:
 *     a) we've reached MAX_VISIBLE_SLICES, OR
 *     b) the remaining slices account for ≤ OTHER_TARGET_PERCENT of total
 *        and we already have at least 2 visible slices.
 * - If only 0-1 items would be grouped, don't create "Other" at all.
 */
export function computeOtherGrouping(
  data: V1MetricsViewAggregationResponseDataItem[],
  measureField: string,
  colorField: string,
  options: { limit?: number; showOther?: boolean },
): OtherGroupResult {
  const total = data.reduce(
    (sum, d) => sum + (Number(d[measureField]) || 0),
    0,
  );

  if (options.showOther === false || data.length <= 1) {
    return { visibleData: data, otherItems: [], total, hasOther: false };
  }

  const sorted = [...data].sort(
    (a, b) => (Number(b[measureField]) || 0) - (Number(a[measureField]) || 0),
  );

  let cutoff: number;

  if (options.limit !== undefined) {
    cutoff = Math.max(1, Math.min(options.limit, sorted.length));
  } else {
    cutoff = computeDynamicCutoff(sorted, measureField, total);
  }

  if (cutoff >= sorted.length || sorted.length - cutoff <= 1) {
    return { visibleData: sorted, otherItems: [], total, hasOther: false };
  }

  const visible = sorted.slice(0, cutoff);
  const others = sorted.slice(cutoff);

  const otherValue = others.reduce(
    (sum, d) => sum + (Number(d[measureField]) || 0),
    0,
  );

  const otherRow: V1MetricsViewAggregationResponseDataItem = {
    [colorField]: OTHER_SLICE_LABEL,
    [measureField]: otherValue,
  };

  return {
    visibleData: [...visible, otherRow],
    otherItems: others,
    total,
    hasOther: true,
  };
}

function computeDynamicCutoff(
  sorted: V1MetricsViewAggregationResponseDataItem[],
  measureField: string,
  total: number,
): number {
  if (total === 0) return sorted.length;

  let accumulated = 0;

  for (let i = 0; i < sorted.length && i < MAX_VISIBLE_SLICES; i++) {
    accumulated += Number(sorted[i][measureField]) || 0;
    const remaining = total - accumulated;
    const remainingPercent = remaining / total;

    if (remainingPercent <= OTHER_TARGET_PERCENT && i >= 1) {
      return i + 1;
    }
  }

  return Math.min(sorted.length, MAX_VISIBLE_SLICES);
}

export interface OtherTooltipItem {
  name: string;
  value: number;
  percent: number;
}

export interface OtherTooltipData {
  items: OtherTooltipItem[];
  remainingCount: number;
  totalValue: number;
  totalPercent: number;
}

export function getOtherTooltipData(
  otherItems: V1MetricsViewAggregationResponseDataItem[],
  measureField: string,
  colorField: string,
  grandTotal: number,
): OtherTooltipData {
  const sorted = [...otherItems].sort(
    (a, b) => (Number(b[measureField]) || 0) - (Number(a[measureField]) || 0),
  );

  const topItems = sorted.slice(0, OTHER_TOOLTIP_TOP_N);
  const remaining = sorted.length - topItems.length;

  const totalValue = sorted.reduce(
    (sum, d) => sum + (Number(d[measureField]) || 0),
    0,
  );

  return {
    items: topItems.map((d) => {
      const val = Number(d[measureField]) || 0;
      return {
        name: String(d[colorField] ?? ""),
        value: val,
        percent: grandTotal > 0 ? (val / grandTotal) * 100 : 0,
      };
    }),
    remainingCount: remaining,
    totalValue,
    totalPercent: grandTotal > 0 ? (totalValue / grandTotal) * 100 : 0,
  };
}
