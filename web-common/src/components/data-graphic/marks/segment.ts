function pathIsDefined(yAccessor: string) {
  return (d: Record<string, number>) => {
    const val = d[yAccessor];
    return !(
      val === undefined ||
      (typeof val === "number" && isNaN(val)) ||
      val === null
    );
  };
}

/**
 * Helper function to compute the contiguous segments of the data
 * based on https://github.com/pbeshai/d3-line-chunked/blob/master/src/lineChunked.js
 */
export function computeSegments(
  lineData: Record<string, number>[],
  yAccessor: string,
): Record<string, number>[][] {
  let startNewSegment = true;

  const defined = pathIsDefined(yAccessor);
  // split into segments of continuous data
  const segments = lineData.reduce(
    (segments: Record<string, number>[][], d) => {
      // skip if this point has no data
      if (!defined(d)) {
        startNewSegment = true;
        return segments;
      }

      // if we are starting a new segment, start it with this point
      if (startNewSegment) {
        segments.push([d]);
        startNewSegment = false;

        // otherwise see if we are adding to the last segment
      } else {
        const lastSegment = segments[segments.length - 1];
        lastSegment.push(d);
        // if we expect this point to come next, add it to the segment
      }

      return segments;
    },
    [],
  );

  return segments;
}

/**
 * Compute the gaps from segments. Takes an array of segments and creates new segments
 * based on the edges of adjacent segments.
 *
 * @param {Array} segments The segments array (e.g. from computeSegments)
 * @return {Array} gaps The gaps array (same form as segments, but representing spaces between segments)
 */
export function gapsFromSegments(segments) {
  const gaps = [];
  for (let i = 0; i < segments.length - 1; i++) {
    const currSegment = segments[i];
    const nextSegment = segments[i + 1];

    gaps.push([currSegment[currSegment.length - 1], nextSegment[0]]);
  }

  return gaps;
}
