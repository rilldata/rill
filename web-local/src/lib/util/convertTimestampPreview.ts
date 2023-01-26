import { removeTimezoneOffset as remove } from "@rilldata/web-common/lib/formatters";

/** used to convert a timestamp preview from the server for a sparkline. */
export function convertTimestampPreview(d, removeTimezoneOffset = false) {
  return d.map((di) => {
    const pi = { ...di };
    pi.ts = removeTimezoneOffset ? remove(new Date(pi.ts)) : new Date(pi.ts);
    return pi;
  });
}
