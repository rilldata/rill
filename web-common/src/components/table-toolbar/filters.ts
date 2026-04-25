import type { SortDirection } from "./types";

/**
 * Apply search, filter, and sort to a table dataset. Each filter is expressed
 * as a predicate so call sites can inline whatever per-dimension matching
 * logic they need.
 *
 * Sort: rows where `getSortKey` returns undefined or an empty string sort to
 * the bottom of the "newest" view (and to the top of "oldest").
 */
export function applyTableFilters<T>(opts: {
  data: T[];
  searchText?: string;
  matchesSearch?: (row: T, query: string) => boolean;
  filterPredicates?: ((row: T) => boolean)[];
  sortDirection: SortDirection;
  getSortKey: (row: T) => string | undefined;
}): T[] {
  const {
    data,
    searchText = "",
    matchesSearch,
    filterPredicates = [],
    sortDirection,
    getSortKey,
  } = opts;
  return data
    .filter((row) => {
      if (searchText && matchesSearch && !matchesSearch(row, searchText)) {
        return false;
      }
      for (const predicate of filterPredicates) {
        if (!predicate(row)) return false;
      }
      return true;
    })
    .slice()
    .sort((a, b) => {
      const aKey = getSortKey(a) ?? "";
      const bKey = getSortKey(b) ?? "";
      const cmp = aKey < bKey ? -1 : aKey > bKey ? 1 : 0;
      return sortDirection === "newest" ? -cmp : cmp;
    });
}

/** Toggle a value in an array: add if absent, remove if present. */
export function toggleArrayValue<T>(arr: T[], value: T): T[] {
  return arr.includes(value) ? arr.filter((v) => v !== value) : [...arr, value];
}
