import type { VLTooltipFormatter } from "@rilldata/web-common/components/vega/types";
import {
  MAX_OTHER_TOOLTIP_ITEMS,
  OTHER_FLAG_FIELD,
  OTHER_LABEL,
  type OtherSliceItem,
} from "./other-grouping";

/**
 * Creates a tooltip formatter for pie/donut charts that shows
 * slice name, formatted value, and percentage. For the "Other" slice,
 * shows a mini-leaderboard breakdown of grouped items.
 */
export function createPieTooltipFormatter(options: {
  colorField: string;
  measureField: string;
  total: number;
  otherItems: OtherSliceItem[] | null;
  /** Color map from dimension value → resolved CSS color string */
  colorMap: Map<string, string>;
  /** Format a measure value for display (e.g., "$12,450") */
  formatValue: (value: number) => string;
  /** Resolved CSS color for the muted token (for "Other" dot) */
  mutedColor: string;
}): VLTooltipFormatter {
  const {
    colorField,
    measureField,
    total,
    otherItems,
    colorMap,
    formatValue,
    mutedColor,
  } = options;

  return (value: unknown): string => {
    if (!value || typeof value !== "object") return "";
    const datum = value as Record<string, unknown>;

    const isOther = datum[OTHER_FLAG_FIELD] === true;
    const label = String(datum[colorField] ?? "");
    const measureValue = Number(datum[measureField]) || 0;
    const pct = total > 0 ? ((measureValue / total) * 100).toFixed(1) : "0.0";

    if (isOther && otherItems) {
      return renderOtherTooltip(
        measureValue,
        pct,
        otherItems,
        total,
        formatValue,
        mutedColor,
      );
    }

    const color = colorMap.get(label) || "#888";
    return renderSliceTooltip(label, measureValue, pct, color, formatValue);
  };
}

function renderSliceTooltip(
  label: string,
  value: number,
  pct: string,
  color: string,
  formatValue: (v: number) => string,
): string {
  const dot = `<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${color};margin-right:6px;vertical-align:middle;"></span>`;
  const formatted = formatValue(value);
  return `<div style="display:flex;align-items:center;gap:4px;white-space:nowrap;">${dot}<span style="font-weight:500;">${escapeHtml(label)}</span><span style="opacity:0.7;"> · ${formatted} · ${pct}%</span></div>`;
}

function renderOtherTooltip(
  otherTotal: number,
  pct: string,
  items: OtherSliceItem[],
  grandTotal: number,
  formatValue: (v: number) => string,
  mutedColor: string,
): string {
  const dot = `<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${mutedColor};margin-right:6px;vertical-align:middle;"></span>`;
  const formatted = formatValue(otherTotal);

  // Header
  let html = `<div style="display:flex;align-items:center;gap:4px;white-space:nowrap;margin-bottom:6px;">${dot}<span style="font-weight:500;">${OTHER_LABEL}</span><span style="opacity:0.7;"> · ${formatted} · ${pct}%</span></div>`;

  // Divider
  html += `<div style="border-bottom:1px solid var(--border, #e5e7eb);margin-bottom:6px;"></div>`;

  // Breakdown rows (max 5)
  const visibleItems = items.slice(0, MAX_OTHER_TOOLTIP_ITEMS);
  html += `<table style="width:100%;border-collapse:collapse;">`;
  for (const item of visibleItems) {
    const itemPct =
      grandTotal > 0 ? ((item.value / grandTotal) * 100).toFixed(1) : "0.0";
    const itemFormatted = formatValue(item.value);
    html += `<tr>`;
    html += `<td style="text-align:left;padding:1px 8px 1px 0;max-width:160px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">${escapeHtml(item.label)}</td>`;
    html += `<td style="text-align:right;padding:1px 4px;white-space:nowrap;font-variant-numeric:tabular-nums;">${itemFormatted}</td>`;
    html += `<td style="text-align:right;padding:1px 0 1px 4px;white-space:nowrap;font-variant-numeric:tabular-nums;opacity:0.7;">${itemPct}%</td>`;
    html += `</tr>`;
  }
  html += `</table>`;

  // Footer "and N more"
  const remaining = items.length - MAX_OTHER_TOOLTIP_ITEMS;
  if (remaining > 0) {
    html += `<div style="font-size:11px;opacity:0.6;font-style:italic;padding-top:4px;">and ${remaining} more</div>`;
  }

  return html;
}

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
