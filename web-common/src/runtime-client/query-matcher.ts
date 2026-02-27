import type { Query } from "@tanstack/query-core";

// Method names that correspond to profiling queries
const profilingMethods = new Set([
  "columnRollupInterval",
  "columnTopK",
  "columnNullCount",
  "columnDescriptiveStatistics",
  "columnTimeGrain",
  "columnNumericHistogram",
  "columnRugHistogram",
  "columnTimeRange",
  "columnCardinality",
  "columnTimeSeries",
  "tableCardinality",
  "tableColumns",
  "tableRows",
]);

const tableProfilingMethods = new Set([
  "tableCardinality",
  "tableColumns",
  "tableRows",
]);

/** Extract the table name from a profiling query's request object */
function getTableNameFromRequest(request: unknown): string | undefined {
  if (typeof request !== "object" || request === null) return undefined;
  return (request as Record<string, unknown>).tableName as string | undefined;
}

function isOlapQuery(query: Query, name: string): boolean {
  const key = query.queryKey;
  if (
    key[0] === "ConnectorService" &&
    (key[1] === "oLAPGetTable" || key[1] === "oLAPListTables")
  ) {
    const request = key[3];
    return (
      typeof request === "object" &&
      request !== null &&
      (request as Record<string, unknown>).table === name
    );
  }
  return false;
}

export function isProfilingQuery(query: Query, name: string): boolean {
  if (
    query.queryKey[0] === "QueryService" &&
    profilingMethods.has(query.queryKey[1] as string)
  ) {
    const tableName = getTableNameFromRequest(query.queryKey[3]);
    return tableName === name;
  }
  return isOlapQuery(query, name);
}

export function isTableProfilingQuery(query: Query, name: string): boolean {
  if (
    query.queryKey[0] === "QueryService" &&
    tableProfilingMethods.has(query.queryKey[1] as string)
  ) {
    const tableName = getTableNameFromRequest(query.queryKey[3]);
    return tableName === name;
  }
  return isOlapQuery(query, name);
}

export function isColumnProfilingQuery(query: Query, name: string) {
  if (
    query.queryKey[0] === "QueryService" &&
    profilingMethods.has(query.queryKey[1] as string) &&
    !tableProfilingMethods.has(query.queryKey[1] as string)
  ) {
    const tableName = getTableNameFromRequest(query.queryKey[3]);
    return tableName === name;
  }
  return false;
}
