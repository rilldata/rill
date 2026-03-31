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
  // Canvas rows handled by flattenCanvasComponents
  "rows",
  // Noisy internal fields
  "cacheKeyTtlSeconds",
  // Already shown in Metadata section
  "refreshedOn",
  "reconcileError",
  "reconcileStatus",
  "reconcileOn",
  "specUpdatedOn",
  "stateUpdatedOn",
]);

// Fields that contain code and should render as multiline blocks
const MULTILINE_KEYS = new Set(["sql", "watermarkExpression"]);

// Array fields that should stay flat in the parent section (not promoted to cards)
const FLAT_ARRAY_KEYS = new Set<string>([]);

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
export function mapResource(
  resource: V1Resource,
  allResources?: V1Resource[],
): DescribeEntry[] {
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

    if (!w.spec) {
      flatten(entries, "Spec", w);
    }

    // Canvas: extract components from rows into a "Components" section
    if (key === "canvas") {
      flattenCanvasComponents(
        entries,
        w.spec as Record<string, unknown>,
        allResources,
      );
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

/** Extract canvas components from rows into a readable "Components" section */
function flattenCanvasComponents(
  entries: DescribeEntry[],
  spec: Record<string, unknown> | undefined,
  allResources?: V1Resource[],
) {
  if (!spec) return;
  const rows = spec.rows as Array<Record<string, unknown>> | undefined;
  if (!rows?.length) return;

  // Build lookup for component resources by name
  const componentMap = new Map<string, V1Resource>();
  if (allResources) {
    for (const r of allResources) {
      if (
        r.meta?.name?.name &&
        (r.component || r.meta.name.kind === "rill.runtime.v1.Component")
      ) {
        componentMap.set(r.meta.name.name, r);
      }
    }
  }

  for (let r = 0; r < rows.length; r++) {
    const row = rows[r];
    const items = row.items as Array<Record<string, unknown>> | undefined;
    if (!items?.length) continue;

    const height = row.height;
    const heightUnit = row.heightUnit || "px";

    for (const item of items) {
      const name = (item.component as string) || `Row ${r + 1}`;
      const group = name;
      const section = "Components";

      // Look up component resource for renderer and metrics info
      const compResource = componentMap.get(name);
      const compSpec = compResource?.component?.spec;

      if (compSpec?.renderer) {
        entries.push({
          section,
          label: "Renderer",
          value: compSpec.renderer,
          mono: true,
          group,
        });
      }
      if (compSpec?.rendererProperties) {
        const props = compSpec.rendererProperties as Record<string, unknown>;
        // Try both camelCase and snake_case for metrics view
        const metricsView = props.metricsView ?? props.metrics_view;
        if (metricsView) {
          entries.push({
            section,
            label: "Metrics View",
            value: String(metricsView),
            mono: true,
            group,
          });
        }
        // Show other renderer properties (skip large content blobs)
        const skipProps = new Set([
          "metricsView",
          "metrics_view",
          "content",
          "color",
        ]);
        for (const [pk, pv] of Object.entries(props)) {
          if (skipProps.has(pk)) continue;
          if (pv === undefined || pv === null || pv === "") continue;
          if (typeof pv === "object") continue;
          entries.push({
            section,
            label: prettyLabel(pk),
            value: String(pv),
            mono: shouldBeMono(pk),
            group,
          });
        }
      }

      entries.push({ section, label: "Row", value: `${r + 1}`, group });
      if (height) {
        entries.push({
          section,
          label: "Row Height",
          value: `${height}${heightUnit}`,
          group,
        });
      }
      if (item.width) {
        const unit = (item.widthUnit as string) || "";
        entries.push({
          section,
          label: "Width",
          value: `${item.width}${unit}`,
          group,
        });
      }
    }
  }
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

      // Flat arrays stay in the parent section as indexed rows
      if (FLAT_ARRAY_KEYS.has(key)) {
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
        continue;
      }

      // Promote arrays of objects to their own section with per-item groups
      const arraySection = prefix ? `${prefix} > ${label}` : label;
      for (let i = 0; i < val.length; i++) {
        const item = val[i];
        if (typeof item === "object" && item !== null) {
          const obj = item as Record<string, unknown>;
          const groupName =
            (typeof obj.name === "string" && obj.name) ||
            (typeof obj.displayName === "string" && obj.displayName) ||
            "";
          if (groupName) {
            flattenGrouped(entries, arraySection, obj, groupName);
          } else {
            // No name — flatten directly into the section (no collapsible wrapper)
            flatten(entries, arraySection, obj);
          }
        } else {
          push(entries, arraySection, `${i + 1}`, String(item), true);
        }
      }
    } else if (typeof val === "object") {
      flatten(entries, section, val as Record<string, unknown>, label);
    } else if (typeof val === "boolean") {
      push(entries, section, label, val ? "Yes" : "No");
    } else if (key.endsWith("On") || key === "refreshedOn") {
      push(entries, section, label, formatDate(String(val)));
    } else {
      const formatted = formatValue(key, val);
      if (formatted !== null) {
        push(entries, section, label, formatted, shouldBeMono(key));
      } else {
        push(entries, section, label, String(val), shouldBeMono(key));
      }
    }
  }
}

