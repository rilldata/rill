const criteriaParserRegex = /criteria\[\d*]\.(.*)/;

// Parses errors in this format
// `criteria[0].value must be a 'number' type, ...`
export function parseCriteriaError(
  errors: Record<string, string[]> | undefined,
): string {
  if (!errors) return "";
  const errStr = Object.values(errors)[0];
  if (!errStr?.[0]) return "";
  const match = criteriaParserRegex.exec(errStr[0]);
  if (!match) return "";
  const [, matchedErr] = match;
  // `value1` in error is not user friendly. replacing with `criteria value`
  return matchedErr.replace(/^value1/, "criteria value");
}
