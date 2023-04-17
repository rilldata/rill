import { convertTimestampPreviewFcn } from "@rilldata/web-local/lib/util/convertTimestampPreview";
import { getDurationMultiple, getOffset } from "../../../lib/time/transforms";
import { TimeOffsetType } from "../../../lib/time/types";

/** sets extents to 0 if it makes sense; otherwise, inflates each extent component */
export function niceMeasureExtents(
  [smallest, largest]: [number, number],
  inflator: number
) {
  if (smallest === 0 && largest === 0) {
    return [0, 1];
  }
  return [
    smallest < 0 ? smallest * inflator : 0,
    largest > 0 ? largest * inflator : 0,
  ];
}

export function toComparisonKeys(d, offsetDuration: string) {
  return Object.keys(d).reduce((acc, key) => {
    if (key === "records") {
      Object.entries(d.records).forEach(([key, value]) => {
        acc[`comparison.${key}`] = value;
      });
    } else if (`comparison.${key}` === "comparison.ts") {
      acc[`comparison.${key}`] = convertTimestampPreviewFcn(d[key], true);
      acc["comparison.ts_position"] = getOffset(
        acc["comparison.ts"],
        offsetDuration,
        TimeOffsetType.ADD
      );
    } else {
      acc[`comparison.${key}`] = d[key];
    }
    return acc;
  }, {});
}

export function prepareTimeSeries(
  original,
  comparison,
  timeGrainDuration: string
) {
  return original.map((originalPt, i) => {
    const comparisonPt = comparison?.[i];

    const ts = convertTimestampPreviewFcn(originalPt.ts, true);
    const offsetDuration = getDurationMultiple(timeGrainDuration, 0.5);
    const ts_position = getOffset(ts, offsetDuration, TimeOffsetType.ADD);
    return {
      ts,
      ts_position,
      bin: originalPt.bin,
      ...originalPt.records,
      ...toComparisonKeys(comparisonPt || {}, offsetDuration),
    };
  });
}
