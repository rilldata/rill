import type { V1Resource } from "@rilldata/web-common/runtime-client";

export interface DescribeEntry {
  section: string;
  label: string;
  value: string;
  mono?: boolean;
  /** If true, render as a code block instead of a single truncated line */
  multiline?: boolean;
  /** Groups entries into visual cards within a section (e.g. each measure, dimension) */
  group?: string;
}

// Fields to skip entirely; these are internal/noisy
const SKIP_KEYS = new Set([
  "trigger",
  "triggerFull",
  "triggerPartitions",
  "specHash",
  "refsHash",
  "testHash",
  "specVersion",
  "stateVersion",
  "version",
  "resultProperties",
  "incrementalState",
  "incrementalStateResolverProperties",
  "partitionsResolverProperties",
  "stageProperties",
  "outputProperties",
  // Already shown in Metadata section
  "refreshedOn",
  "reconcileError",
  "reconcileStatus",
  "reconcileOn",
  "specUpdatedOn",
  "stateUpdatedOn",
]);

// Fields that contain code and should render as multiline blocks
const MULTILINE_KEYS = new Set(["sql", "expression", "watermarkExpression"]);

// Human-readable labels for common proto field names
const LABEL_MAP: Record<string, string> = {
  sourceConnector: "Source Connector",
  sinkConnector: "Sink Connector",
  refreshSchedule: "Refresh Schedule",
  timeoutSeconds: "Timeout (seconds)",
  stageChanges: "Stage Changes",
  streamIngestion: "Stream Ingestion",
  inputConnector: "Input Connector",
  outputConnector: "Output Connector",
  inputProperties: "Input Properties",
  smallestTimeGrain: "Smallest Time Grain",
  timeDimension: "Time Dimension",
  watermarkExpression: "Watermark Expression",
  firstDayOfWeek: "First Day of Week",
  firstMonthOfYear: "First Month of Year",
  databaseSchema: "Database Schema",
  metricsView: "Metrics View",
  definedInMetricsView: "Defined in Metrics View",
  definedInCanvas: "Defined in Canvas",
  definedAsSource: "Defined as Source",
  securityRules: "Security Rules",
  rendererProperties: "Renderer Properties",
  filtersEnabled: "Filters Enabled",
  refreshedOn: "Last Refreshed",
  reconcileError: "Reconcile Error",
  reconcileStatus: "Reconcile Status",
  reconcileOn: "Next Reconcile",
  specUpdatedOn: "Spec Updated",
  stateUpdatedOn: "State Updated",
  resultConnector: "Result Connector",
  resultTable: "Result Table",
  executorConnector: "Executor Connector",
  provisionArgs: "Provision Args",
  intervalsIsoDuration: "Interval Duration",
  notificationChannels: "Notification Channels",
  displayName: "Display Name",
  retryAttempts: "Retry Attempts",
  retryDelaySeconds: "Retry Delay (seconds)",
  retryExponentialBackoff: "Retry Exponential Backoff",
  retryIfErrorMatches: "Retry If Error Matches",
  changeMode: "Change Mode",
  partitionsWatermarkField: "Partitions Watermark Field",
  partitionsConcurrencyLimit: "Partitions Concurrency Limit",
};

/**
 * Extracts structured entries from any V1Resource by walking its spec and state.
 * No per-type components needed; new fields automatically appear.
 */