// Preferred field order for grouped items; unlisted keys appear before BOTTOM_KEYS
const FIELD_ORDER = [
  "name",
  "displayName",
  "column",
  "expression",
  "sql",
  "renderer",
  "type",
  "smallestTimeGrain",
  "description",
  "label",
  "formatPreset",
  "formatD3",
];
const BOTTOM_KEYS = new Set(["unnest", "uri"]);

/** Flatten an object's fields into grouped entries (for array items like measures/dimensions) */
function flattenGrouped(
  entries: DescribeEntry[],
  section: string,
  obj: Record<string, unknown>,
  group: string,
) {
  const sortedKeys = Object.keys(obj).sort((a, b) => {
    const aBottom = BOTTOM_KEYS.has(a);
    const bBottom = BOTTOM_KEYS.has(b);
    if (aBottom !== bBottom) return aBottom ? 1 : -1;
    const aIdx = FIELD_ORDER.indexOf(a);
    const bIdx = FIELD_ORDER.indexOf(b);
    if (aIdx !== -1 && bIdx !== -1) return aIdx - bIdx;
    if (aIdx !== -1) return -1;
    if (bIdx !== -1) return 1;
    return 0;
  });

  for (const key of sortedKeys) {
    const val = obj[key];
    if (val === undefined || val === null || val === "") continue;
    if (SKIP_KEYS.has(key)) continue;

    const label = prettyLabel(key);

    if (MULTILINE_KEYS.has(key) && typeof val === "string") {
      entries.push({
        section,
        label,
        value: val,
        mono: true,
        multiline: true,
        group,
      });
      continue;
    }

    if (Array.isArray(val)) {
      if (val.length === 0) continue;
      if (val.every((v) => typeof v !== "object" || v === null)) {
        entries.push({
          section,
          label,
          value: val.join(", "),
          mono: true,
          group,
        });
      }
      // Skip nested arrays of objects within grouped items to avoid deep nesting
      continue;
    }

    if (typeof val === "object") {
      // Inline shallow nested objects as "Parent.Child" labels
      for (const [subKey, subVal] of Object.entries(
        val as Record<string, unknown>,
      )) {
        if (subVal === undefined || subVal === null || subVal === "") continue;
        if (SKIP_KEYS.has(subKey)) continue;
        entries.push({
          section,
          label: `${label}.${prettyLabel(subKey)}`,
          value: String(subVal),
          mono: shouldBeMono(subKey),
          group,
        });
      }
      continue;
    }

    if (typeof val === "boolean") {
      entries.push({ section, label, value: val ? "Yes" : "No", group });
    } else {
      const formatted = formatValue(key, val);
      const display = formatted !== null ? formatted : String(val);
      if (display) {
        entries.push({
          section,
          label,
          value: display,
          mono: shouldBeMono(key),
          group,
        });
      }
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
  // Strip common proto enum prefixes
  const cleaned = val
    .replace(/^[A-Z_]+_STATUS_/, "")
    .replace(/^[A-Z_]+_MODE_/, "")
    .replace(/^[A-Z_]+_GRAIN_/, "")
    .replace(/^[A-Z_]+_TYPE_/, "");
  // Hide "unspecified" values entirely
  if (cleaned === "UNSPECIFIED") return "";
  return cleaned
    .toLowerCase()
    .replace(/_/g, " ")
    .replace(/^./, (s) => s.toUpperCase());
}

const DAY_NAMES = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];
const MONTH_NAMES = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];

/** Format values that need special handling based on field key */
function formatValue(key: string, val: unknown): string | null {
  const s = String(val);
  if (key === "firstDayOfWeek") {
    const idx = Number(val);
    return DAY_NAMES[idx] ?? s;
  }
  if (key === "firstMonthOfYear") {
    const idx = Number(val);
    return MONTH_NAMES[idx] ?? s;
  }
  // Clean up proto enum values
  if (typeof val === "string" && /^[A-Z_]{2,}_[A-Z_]+$/.test(val)) {
    return cleanEnum(val) || null;
  }
  return null;
}

function shouldBeMono(key: string): boolean {
  const monoKeys = [
    "sql",
    "connector",
    "table",
    "column",
    "name",
    "driver",
    "path",
    "resolver",
  ];
  return monoKeys.some((mk) => key.toLowerCase().includes(mk));
}
