/**
 * Value-based keys for pivot row expansion. A row is keyed by the
 * (dimension, value) pairs from the root to it, NUL-joined and hierarchical,
 * so the key stays stable across sorting, adding a field, and data refreshes.
 *
 * Each segment names its dimension, so a key only matches a row when the same
 * dimensions sit at the same depths. After the row dimensions are removed,
 * reordered or swapped, keys from the old layout match nothing instead of
 * landing on a different dimension's row that happens to share a value
 * (e.g. "Not Available" or null).
 */

// NUL separator; JSON escapes control characters, so it never occurs inside
// a segment.
export const EXPAND_KEY_SEP = "\0";

// JSON keeps the value's type (null, "null" and 5 all differ) and never
// yields an empty segment, so an empty-string value can't be mistaken for the
// root key "".
export function encodeExpandKeySegment(
  dimension: string,
  value: unknown,
): string {
  return JSON.stringify([dimension, value ?? null]);
}

export function buildExpandKey(
  dimensions: string[],
  values: unknown[],
): string {
  return values
    .map((value, i) => encodeExpandKeySegment(dimensions[i], value))
    .join(EXPAND_KEY_SEP);
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

export function childExpandKey(
  parentKey: string,
  dimension: string,
  value: unknown,
): string {
  const segment = encodeExpandKeySegment(dimension, value);
  return parentKey === "" ? segment : parentKey + EXPAND_KEY_SEP + segment;
}
