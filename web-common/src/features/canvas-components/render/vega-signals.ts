/**
 * Util methods for handling vega signals
 */

export function resolveSignalField(value: unknown, field: string) {
  if (typeof value === "object" && value !== null) {
    return Array.isArray(value[field]) ? value[field][0] : undefined;
  }
  return undefined;
}

export function resolveSignalTimeField(value: unknown) {
  /**
   * Time fields end with `_ts`
   * We iterate over the keys of the object and return the first key that ends with `_ts`
   * */
  if (typeof value === "object" && value !== null) {
    for (const key in value) {
      if (key.endsWith("_ts")) {
        const ts = resolveSignalField(value, key);

        if (ts !== undefined) {
          return new Date(ts);
        }
      }
    }
  }
  return undefined;
}
