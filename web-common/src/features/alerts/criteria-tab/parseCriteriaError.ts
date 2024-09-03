const criteriaParserRegex = /criteria\[(\d)*]\.(.*)/;

// Parses errors in this format
// `criteria[0].value must be a 'number' type, ...`
export function parseCriteriaError(errStr: string, index: number): string {
  const match = criteriaParserRegex.exec(errStr);
  if (!match) return "";
  const [, matchedIndex, matchedErr] = match;
  return Number(matchedIndex) === index
    ? // `value1` in error is not user friendly. replacing with `criteria value`
      matchedErr.replace(/^value1/, "criteria value")
    : "";
}
