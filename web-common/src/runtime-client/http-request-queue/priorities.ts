export const ActivePriority = 10;
export const ActiveColumnPriorityOffset = 50;
export const InactivePriority = 5;

// default is lower. override with higher
// NOTE: should always be less than 100 (reconcile priority)
export const DefaultQueryPriority = 10;
export const QueryPriorities = {
  rows: 50,
  "columns-profile": 45,
  "null-count": 40,
  "table-cardinality": 35,
  "column-cardinality": 35,
  "numeric-histogram": 30,
  timeseries: 30,
  topk: 10,
  "rug-histogram": 10,
  "descriptive-statistics": 10,
  totals: 30,
  "time-range-summary": 50,
};

export function getPriority(type: string): number {
  return QueryPriorities[type] ?? DefaultQueryPriority;
}

export function getPriorityForColumn(type: string, active: boolean): number {
  return getPriority(type) + (active ? ActiveColumnPriorityOffset : 0);
}
