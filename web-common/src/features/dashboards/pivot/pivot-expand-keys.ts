/**
 * Value-based keys for pivot row expansion. A row is keyed by the dimension
 * values from the root to it, NUL-joined and hierarchical, so the key stays
 * stable across sorting, adding a field, and data refreshes.
 */

// NUL separator, since dimension values won't contain it.
export const EXPAND_KEY_SEP = "\0";

// Every segment is tagged with its type, so no segment is ever empty
// (an empty-string value would otherwise look like the root key "")
// and a null can't collide with a string value such as "<NULL>" or "null".
const VALUE_PREFIX = "=";
const NULL_SEGMENT = "null";

export function encodeExpandKeyValue(value: unknown): string {
  return value === null || value === undefined
    ? NULL_SEGMENT
    : VALUE_PREFIX + String(value);
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
