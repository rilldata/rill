const SPLIT_DIMENSION_BY_NEW_LINE = /\s*\n\s*/g;
const SPLIT_DIMENSION_BY_COMMA = /\s*,\s*/g;

/**
 * Splits the dimension search text using newline is there is a newline in text, otherwise uses comma as a delimiter.
 */
export function splitDimensionSearchText(searchText: string) {
  // This is a hotfix to make sure commas in dimension value can be processed.
  // TODO: this is a hotfix, find a better way to split
  const hasNewLine = searchText.includes("\n");
  const values = hasNewLine
    ? // if the search text has newline then only split by newline
      searchText.split(SPLIT_DIMENSION_BY_NEW_LINE)
    : // else use comma as a delimiter
      searchText.split(SPLIT_DIMENSION_BY_COMMA);

  const hasEmptyLastValue =
    values.length > 0 && values[values.length - 1] === "";
  if (hasEmptyLastValue) {
    // Remove the last empty value when the last character is a comma/newline
    return values.slice(0, values.length - 1);
  }
  return values;
}
