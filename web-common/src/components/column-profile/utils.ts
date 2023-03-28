import {
  BOOLEANS,
  CATEGORICALS,
  FLOATS,
  INTEGERS,
  isNested,
  TIMESTAMPS,
} from "@rilldata/web-common/lib/duckdb-data-types";
import type { NumericHistogramBinsBin } from "@rilldata/web-common/runtime-client";

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
  else if (isNested(a.type) && !isNested(b.type)) return 1;
  else if (!isNested(a.type) && isNested(b.type)) return -1;
  else if (INTEGERS.has(a.type) && !INTEGERS.has(b.type)) return 1;
  else if (!INTEGERS.has(a.type) && INTEGERS.has(b.type)) return -1;
  else if (TIMESTAMPS.has(a.type) && TIMESTAMPS.has(b.type)) {
    return -1;
  } else if (!TIMESTAMPS.has(a.type) && TIMESTAMPS.has(b.type)) {
    return 1;
  }
  return 0;
}

export function sortByName(a, b) {
  return a.name > b.name ? 1 : -1;
}

export function defaultSort(a, b) {
  const byType = sortByType(a, b);
  if (byType !== 0) return byType;
  /** sort nested types by cardinality, regardless of type. This should indicate
   * to the user if the nested type could easily be unnested into another simple column type
   * (e.g. low cardinality nested types may be better expressed as a VARCHAR or INTEGER)
   */
  if (isNested(a.type) && isNested(b.type)) return sortByCardinality(a, b);
  /** for all other non-categorical types, sort by nullity (e.g. timestamps, numerics) */
  if (!CATEGORICALS.has(a.type) && !CATEGORICALS.has(b.type))
    return sortByNullity(b, a);
  return sortByCardinality(a, b);
}

/** this is a temporary function for floating point numbers until we
 * move toward KDEs.
 */
export function chooseBetweenDiagnosticAndStatistical(
  diagnostic: NumericHistogramBinsBin[],
  statistical: NumericHistogramBinsBin[]
) {
  if (diagnostic?.length > 10) {
    return diagnostic;
  }
  if (diagnostic?.length > statistical?.length) {
    return diagnostic;
  }
  return statistical;
}
