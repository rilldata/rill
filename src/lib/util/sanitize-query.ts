export function sanitizeQuery(query: string, toLower = true) {
  if (query.startsWith('from')) {
    return query;
  }
  else {
    const noComments = query.replace(/--.*/g, " ");
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
}
