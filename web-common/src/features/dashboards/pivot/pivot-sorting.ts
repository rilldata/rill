import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  PivotChipType,
  type PivotDataStoreConfig,
  type PivotState,
} from "./types";

// A measure under one column value in a nested table, as getAccessorForCell
// builds it: `c<column>v<value>` per column dimension joined by `_`, then
// `m<measure>`; just `m<measure>` when there are no column dimensions.
const MinimizedAccessorRegex = /^(?:c\d+v\d+(?:_c\d+v\d+)*)?m\d+$/;

/**
 * Whether a sort id is a minimized accessor, which points at its column
 * dimensions and measure by position rather than by name.
 */
export function isMinimizedAccessor(sortId: string): boolean {
  return MinimizedAccessorRegex.test(sortId);
}

/**
 * Whether the pivot still holds the field a sort refers to. A sort id is a
 * table column id: a dimension or measure name, a measure's comparison column
 * (`<measure>__delta_abs`), a time column (`<time dimension>_rill_<grain>`,
 * matched by its grain since the time dimension is not always at hand) or,
 * in a nested table, a measure under one column value (`c0v2m0`), which is
 * positional and so only checked for its shape here.
 * A sort outlives its field in URLs, bookmarks, reports and the editor
 * preview, so state loaders drop it when this is false.
 */
export function pivotHasSortField(
  sortId: string,
  pivot: Pick<PivotState, "rows" | "columns" | "tableMode">,
): boolean {
  if (pivot.tableMode !== "flat" && isMinimizedAccessor(sortId)) {
    return true;
  }
  const field = stripComparisonSuffix(sortId);
  return [...pivot.rows, ...pivot.columns].some((chip) =>
    chip.type === PivotChipType.Time
      ? sortId === chip.id || sortId.endsWith(`_rill_${chip.id}`)
      : chip.id === field,
  );
}

/**
 * Whether the pivot's queries can sort by `sortId` once the config is
 * resolved: in a flat table any of its dimension or measure columns; in a
 * nested table the first row dimension (the row axis), a measure (its row
 * total) or a measure under one column value whose column and measure
 * indexes exist. Comparison columns are in `measureNames` only while a
 * comparison is active, so a comparison sort stops applying when the
 * comparison is switched off.
 */
export function canSortPivotBy(
  sortId: string,
  config: Pick<
    PivotDataStoreConfig,
    "rowDimensionNames" | "colDimensionNames" | "measureNames" | "isFlat"
  >,
): boolean {
  const { rowDimensionNames, colDimensionNames, measureNames, isFlat } = config;
  if (measureNames.includes(sortId)) return true;
  // A flat table keeps all of its dimensions in rowDimensionNames.
  if (isFlat) return rowDimensionNames.includes(sortId);
  if (rowDimensionNames[0] === sortId) return true;
  if (!isMinimizedAccessor(sortId)) return false;
  const columnIndexes = [...sortId.matchAll(/c(\d+)v/g)].map((match) =>
    Number(match[1]),
  );
  const measureIndex = Number(sortId.slice(sortId.lastIndexOf("m") + 1));
  return (
    columnIndexes.every((index) => index < colDimensionNames.length) &&
    measureIndex < measureNames.length
  );
}

function stripComparisonSuffix(sortId: string) {
  for (const suffix of [COMPARISON_DELTA, COMPARISON_PERCENT]) {
    if (sortId.endsWith(suffix)) return sortId.slice(0, -suffix.length);
  }
  return sortId;
}
