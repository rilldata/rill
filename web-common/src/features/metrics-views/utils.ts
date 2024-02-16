export function getModelOutOfPossiblyMalformedYAML(
  yaml: string,
): string | null {
  // Regular expression to match model key followed by its value
  // The regex looks for 'model:' followed by any number of whitespaces and captures any non-whitespace characters after that
  const regex = /model:\s*(\S+)/;

  // Extract the match groups
  const matches = regex.exec(yaml);

  // If matches were found, return the value of the model field, otherwise return null
  return matches && matches[1] ? matches[1] : null;
}

export function getTableOutOfPossiblyMalformedYAML(
  yaml: string,
): string | null {
  // Regular expression to match table key followed by its value
  // The regex looks for 'table:' followed by any number of whitespaces and captures any non-whitespace characters after that
  const regex = /table:\s*(\S+)/;

  // Extract the match groups
  const matches = regex.exec(yaml);

  // If matches were found, return the value of the table field, otherwise return null
  return matches && matches[1] ? matches[1] : null;
}
