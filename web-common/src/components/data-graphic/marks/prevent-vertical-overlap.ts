/** This function implements a point-pushing algorithm that prevents
 * y values from overlapping. It is used to prevent labels from
 * overlapping in a data graphic.
 * The algorithm is as follows:
 * 1. Sort the points by y value.
 * 2. If there is only one point, return it.
 * 3. If there are no points, return an empty array.
 * 4. Calculate the middle index.
 * 5. Adjust the position of the points above the middle index.
 * 6. Adjust the position of the top point and push down overlapping points.
 * 7. Adjust the position of the points below the middle index.
 * 8. Adjust the position of the bottom point and push up overlapping points.
 * We can't claim that this works for every case, but it seems to work
 * in most cases we care about.
 */
export function preventVerticalOverlap(
  pt: { key: unknown; value: number }[],
  topBoundary,
  bottomBoundary,
  elementHeight,
  yBuffer
): { key: unknown; value: number }[] {
  // this is where the boundary condition lives.

  const locations = [...pt.map((p) => ({ ...p }))];
  // sort the locations by y value.
  locations.sort((a, b) => a.value - b.value);

  if (locations.length === 1) {
    return locations;
  }

  if (!locations.length) return locations;

  // calculate the middle index.
  const middle = ~~(locations.length / 2); // eslint-disable-line

  // Adjust position of labels above the middle index
  let i = middle;
  while (i >= 0) {
    if (i !== middle) {
      const diff = locations[i + 1].value - locations[i].value;
      if (diff <= elementHeight + yBuffer) {
        locations[i].value -= elementHeight + yBuffer - diff;
      }
    }
    i -= 1;
  }

  // Adjust position of top label and push down overlapping labels
  if (locations[0].value < topBoundary + yBuffer) {
    locations[0].value = topBoundary + yBuffer;
    i = 0;
    while (i < middle) {
      const diff = locations[i + 1].value - locations[i].value;
      if (diff <= elementHeight + yBuffer) {
        locations[i + 1].value += elementHeight + yBuffer - diff;
      }
      i += 1;
    }
  }

  // Adjust position of labels below the middle index
  i = middle;
  while (i < locations.length) {
    if (i !== middle) {
      const diff = locations[i].value - locations[i - 1].value;
      if (diff < elementHeight + yBuffer) {
        locations[i].value += elementHeight + yBuffer - diff;
      }
    }
    i += 1;
  }

  // Adjust position of bottom label and push up overlapping labels
  if (locations[locations.length - 1].value > bottomBoundary - yBuffer) {
    locations[locations.length - 1].value = bottomBoundary - yBuffer;
    i = locations.length - 1;
    while (i > 0) {
      const diff = locations[i].value - locations[i - 1].value;
      if (diff <= elementHeight + yBuffer) {
        locations[i - 1].value -= elementHeight + yBuffer - diff;
      }
      i -= 1;
    }
  }
  return locations;
}
