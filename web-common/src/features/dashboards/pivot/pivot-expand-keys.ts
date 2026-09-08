/**
 * Value-based keys for pivot row expansion. A row is keyed by the dimension
 * values from the root to it, NUL-joined and hierarchical, so the key stays
 * stable across sorting, adding a field, and data refreshes. Mirrors the
 * encoding in pivot-click-selection.ts.
 */

// NUL separator, since dimension values won't contain it.
export const EXPAND_KEY_SEP = "\0";

// Sentinel for null values, so a null at depth N differs from an absent level.
const NULL_SENTINEL = "<NULL>";

// The leading separator can't occur in a real id, so this can't collide.
export const PIVOT_TOTALS_ROW_ID = EXPAND_KEY_SEP + "totals";

export function encodeExpandKeyValue(value: unknown): string {
  return value === null || value === undefined ? NULL_SENTINEL : String(value);
}

export function buildExpandKey(values: unknown[]): string {
  return values.map(encodeExpandKeyValue).join(EXPAND_KEY_SEP);
}

export function expandKeySegments(key: string): string[] {
  return key === "" ? [] : key.split(EXPAND_KEY_SEP);
}

export function expandKeyDepth(key: string): number {
  return expandKeySegments(key).length;
}

export function parentExpandKey(key: string): string {
  const i = key.lastIndexOf(EXPAND_KEY_SEP);
  return i === -1 ? "" : key.slice(0, i);
}

export function childExpandKey(parentKey: string, value: unknown): string {
  const segment = encodeExpandKeyValue(value);
  return parentKey === "" ? segment : parentKey + EXPAND_KEY_SEP + segment;
}
