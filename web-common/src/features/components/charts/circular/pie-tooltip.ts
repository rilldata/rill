import type { VLTooltipFormatter } from "@rilldata/web-common/components/vega/types";
import type { V1MetricsViewAggregationResponseDataItem } from "@rilldata/web-common/runtime-client";
import type { ColorMapping } from "../types";
import {
  OTHER_SLICE_COLOR_DARK,
  OTHER_SLICE_COLOR_LIGHT,
  OTHER_SLICE_LABEL,
  getOtherTooltipData,
} from "./other-grouping";

export interface PieTooltipConfig {
  colorField: string;
  measureField: string;
  colorFieldLabel: string;
  measureFieldLabel: string;
  otherItems: V1MetricsViewAggregationResponseDataItem[];
  grandTotal: number;
  colorMapping: ColorMapping;
  isDarkMode: boolean;
  formatValue: (val: number) => string;
}

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

export function createPieTooltipFormatter(
  config: PieTooltipConfig,
): VLTooltipFormatter {
  const colorMap = new Map<string, string>(
    config.colorMapping.map((m) => [m.value, m.color]),
  );

  return (value: unknown): string => {
    if (!value || typeof value !== "object") return "";

    const record = value as Record<string, unknown>;
    const sliceName = String(record[config.colorField] ?? "");
    const sliceValue = Number(record[config.measureField] ?? 0);
    const percentage = Number(record["__percentage"] ?? 0);

    if (sliceName === OTHER_SLICE_LABEL) {
      return buildOtherTooltip(config, sliceValue);
    }

    return buildNamedSliceTooltip(
      sliceName,
      sliceValue,
      percentage,
      colorMap.get(sliceName),
      config,
    );
  };
}

function buildNamedSliceTooltip(
  name: string,
  value: number,
  percentage: number,
  color: string | undefined,
  config: PieTooltipConfig,
): string {
  const formattedValue = config.formatValue(value);
  const pctStr = percentage.toFixed(1) + "%";
  const colorDot = color
    ? `<svg class="key-color"><circle cx="6" cy="6" r="6" style="fill:${color};"/></svg>`
    : "";

  return `<table><tbody>
    <tr><td class="key">${colorDot}<span>${escapeHtml(name)}</span></td><td class="value">${escapeHtml(formattedValue)}</td><td class="value" style="opacity:0.7">${pctStr}</td></tr>
  </tbody></table>`;
}

function buildOtherTooltip(
  config: PieTooltipConfig,
  otherValue: number,
): string {
  const data = getOtherTooltipData(
    config.otherItems,
    config.measureField,
    config.colorField,
    config.grandTotal,
  );

  const otherColor = config.isDarkMode
    ? OTHER_SLICE_COLOR_DARK
    : OTHER_SLICE_COLOR_LIGHT;
  const otherPct = data.totalPercent.toFixed(1) + "%";
  const formattedOtherValue = config.formatValue(otherValue);

  const rows: string[] = [];

  rows.push(
    `<tr><td colspan="3" style="font-weight:600;padding-bottom:4px"><svg class="key-color"><circle cx="6" cy="6" r="6" style="fill:${otherColor};"/></svg>${escapeHtml(OTHER_SLICE_LABEL)} &mdash; ${escapeHtml(formattedOtherValue)} (${otherPct})</td></tr>`,
  );

  for (const item of data.items) {
    const fmtVal = config.formatValue(item.value);
    const pct = item.percent.toFixed(1) + "%";
    rows.push(
      `<tr><td class="key" style="padding-left:8px"><span>${escapeHtml(item.name)}</span></td><td class="value">${escapeHtml(fmtVal)}</td><td class="value" style="opacity:0.7">${pct}</td></tr>`,
    );
  }

  if (data.remainingCount > 0) {
    rows.push(
      `<tr><td colspan="3" class="key" style="padding-left:8px;padding-top:4px;opacity:0.6">and ${data.remainingCount} more</td></tr>`,
    );
  }

  return `<table><tbody>${rows.join("")}</tbody></table>`;
}
