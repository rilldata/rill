import {
  addZoneOffset,
  removeLocalTimezoneOffset as remove,
} from "@rilldata/web-common/lib/time/timezone";
import type { V1TimeSeriesValue } from "../runtime-client";
export function convertTimestampPreviewFcn(
  ts,
  removeLocalTimezoneOffset = false,
) {
  return removeLocalTimezoneOffset ? remove(new Date(ts)) : new Date(ts);
}

/** used to convert a timestamp preview from the server for a sparkline. */
export function convertTimestampPreview(
  d: V1TimeSeriesValue[],
  removeLocalTimezoneOffset = false,
) {
  return d.map((di) => {
    return {
      ...di,
      ts: convertTimestampPreviewFcn(di.ts, removeLocalTimezoneOffset),
    };
  });
}

/** used to remove local timezone offset and add dashboard selected zone offset */
export function adjustOffsetForZone(
  ts: Date | string | undefined,
  zone: string,
) {
  if (!ts) return ts;
  return addZoneOffset(remove(new Date(ts)), zone);
}
