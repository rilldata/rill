/**
 * Helpers for the "@" measure picker in the ephemeral measure expression input.
 *
 * Typing "@" in the expression opens a picker of the referenceable measures,
 * searchable by display name as well as by name, so users do not need to know
 * a measure's field name to reference it. Picking one replaces the "@query"
 * with the measure's (quoted if needed) name; "@" itself is never part of a
 * valid expression.
 */
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import { formatMeasureRef } from "./expression-parser";

export type MeasureMention = {
  // Offset of the "@" in the expression.
  start: number;
  // Text between the "@" and the cursor.
  query: string;
};

// A "@" only starts a mention when it does not directly follow an identifier or
// a quoted name, so "@" pasted in the middle of a name is not picked up.
const IDENTIFIER_CHAR = /[A-Za-z0-9_"]/;

/**
 * Returns the mention the cursor is in, if any: the closest "@" before the
 * cursor with no other "@" in between. Spaces are allowed in the query since
 * display names contain them.
 */
export function findMeasureMention(
  value: string,
  cursor: number,
): MeasureMention | null {
  const start = value.lastIndexOf("@", cursor - 1);
  if (start < 0) return null;
  if (start > 0 && IDENTIFIER_CHAR.test(value[start - 1])) return null;
  const query = value.slice(start + 1, cursor);
  if (query.includes("\n")) return null;
  return { start, query };
}

/**
 * Returns the measures matching the mention query by display name or name,
 * keeping the metrics view's order. An empty query matches everything.
 */
export function filterMeasureMentions(
  measures: MetricsViewSpecMeasure[],
  query: string,
): MetricsViewSpecMeasure[] {
  const needle = query.trim().toLowerCase();
  if (!needle) return measures;
  return measures.filter(
    (mes) =>
      (mes.displayName ?? "").toLowerCase().includes(needle) ||
      (mes.name ?? "").toLowerCase().includes(needle),
  );
}

/**
 * Replaces the mention with a reference to the picked measure, followed by a
 * space unless one is already there, and returns the new value and cursor.
 */
export function applyMeasureMention(
  value: string,
  mention: MeasureMention,
  name: string,
): { value: string; cursor: number } {
  const end = mention.start + 1 + mention.query.length;
  const ref = formatMeasureRef(name);
  const separator = /^\s/.test(value.slice(end, end + 1)) ? "" : " ";
  const inserted = ref + separator;
  return {
    value: value.slice(0, mention.start) + inserted + value.slice(end),
    cursor: mention.start + inserted.length,
  };
}
