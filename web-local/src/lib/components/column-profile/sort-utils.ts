import {
  BOOLEANS,
  CATEGORICALS,
  FLOATS,
  INTEGERS,
  TIMESTAMPS,
} from "@rilldata/web-common/lib/duckdb-data-types";

export function sortByCardinality(a, b) {
  if (a.cardinality && b.cardinality) {
    if (a.cardinality < b.cardinality) {
      return 1;
    } else if (a.cardinality > b.cardinality) {
      return -1;
    } else {
      return sortByType(a, b);
    }
  } else {
    return sortByType(a, b);
  }
}

export function sortByNullity(a, b) {
  if (a.nullCount !== undefined && b.nullCount !== undefined) {
    if (a.nullCount < b.nullCount) {
      return 1;
    } else if (a.nullCount > b.nullCount) {
      return -1;
    } else {
      const byType = sortByType(a, b);
      if (byType) return byType;
      return sortByName(a, b);
    }
  }

  return sortByName(a, b);
}

export function sortByType(a, b) {
  if (BOOLEANS.has(a.type) && !BOOLEANS.has(b.type)) return 1;
  else if (!BOOLEANS.has(a.type) && BOOLEANS.has(b.type)) return -1;
  else if (CATEGORICALS.has(a.type) && !CATEGORICALS.has(b.type)) return 1;
  else if (!CATEGORICALS.has(a.type) && CATEGORICALS.has(b.type)) return -1;
  else if (FLOATS.has(a.type) && !FLOATS.has(b.type)) return 1;
  else if (!FLOATS.has(a.type) && FLOATS.has(b.type)) return -1;
  else if (INTEGERS.has(a.type) && !INTEGERS.has(b.type)) return 1;
  else if (!INTEGERS.has(a.type) && INTEGERS.has(b.type)) return -1;
  else if (TIMESTAMPS.has(a.type) && TIMESTAMPS.has(b.type)) {
    return -1;
  } else if (!TIMESTAMPS.has(a.type) && TIMESTAMPS.has(b.type)) {
    return 1;
  }
  return 0; //sortByName(a, b);
}

export function sortByName(a, b) {
  return a.name > b.name ? 1 : -1;
}

export function defaultSort(a, b) {
  const byType = sortByType(a, b);
  if (byType !== 0) return byType;
  if (!CATEGORICALS.has(a.type) && !CATEGORICALS.has(b.type))
    return sortByNullity(b, a);
  return sortByCardinality(a, b);
}
