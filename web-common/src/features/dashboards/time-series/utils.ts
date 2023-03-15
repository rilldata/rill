export function niceMeasureExtents(
  [smallest, largest]: [number, number],
  inflator: number
) {
  if (smallest === 0 && largest === 0) {
    return [0, 1];
  }
  return [
    // If the smallest value is negative, we want to inflate it by the inflation factor.
    smallest < 0 ? smallest * inflator : 0,
    // If the largest value is positive, we want to inflate it by the inflation factor.
    largest > 0 ? largest * inflator : 0,
  ];
}
