export const ActivePriority = 10;
export const InactivePriority = 5;

// default is lower. override with higher
// NOTE: should always be less than 100 (reconcile priority)
export const DefaultQueryPriority = 10;
export const QueryPriorities = {
  rows: 15,
  topk: 15,
  "columns-profile": 5,
  "null-count": 10,
  cardinality: 10,

  "numeric-histogram": 20,
  "rug-histogram": 25,
  "descriptive-statistics": 30,

  "rollup-interval": 10,
  "smallest-time-grain": 10,
  "time-range-summary": 10,
};
