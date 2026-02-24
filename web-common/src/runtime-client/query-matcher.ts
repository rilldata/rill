import type { Query } from "@tanstack/query-core";
import { ResourceKind } from "../features/entity-management/resource-selectors";

export function isRuntimeQuery(query: Query): boolean {
  const key = query.queryKey;
  // New format: [ServiceName, methodName, instanceId, request]
  const svc = key[0];
  if (
    svc === "QueryService" ||
    svc === "RuntimeService" ||
    svc === "ConnectorService"
  ) {
    return true;
  }
  // Old format: ["/v1/instances/..."]
  return typeof svc === "string" && svc.startsWith("/v1/instances/");
}

export function isGetResourceMetricsViewQuery(query: Query): boolean {
  const key = query.queryKey;
  // New format: ["RuntimeService", "getResource", instanceId, { "name.kind": ... }]
  if (key[0] === "RuntimeService" && key[1] === "getResource") {
    const request = key[3];
    return (
      typeof request === "object" &&
      request !== null &&
      (request as Record<string, unknown>)["name.kind"] ===
        ResourceKind.MetricsView
    );
  }
  // Old format
  const [apiPath, queryParams] = key;
  if (
    typeof apiPath !== "string" ||
    typeof queryParams !== "object" ||
    queryParams === null
  )
    return false;
  return (
    apiPath.startsWith("/v1/instances/") &&
    (queryParams as Record<string, unknown>)["name.kind"] ===
      ResourceKind.MetricsView
  );
}

export enum QueryRequestType {
  MetricsViewTopList = "toplist",
  MetricsViewCompareTopList = "compare-toplist",
  MetricsViewTimeSeries = "timeseries",
  MetricsViewTotals = "totals",
  MetricsViewRows = "rows",
  MetricsViewTimeRange = "time-range-summary",
  ColumnRollupInterval = "rollup-interval",
  ColumnTopK = "topk",
  ColumnNullCount = "null-count",
  ColumnDescriptiveStatistics = "descriptive-statistics",
  ColumnTimeGrain = "smallest-time-grain",
  ColumnNumericHistogram = "numeric-histogram",
  ColumnRugHistogram = "rug-histogram",
  ColumnTimeRange = "time-range-summary",
  ColumnCardinality = "column-cardinality",
  ColumnTimeSeries = "timeseries",
  TableCardinality = "table-cardinality",
  TableColumns = "columns-profile",
  TableRows = "rows",
}

// New format method names that correspond to profiling queries
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

/** Extract the table name from a new-format profiling query's request object */
function getTableNameFromRequest(request: unknown): string | undefined {
  if (typeof request !== "object" || request === null) return undefined;
  return (request as Record<string, unknown>).tableName as string | undefined;
}

const TableProfilingQuery: Partial<Record<QueryRequestType, boolean>> = {
  [QueryRequestType.TableCardinality]: true,
  [QueryRequestType.TableColumns]: true,
  [QueryRequestType.TableRows]: true,
};

const ProfilingQueryExtractor =
  /v1\/instances\/[a-zA-Z0-9-]+\/queries\/([a-zA-Z0-9-]+)\/tables\/(.+?)"/;

function isOlapQuery(queryHash: string, name: string) {
  return (
    queryHash.includes(`"/v1/connectors/olap/table"`) &&
    queryHash.includes(`"${name}"`)
  );
}

function isNewFormatOlapQuery(query: Query, name: string): boolean {
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
  // New format
  if (
    query.queryKey[0] === "QueryService" &&
    profilingMethods.has(query.queryKey[1] as string)
  ) {
    const tableName = getTableNameFromRequest(query.queryKey[3]);
    return tableName === name;
  }
  if (isNewFormatOlapQuery(query, name)) return true;
  // Old format
  const queryExtractorMatch = ProfilingQueryExtractor.exec(query.queryHash);
  if (!queryExtractorMatch) return isOlapQuery(query.queryHash, name);

  const [, , table] = queryExtractorMatch;
  return table === name;
}

export function isTableProfilingQuery(query: Query, name: string): boolean {
  // New format
  if (
    query.queryKey[0] === "QueryService" &&
    tableProfilingMethods.has(query.queryKey[1] as string)
  ) {
    const tableName = getTableNameFromRequest(query.queryKey[3]);
    return tableName === name;
  }
  if (isNewFormatOlapQuery(query, name)) return true;
  // Old format
  const queryExtractorMatch = ProfilingQueryExtractor.exec(query.queryHash);
  if (!queryExtractorMatch) return isOlapQuery(query.queryHash, name);

  const [, type, table] = queryExtractorMatch;
  return table === name && type in TableProfilingQuery;
}

export function isColumnProfilingQuery(query: Query, name: string) {
  // New format
  if (
    query.queryKey[0] === "QueryService" &&
    profilingMethods.has(query.queryKey[1] as string) &&
    !tableProfilingMethods.has(query.queryKey[1] as string)
  ) {
    const tableName = getTableNameFromRequest(query.queryKey[3]);
    return tableName === name;
  }
  // Old format
  const queryExtractorMatch = ProfilingQueryExtractor.exec(query.queryHash);
  if (!queryExtractorMatch) return false;

  const [, type, table] = queryExtractorMatch;
  return table === name && !(type in TableProfilingQuery);
}
