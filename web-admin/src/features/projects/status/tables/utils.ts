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
 * Parses a size value (string or number) to a number for sorting.
 * Returns -1 for invalid/missing values.
 */
export function parseSizeForSorting(size: string | number | undefined): number {
  if (!size || size === "-1") {
    return -1;
  }
  return typeof size === "number" ? size : parseInt(size, 10);
}

/**
 * Compares two size values for descending sort order.
 * Used for sorting tables by database size.
 */
export function compareSizesDescending(
  sizeA: string | number | undefined,
  sizeB: string | number | undefined,
): number {
  const numA = parseSizeForSorting(sizeA);
  const numB = parseSizeForSorting(sizeB);
  return numB - numA;
}

/**
 * Formats a timestamp string to locale time (HH:MM:SS).
 */
export function formatLogTime(time: string | undefined): string {
  if (!time) return "";
  const date = new Date(time);
  return date.toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

/**
 * Returns the CSS class for a log level.
 */
export function getLogLevelClass(level: string | undefined): string {
  switch (level) {
    case "LOG_LEVEL_ERROR":
    case "LOG_LEVEL_FATAL":
      return "text-red-600";
    case "LOG_LEVEL_WARN":
      return "text-yellow-600";
    default:
      return "text-fg-muted";
  }
}

/**
 * Returns the display label for a log level.
 */
export function getLogLevelLabel(level: string | undefined): string {
  switch (level) {
    case "LOG_LEVEL_ERROR":
      return "ERROR";
    case "LOG_LEVEL_FATAL":
      return "FATAL";
    case "LOG_LEVEL_WARN":
      return "WARN";
    case "LOG_LEVEL_INFO":
      return "INFO";
    case "LOG_LEVEL_DEBUG":
      return "DEBUG";
    default:
      return "INFO";
  }
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
