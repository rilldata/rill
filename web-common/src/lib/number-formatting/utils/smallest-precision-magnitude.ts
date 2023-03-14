// returns the smallest order of magnitude to which a number
// has precision -- basically, the smallest OoM that has a non-zero
// digit. Returm NaN if number is NaN or infinite
export const smallestPrecisionMagnitude = (x: number): number => {
  if (typeof x !== "number") throw new Error("input must be a number");
  if (!isFinite(x)) return NaN;
  if (x === 0) return 0;
  // if the number is not an integer, find the smallest fractional digit
  if (!Number.isInteger(x)) {
    let e = 1;
    let p = 0;
    while (Math.round(x * e) / e !== x) {
      e *= 10;
      p++;
    }
    return -p;
  }

  let e = 10;
  let p = 0;
  while (Math.round(x / e) * e === x) {
    e *= 10;
    p++;
  }
  return p;
};