export function mapResource(resource: V1Resource): DescribeEntry[] {
  const entries: DescribeEntry[] = [];

  const kindKeys = [
    "source",
    "model",
    "metricsView",
    "explore",
    "theme",
    "component",
    "canvas",
    "api",
    "connector",
    "report",
    "alert",
  ] as const;

  for (const key of kindKeys) {
    const wrapper = resource[key];
    if (!wrapper) continue;

    const w = wrapper as Record<string, unknown>;
    if (w.spec) {
      flatten(entries, "Spec", w.spec as Record<string, unknown>);
    }
    if (w.state) {
      flatten(entries, "State", w.state as Record<string, unknown>);
    }

    if (!w.spec && !w.state) {
      flatten(entries, "Spec", w);
    }
    break;
  }

  // Resource metadata (only useful fields)
  const meta = resource.meta;
  if (meta) {
    push(entries, "Metadata", "Spec Updated", formatDate(meta.specUpdatedOn));
    push(entries, "Metadata", "State Updated", formatDate(meta.stateUpdatedOn));
    push(
      entries,
      "Metadata",
      "Reconcile Status",
      cleanEnum(meta.reconcileStatus),
    );
    if (meta.reconcileError) {
      push(entries, "Metadata", "Reconcile Error", meta.reconcileError);
    }
    push(entries, "Metadata", "Next Reconcile", formatDate(meta.reconcileOn));
    if (meta.refs?.length) {
      push(
        entries,
        "Metadata",
        "References",
        meta.refs.map((ref) => `${ref.kind}/${ref.name}`).join(", "),
        true,
      );
    }
  }

  return entries;
}

function flatten(
  entries: DescribeEntry[],
  section: string,
  obj: Record<string, unknown>,
  prefix = "",
) {
  for (const [key, val] of Object.entries(obj)) {
    if (val === undefined || val === null || val === "") continue;
    if (SKIP_KEYS.has(key)) continue;

    const label = prefix ? `${prefix}.${prettyLabel(key)}` : prettyLabel(key);

    // Multiline code fields (SQL, expressions)
    if (MULTILINE_KEYS.has(key) && typeof val === "string") {
      entries.push({
        section,
        label,
        value: val,
        mono: true,
        multiline: true,
      });
      continue;
    }

    if (Array.isArray(val)) {
      if (val.length === 0) continue;

      if (val.every((v) => typeof v !== "object" || v === null)) {
        push(entries, section, label, val.join(", "), true);
        continue;
      }

      for (let i = 0; i < val.length; i++) {
        const item = val[i];
        if (typeof item === "object" && item !== null) {
          flatten(
            entries,
            section,
            item as Record<string, unknown>,
            `${label}[${i}]`,
          );
        } else {
          push(entries, section, `${label}[${i}]`, String(item), true);
        }
      }
    } else if (typeof val === "object") {
      flatten(entries, section, val as Record<string, unknown>, label);
    } else if (typeof val === "boolean") {
      push(entries, section, label, val ? "Yes" : "No");
    } else if (key.endsWith("On") || key === "refreshedOn") {
      push(entries, section, label, formatDate(String(val)));
    } else {
      push(entries, section, label, String(val), shouldBeMono(key));
    }
  }
}

function push(
  entries: DescribeEntry[],
  section: string,
  label: string,
  value: string | undefined | null,
  mono = false,
) {
  if (value === undefined || value === null || value === "") return;
  entries.push({ section, label, value, mono });
}

function prettyLabel(key: string): string {
  if (LABEL_MAP[key]) return LABEL_MAP[key];
  return key
    .replace(/([A-Z])/g, " $1")
    .replace(/^./, (s) => s.toUpperCase())
    .trim();
}

function formatDate(date: string | undefined): string {
  if (!date) return "";
  try {
    return new Date(date).toLocaleString();
  } catch {
    return date;
  }
}

function cleanEnum(val: string | undefined): string {
  if (!val) return "";
  return val
    .replace(/^[A-Z_]+_STATUS_/, "")
    .replace(/^[A-Z_]+_MODE_/, "")
    .replace(/^[A-Z_]+_GRAIN_/, "")
    .toLowerCase()
    .replace(/^./, (s) => s.toUpperCase());
}

function shouldBeMono(key: string): boolean {
  const monoKeys = ["sql", "connector", "table", "column", "name", "driver", "path", "resolver"];
  return monoKeys.some((mk) => key.toLowerCase().includes(mk));
}
