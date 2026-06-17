/**
 * Rendering sparse data with null gaps
 *
 * 1. Null bridging: `bridgeGaps` fills all gaps with zeros when
 *    `connectNulls` is on, treating missing data as zero, so lines route
 *    through zero rather than breaking. With `connectNulls` off, gaps
 *    remain as nulls and produce natural line breaks.
 *
 * 2. Clip paths: The primary series needs clip paths because its area
 *    fill gradient would otherwise render across gaps (`defined` only
 *    affects line generators, not the filled path).
 *      - `seg-clip`:   real data segments only (connectNulls off)
 *      - `full-clip`:  real + bridged segments (connectNulls on, area fill)
 *      - `scrub-clip`: scrub selection rect — chart draws muted, then
 *                      re-draws with original colors inside this clip
 *    Secondary series have no area fill, so they rely on the line
 *    generator's `defined` callback and only use `scrub-clip`.
 *
 */

/**
 * Utilities for handling sparse (null-gapped) time-series data.
 * Used by both TimeSeriesChart (explore dashboards) and Chart (KPI sparkline).
 */

export interface Segment {
  startIndex: number;
  endIndex: number;
}

export interface BridgeResult<T> {
  values: T[];
  /** Segments from the original (un-bridged) data */
  inputSegments: Segment[];
  /** Segments from the bridged data (inputSegments + small gaps filled) */
  bridgedSegments: Segment[];
}

/**
 * Find contiguous non-null segments in a values array.
 */
export function computeSegments<T>(
  data: T[],
  valueAccessor: (d: T) => number | null | undefined,
): Segment[] {
  const segments: Segment[] = [];
  let segStart = -1;
  for (let i = 0; i < data.length; i++) {
    const v = valueAccessor(data[i]);
    if (v !== null && v !== undefined) {
      if (segStart === -1) segStart = i;
    } else if (segStart !== -1) {
      segments.push({ startIndex: segStart, endIndex: i - 1 });
      segStart = -1;
    }
  }
  if (segStart !== -1)
    segments.push({ startIndex: segStart, endIndex: data.length - 1 });
  return segments;
}

/**
 * Fill all gaps with zeros, treating missing data as zero so adjacent
 * segments connect through zero.
 * Returns a new array with gap values filled, plus segment metadata.
 *
 * When `connectNulls` is false, no bridging is performed; the result
 * still includes segment metadata for clip paths and singleton detection.
 *
 * @param data           The data array
 * @param valueAccessor  Extracts the numeric value (or null) from a data point
 * @param cloneWithValue Creates a new data point with the interpolated value
 * @param connectNulls   Whether to bridge gaps (default true)
 */
export function bridgeGaps<T>(
  data: T[],
  valueAccessor: (d: T) => number | null | undefined,
  cloneWithValue: (d: T, value: number) => T,
  connectNulls: boolean = true,
): BridgeResult<T> {
  const inputSegments = computeSegments(data, valueAccessor);

  if (!connectNulls || data.length < 3 || inputSegments.length <= 1) {
    return {
      values: data,
      inputSegments,
      bridgedSegments: inputSegments,
    };
  }

  const result = [...data];

  for (let i = 0; i < inputSegments.length - 1; i++) {
    const prev = inputSegments[i];
    const next = inputSegments[i + 1];
    for (let j = prev.endIndex + 1; j < next.startIndex; j++) {
      result[j] = cloneWithValue(data[j], 0);
    }
  }

  const bridgedSegments = computeSegments(result, valueAccessor);
  return { values: result, inputSegments, bridgedSegments };
}
