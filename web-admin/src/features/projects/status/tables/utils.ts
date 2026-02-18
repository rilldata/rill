import type {
  V1OlapTableInfo,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";

/**
 * Filters out temporary tables (e.g., __rill_tmp_ prefixed tables)
 */
export function filterTemporaryTables(
  tables: V1OlapTableInfo[] | undefined,
): V1OlapTableInfo[] {
  return (
    tables?.filter(
      (t): t is V1OlapTableInfo =>
        !!t.name && !t.name.startsWith("__rill_tmp_"),
    ) ?? []
  );
}

/**
 * Determines whether a table is likely a view based on its metadata.
 * Uses the view flag from OLAPGetTable, falling back to size heuristics.
 *
 * Note: returns true when metadata hasn't loaded yet (both params undefined),
 * so callers should account for the loading state separately if needed.
 */
export function isLikelyView(
  viewFlag: boolean | undefined,
  physicalSizeBytes: string | number | undefined,
): boolean {
  return (
    viewFlag === true ||
    physicalSizeBytes === "-1" ||
    physicalSizeBytes === 0 ||
    physicalSizeBytes === "0" ||
    !physicalSizeBytes
  );
}

/**
 * Parses a size value (string or number) to a number for sorting.
 * Returns -1 for invalid/missing values.
 */
export function parseSizeForSorting(size: string | number | undefined): number {
  if (size === undefined || size === null || size === "" || size === "-1") {
    return -1;
  }
  if (typeof size === "number") return size;
  const parsed = parseInt(size, 10);
  return isNaN(parsed) ? -1 : parsed;
}

/**
 * Compares two size values for ascending sort order.
 * TanStack Table handles direction via sortDescFirst.
 */
export function compareSizes(
  sizeA: string | number | undefined,
  sizeB: string | number | undefined,
): number {
  const numA = parseSizeForSorting(sizeA);
  const numB = parseSizeForSorting(sizeB);
  return numA - numB;
}

// ============================================
// Model Size Utils
// ============================================

/**
 * Formats a byte size for display. Returns "-" for invalid values.
 */
export function formatModelSize(bytes: string | number | undefined): string {
  if (bytes === undefined || bytes === null || bytes === "-1") return "-";

  let numBytes: number;
  if (typeof bytes === "number") {
    numBytes = bytes;
  } else {
    numBytes = parseInt(bytes, 10);
  }

  if (isNaN(numBytes) || numBytes < 0) return "-";
  return formatMemorySize(numBytes);
}

// ============================================
// Model Actions Utils
// ============================================

/**
 * Checks if a model resource is partitioned.
 */
export function isModelPartitioned(resource: V1Resource | undefined): boolean {
  return !!resource?.model?.spec?.partitionsResolver;
}

/**
 * Checks if a model resource is incremental.
 */
export function isModelIncremental(resource: V1Resource | undefined): boolean {
  return !!resource?.model?.spec?.incremental;
}

/**
 * Checks if a model resource has errored partitions.
 */
export function hasModelErroredPartitions(
  resource: V1Resource | undefined,
): boolean {
  return (
    !!resource?.model?.state?.partitionsModelId &&
    !!resource?.model?.state?.partitionsHaveErrors
  );
}

// ============================================
// Model Partitions Filter Utils
// ============================================

export type PartitionFilterType = "all" | "errors" | "pending";

/**
 * Returns whether to filter by errored partitions based on filter selection.
 */
export function shouldFilterByErrored(filter: PartitionFilterType): boolean {
  return filter === "errors";
}

/**
 * Returns whether to filter by pending partitions based on filter selection.
 */
export function shouldFilterByPending(filter: PartitionFilterType): boolean {
  return filter === "pending";
}
