import type { Query } from "@tanstack/query-core";
import { ResourceKind } from "../features/entity-management/resource-selectors";

export function isRuntimeQuery(query: Query): boolean {
  const [apiPath] = query.queryKey as string[];
  return apiPath.startsWith("/v1/instances/");
}

export function isGetResourceMetricsViewQuery(query: Query): boolean {
  const [apiPath, queryParams] = query.queryKey; // Renamed for clarity
  if (
    typeof apiPath !== "string" ||
    typeof queryParams !== "object" ||
    queryParams === null
  )
    return false;
  return (
    apiPath.startsWith("/v1/instances/") &&
    queryParams["name.kind"] === ResourceKind.MetricsView
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

export function isProfilingQuery(query: Query, name: string): boolean {
  const queryExtractorMatch = ProfilingQueryExtractor.exec(query.queryHash);
  if (!queryExtractorMatch) return isOlapQuery(query.queryHash, name);

  const [, , table] = queryExtractorMatch;
  return table === name;
}

export function isTableProfilingQuery(query: Query, name: string): boolean {
  const queryExtractorMatch = ProfilingQueryExtractor.exec(query.queryHash);
  if (!queryExtractorMatch) return isOlapQuery(query.queryHash, name);

  const [, type, table] = queryExtractorMatch;
  return table === name && type in TableProfilingQuery;
}

export function isColumnProfilingQuery(query: Query, name: string) {
  const queryExtractorMatch = ProfilingQueryExtractor.exec(query.queryHash);
  if (!queryExtractorMatch) return false;

  const [, type, table] = queryExtractorMatch;
  return table === name && !(type in TableProfilingQuery);
}
