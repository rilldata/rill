import { sanitizeQuery } from "@rilldata/web-local/lib/util/sanitize-query";

/**
 * @returns true iff @param query is likely to be a PRQL query
 */
function isPRQL(query: string) {
  return query.trim().toLowerCase().startsWith("from");
}

/**
 * Compile PRQL and correctly throw errors.
 */
function compilePRQL(prqlQuery: string) {
  const prql = require("prql-js/dist/node");

  const result = prql.compile(prqlQuery);

  if (result.error) {
    throw new Error(result.error.message);
  } else {
    return result.sql;
  }
}

export function preprocessQuery(query: string) {
  const sqlQuery = isPRQL(query) ? compilePRQL(query) : query;

  return sanitizeQuery(sqlQuery, false);
}
