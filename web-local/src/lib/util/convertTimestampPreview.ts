import { removeTimezoneOffset as remove } from "@rilldata/web-common/lib/formatters";

export function convertTimestampPreviewFcn(ts, removeTimezoneOffset = false) {
  return removeTimezoneOffset ? remove(new Date(ts)) : new Date(ts);
}

/** used to convert a timestamp preview from the server for a sparkline. */
export function convertTimestampPreview(d, removeTimezoneOffset = false) {
  return d.map((di) => {
    pi = { ...di };
    pi.ts = convertTimestampPreviewFcn(di.ts, removeTimezoneOffset);
  });
}
