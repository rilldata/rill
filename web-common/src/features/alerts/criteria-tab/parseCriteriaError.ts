const criteriaParserRegex = /criteria\[\d*]\.(.*)/;

// Parses errors in this format
// `criteria[0].value must be a 'number' type, ...`
export function parseCriteriaError(
  errors: Record<string, string[]> | undefined,
): string {
  if (!errors) return "";
  const fieldWithError = Object.keys(errors).filter((f) => !!errors[f])[0];
  if (!fieldWithError) return "";
  const firstError = errors[fieldWithError][0];
  if (!firstError) return "";

  let errorToReturn = "";
  const match = criteriaParserRegex.exec(firstError);
  if (match) {
    [, errorToReturn] = match;
  } else {
    // Error doesnt have the field already.
    if (firstError === "Required") {
      errorToReturn = `"${fieldWithError}" is required`;
    } else {
      errorToReturn = `${fieldWithError}: ${firstError}`;
    }
  }
  // `value1` in error is not user friendly. replacing with `criteria value`
  return errorToReturn.replace(/value1/, "criteria value");
}
