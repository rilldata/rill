/**
 * Deep-strip undefined values from an object tree.
 *
 * Proto's `fromJson()` rejects `undefined` values; Orval's HTTP client
 * silently omitted them. This bridges the gap by recursively removing
 * undefined entries before passing objects to proto serialization.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function stripUndefined(
  obj: Record<string, any>,
): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj)) {
    if (value === undefined) continue;
    if (Array.isArray(value)) {
      result[key] = value.map((item) =>
        item && typeof item === "object" && !Array.isArray(item)
          ? stripUndefined(item)
          : item,
      );
    } else if (value && typeof value === "object" && !(value instanceof Date)) {
      result[key] = stripUndefined(value);
    } else {
      result[key] = value;
    }
  }
  return result;
}
