/**
 * Priority mapping for ConnectRPC method names.
 * Mirrors the URL-based priorities from http-request-queue/priorities.ts,
 * mapped to ConnectRPC service method names instead of URL path segments.
 */

// Maps ConnectRPC method names to priority weights.
// Higher priority = dispatched first.
const MethodPriorities: Record<string, number> = {
  // High priority: user-visible data
  MetricsViewRows: 50,
  MetricsViewTimeRange: 50,
  ColumnProfile: 45,

  // Medium: charts and summaries
  ColumnNullCount: 40,
  TableCardinality: 35,
  ColumnCardinality: 35,
  MetricsViewAggregation: 30,
  MetricsViewTimeSeries: 30,
  NumericHistogram: 30,
  MetricsViewTotals: 30,

  // Low: exploratory queries
  MetricsViewToplist: 10,
  RugHistogram: 10,
  DescriptiveStatistics: 10,
};

export const DEFAULT_PRIORITY = 10;
export const ACTIVE_COLUMN_PRIORITY_OFFSET = 50;
export const ACTIVE_PRIORITY = 10;
export const INACTIVE_PRIORITY = 5;

export function getPriorityForMethod(methodName: string): number {
  return MethodPriorities[methodName] ?? DEFAULT_PRIORITY;
}

// URL-segment-style priority keys used by column profile queries.
const ColumnQueryPriorities: Record<string, number> = {
  topk: 10,
  timeseries: 30,
  "numeric-histogram": 30,
  "rug-histogram": 10,
  "descriptive-statistics": 10,
};

export function getPriorityForColumn(type: string, active: boolean): number {
  const base = ColumnQueryPriorities[type] ?? DEFAULT_PRIORITY;
  return base + (active ? ACTIVE_COLUMN_PRIORITY_OFFSET : 0);
}
