import type {
  V1ModelPartition,
  V1ModelSpec,
  V1ModelState,
} from "../../../runtime-client";

/**
 * Returns the sorted set of field names found in the partitions' data.
 */
export function getPartitionFields(partitions: V1ModelPartition[]): string[] {
  const fields = new Set<string>();
  for (const p of partitions) {
    for (const key of Object.keys(p.data ?? {})) fields.add(key);
  }
  return [...fields].sort();
}

function toNumber(v: unknown): number | undefined {
  if (typeof v === "number") return Number.isFinite(v) ? v : undefined;
  if (typeof v !== "string" || v.trim() === "") return undefined;
  const n = Number(v);
  return Number.isFinite(n) ? n : undefined;
}

/**
 * Returns true if `value` is within [start, end], both inclusive.
 * Values compare numerically when the value and both bounds are numeric,
 * so `"9"` sorts before `"12"`.
 * Otherwise they compare as strings, which orders ISO 8601 dates and timestamps correctly;
 * in that case an end bound that is a prefix of the value also matches,
 * so an end of `2026-03-08` includes `2026-03-08T00:00:00Z`.
 */
function isInRange(value: unknown, start: string, end: string): boolean {
  const n = toNumber(value);
  const lo = toNumber(start);
  const hi = toNumber(end);
  if (n !== undefined && lo !== undefined && hi !== undefined) {
    return n >= lo && n <= hi;
  }
  const s = String(value);
  return s >= start && (s <= end || s.startsWith(end));
}

export type PartitionRangeError = "missing" | "start_after_end";

/**
 * Returns the reason the range is invalid, or undefined if it is valid.
 * Uses the same ordering as filterPartitionsInRange.
 */
export function validatePartitionRange(
  field: string,
  start: string,
  end: string,
): PartitionRangeError | undefined {
  if (!field || !start || !end) return "missing";
  if (!isInRange(start, start, end)) return "start_after_end";
  return undefined;
}

/**
 * Returns the partitions whose `data[field]` falls within [start, end], both inclusive.
 * Partitions without the field, or with a non-scalar value, are excluded.
 */
export function filterPartitionsInRange(
  partitions: V1ModelPartition[],
  field: string,
  start: string,
  end: string,
): V1ModelPartition[] {
  return partitions.filter((p) => {
    const value = p.data?.[field];
    if (value === undefined || value === null || typeof value === "object") {
      return false;
    }
    return isInRange(value, start, end);
  });
}

/**
 * Returns true if the partition is waiting to be executed.
 * Mirrors the catalog's pending filter: not executed and not skipped.
 */
export function isPartitionPending(p: V1ModelPartition): boolean {
  return !p.executedOn && !p.skipped;
}

// Mirrors the defaults applied in executeWithRetry in runtime/reconcilers/model.go.
const DEFAULT_RETRY_ATTEMPTS = 3;
const DEFAULT_RETRY_DELAY_SECONDS = 5;
const DEFAULT_RETRY_EXPONENTIAL_BACKOFF = true;
const DEFAULT_RETRY_IF_ERROR_MATCHES = [
  ".*OvercommitTracker.*",
  ".*Bad Gateway.*",
  ".*Timeout.*",
  ".*Connection refused.*",
];

export type PartitionExecutionSettings = {
  /** Partitions executed at a time; undefined when unknown until the model runs with partitions. */
  concurrency: number | undefined;
  /** "observed" comes from the last partitioned run; "configured" comes from partitions_concurrency in the model file. */
  concurrencySource: "observed" | "configured" | "unknown";
  /** Total attempts per partition, including the first. */
  retryAttempts: number;
  retryDelaySeconds: number;
  retryExponentialBackoff: boolean;
  retryIfErrorMatches: string[];
  /** True when no retry option is set in the model file. */
  retryIsDefault: boolean;
};

/**
 * Resolves the read-only execution settings that apply to each partition of a backfill.
 */
export function getPartitionExecutionSettings(
  spec: V1ModelSpec | undefined,
  state: V1ModelState | undefined,
): PartitionExecutionSettings {
  // A configured limit is what the executor uses, so it wins over a value observed before the file changed.
  // When unset, the connector picks a default that is only known after a partitioned run.
  let concurrency: number | undefined;
  let concurrencySource: PartitionExecutionSettings["concurrencySource"] =
    "unknown";
  if (spec?.partitionsConcurrencyLimit) {
    concurrency = spec.partitionsConcurrencyLimit;
    concurrencySource = "configured";
  } else if (state?.partitionsConcurrency) {
    concurrency = state.partitionsConcurrency;
    concurrencySource = "observed";
  }

  const retryIfErrorMatches = spec?.retryIfErrorMatches?.length
    ? spec.retryIfErrorMatches
    : DEFAULT_RETRY_IF_ERROR_MATCHES;

  return {
    concurrency,
    concurrencySource,
    retryAttempts: Math.max(1, spec?.retryAttempts ?? DEFAULT_RETRY_ATTEMPTS),
    retryDelaySeconds: spec?.retryDelaySeconds ?? DEFAULT_RETRY_DELAY_SECONDS,
    retryExponentialBackoff:
      spec?.retryExponentialBackoff ?? DEFAULT_RETRY_EXPONENTIAL_BACKOFF,
    retryIfErrorMatches,
    retryIsDefault:
      spec?.retryAttempts === undefined &&
      spec?.retryDelaySeconds === undefined &&
      spec?.retryExponentialBackoff === undefined &&
      !spec?.retryIfErrorMatches?.length,
  };
}
