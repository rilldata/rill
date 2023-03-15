// returns the smallest order of magnitude to which a number
// has precision -- basically, the smallest OoM that has a non-zero
// digit. Returm NaN if number is NaN or infinite
export const smallestPrecisionMagnitude = (x: number): number => {
  if (!isFinite(x)) return NaN;
  if (x === 0) return 0;
  // if the number is not an integer, find the smallest fractional digit
  if (!Number.isInteger(x)) {
    // x = x - Math.round(x);
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

const testCases: [number, number][] = [
  [Infinity, NaN],
  [-Infinity, NaN],
  [NaN, NaN],
  [0, 0],
  [-0, 0],
  [0.1, -1],
  [0.000000001, -9],
  [0.000027391, -9],
  [0.973427391, -9],
  [123, 0],
  [1230, 1],
  [123000000000, 9],
  [1239347293742974, 0],
  [6000000000, 9],
  [-0.1, -1],
  [-0.000000001, -9],
  [-0.000027391, -9],
  [-0.973427391, -9],
  [-123, 0],
  [-1230, 1],
  [-123000000000, 9],
  [-1239347293742974, 0],
  [-6000000000, 9],
  [710.7237956, -7],
  [-710.7237956, -7],

  [79879879710.7237, -4], // NOTE: most digits representable in js
  [-79879879710.7237, -4], // NOTE: most digits representable in js
];

export const runTestsmallestPrecisionMagnitude = () => {
  const caseResults = testCases.map((tc) => {
    const out = smallestPrecisionMagnitude(tc[0]);
    const correct = Number.isFinite(tc[0]) ? out === tc[1] : Number.isNaN(out);
    return [correct ? "pass" : "fail", tc, out];
  });

  const overallPass = caseResults.reduce(
    (prev, tc) => prev && tc[0] === "pass",
    true
  );

  if (!overallPass) {
    console.error("smallestPrecisionMagnitude test FAIL:", caseResults);
  } else {
    console.log("smallestPrecisionMagnitude test pass:", caseResults);
  }
};
