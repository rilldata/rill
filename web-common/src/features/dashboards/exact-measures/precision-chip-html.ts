import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { PrecisionChipProps } from "./exact-measures";

/**
 * Attribute that marks the precision chip in headers rendered as HTML strings
 * (the time dimension detail table). The table's mouse handler dispatches
 * clicks on elements carrying it to `PrecisionChipProps.onToggle`.
 */
export const PRECISION_CHIP_TOGGLE_ATTRIBUTE = "exact-toggle";

/**
 * HTML-string equivalent of `PrecisionChip.svelte`, for tables that render
 * their headers as markup strings instead of components. Styles mirror the
 * component; the tooltip is the native `title`.
 */
export function precisionChipHTML(precision: PrecisionChipProps): string {
  const { state, loading, error } = precision;
  let glyph = state === "approx" ? "≈" : "=";
  let tooltip: string;
  let colors = "border-border text-fg-muted bg-surface-background";
  if (loading) {
    glyph = "…";
    tooltip = m.dashboard_precision_loading();
    colors = "border-primary-300 text-primary-700 bg-surface-background";
  } else if (error) {
    glyph = "!";
    tooltip = m.dashboard_precision_error({ message: error });
    colors = "border-red-300 text-red-700 bg-red-50";
  } else if (state === "exact") {
    tooltip = m.dashboard_precision_exact();
    colors = "border-primary-400 text-primary-700 bg-primary-100";
  } else if (state === "partial") {
    tooltip = m.dashboard_precision_partial();
    colors = "border-primary-300 border-dashed text-primary-700 bg-primary-50";
  } else {
    tooltip = m.dashboard_precision_approx();
  }
  const cursor = precision.onToggle ? "cursor-pointer" : "cursor-default";
  return `<span ${PRECISION_CHIP_TOGGLE_ATTRIBUTE}="true" title="${escapeAttribute(tooltip)}" class="inline-flex items-center justify-center shrink-0 h-[15px] min-w-[15px] px-1 ml-1 rounded-full border text-[10px] font-bold leading-none not-italic ${cursor} ${colors}">${glyph}</span>`;
}

function escapeAttribute(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll('"', "&quot;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");
}
