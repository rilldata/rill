// returns the smallest order of magnitude to which a number
// has precision -- basically, the smallest OoM that has a non-zero
// digit. Returm NaN if number is NaN or infinite
export const smallestPrecisionMagnitude = (x: number): number => {
  if (typeof x !== "number") throw new Error("input must be a number");
  if (!isFinite(x)) return NaN;
  if (x === 0) return 0;

  // in the rare case of numbers that are too large to use with
  // the more efficient numeric approach belwo, fall back to a
  // string-based approach
  if (Math.abs(x) > 1e280) {
    const s = x.toExponential();
    const e_index = s.indexOf("e");
    const dot_index = s.indexOf(".");
    const exp = +s.slice(e_index + 1);
    const digits_after_dot = e_index - dot_index - 1;
    return exp - digits_after_dot;
  }

  // if the number is not an integer, find the smallest fractional digit
  if (!Number.isInteger(x)) {
    let e = 1;
    let p = 0;
    // can never represent a number with SMP < -324
    while (Math.round(x * e) / e !== x && p < 324) {
      e *= 10;
      p++;
    }
    return -p;
  }

  // if the number is an integer, find the smallest integer digit
  let e = 10;
  let p = 0;
  while (Math.round(x / e) * e === x) {
    e *= 10;
    p++;
  }
  return p;
};
