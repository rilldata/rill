export function sanitizeQuery(query: string, toLower = true) {
  // remove comments;
  const noComments = query.replace(/--.*/g, " ");
  // remove double+ spaces, \ns.
  let output = noComments
    .replace(/\n/g, " ")
    .replace(/\s\s+/g, " ")
    .replace(/,\s+/g, ",")
    .replace(/;/g, "")
    .trim();
  if (toLower) {
    output = output.toLowerCase();
  }
  return output;
}
