import {
  clampColumnWidth,
  type ColumnWidthRole,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  type PivotColumnAlign,
  type PivotColumnStyles,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import type { SortingState } from "tanstack-table-8-svelte-5";
import type { PivotSpec, TableSpec } from "./index";

/**
 * An entry of a table's `columns` list or a pivot's `measures`,
 * `row_dimensions`, or `col_dimensions` list: the field name alone, or an
 * object with the name plus per-column presentation overrides.
 */
// A measure's comparison columns, shown while the component compares.
export type PivotComparison = "delta" | "percent_change";

export const PIVOT_COMPARISONS: readonly PivotComparison[] = [
  "delta",
  "percent_change",
];

/** Overrides for a measure's delta or percent-change column. */
export interface PivotComparisonColumnConfig {
  // Pixel width within the measure bounds.
  width?: number;
}

export interface PivotFieldConfig {
  name: string;
  // Pixel width within the resizer bounds of the column's role.
  width?: number;
  // Wrap cell text instead of truncating it. Dimension columns only.
  wrap?: boolean;
  align?: PivotColumnAlign;
  // Header text, replacing the metrics view display name.
  label?: string;
  // Number format overrides for measures; mutually exclusive.
  format_preset?: string;
  format_d3?: string;
  // The measure's comparison columns. Measures of the metrics view only, as
  // adhoc measures get no comparison columns.
  delta?: PivotComparisonColumnConfig;
  percent_change?: PivotComparisonColumnConfig;
}

export type PivotFieldEntry = string | PivotFieldConfig;

export type PivotFieldListKey =
  | "columns"
  | "measures"
  | "row_dimensions"
  | "col_dimensions";

// A patch for one entry's overrides; null removes the key.
export type PivotComparisonColumnPatch = {
  [K in keyof PivotComparisonColumnConfig]?:
    | PivotComparisonColumnConfig[K]
    | null;
};

// A null removes the key; a comparison patch merges into the nested object.
export type PivotFieldConfigPatch = {
  [K in keyof Omit<PivotFieldConfig, "name" | PivotComparison>]?:
    | PivotFieldConfig[K]
    | null;
} & {
  [K in PivotComparison]?: PivotComparisonColumnPatch | null;
};

export type PivotSortDir = "asc" | "desc";

export const PIVOT_SORT_DIRS: readonly PivotSortDir[] = ["asc", "desc"];

/** The column id suffix the pivot gives a measure's comparison column. */
export function comparisonColumnSuffix(comparison: PivotComparison): string {
  return comparison === "delta" ? COMPARISON_DELTA : COMPARISON_PERCENT;
}

/** Splits a comparison column id into its measure and comparison; undefined for other ids. */
export function splitComparisonColumnId(
  id: string,
): { measure: string; comparison: PivotComparison } | undefined {
  if (id.endsWith(COMPARISON_DELTA)) {
    return {
      measure: id.slice(0, -COMPARISON_DELTA.length),
      comparison: "delta",
    };
  }
  if (id.endsWith(COMPARISON_PERCENT)) {
    return {
      measure: id.slice(0, -COMPARISON_PERCENT.length),
      comparison: "percent_change",
    };
  }
  return undefined;
}

export const PIVOT_COLUMN_ALIGNS: readonly PivotColumnAlign[] = [
  "left",
  "center",
  "right",
];

export const WRAP_LINES_MIN = 1;
export const WRAP_LINES_MAX = 5;
export const WRAP_LINES_DEFAULT = 2;

// The flat override keys; the comparison objects are handled separately.
const PRESENTATION_KEYS = [
  "width",
  "wrap",
  "align",
  "label",
  "format_preset",
  "format_d3",
] as const satisfies readonly (keyof Omit<
  PivotFieldConfig,
  "name" | PivotComparison
>)[];

function hasOverrides(config: PivotFieldConfig): boolean {
  return (
    PRESENTATION_KEYS.some((key) => key in config) ||
    PIVOT_COMPARISONS.some((key) => key in config)
  );
}

/** The field name of an entry; "" for malformed hand-edited entries. */
export function fieldName(entry: PivotFieldEntry | undefined | null): string {
  if (typeof entry === "string") return entry;
  if (entry && typeof entry === "object" && typeof entry.name === "string") {
    return entry.name;
  }
  return "";
}

export function fieldNames(
  entries: PivotFieldEntry[] | undefined | null,
): string[] {
  if (!Array.isArray(entries)) return [];
  return entries.map(fieldName).filter((name) => name !== "");
}

/** The entry for a field as an object, with malformed override values dropped. */
export function fieldConfig(
  entries: PivotFieldEntry[] | undefined | null,
  name: string,
): PivotFieldConfig | undefined {
  if (!Array.isArray(entries)) return undefined;
  const entry = entries.find((e) => fieldName(e) === name);
  if (entry === undefined) return undefined;
  return sanitizeFieldConfig(entry);
}

function sanitizeFieldConfig(entry: PivotFieldEntry): PivotFieldConfig {
  if (typeof entry === "string") return { name: entry };
  const config: PivotFieldConfig = { name: entry.name };
  if (typeof entry.width === "number" && Number.isFinite(entry.width)) {
    config.width = entry.width;
  }
  if (typeof entry.wrap === "boolean") config.wrap = entry.wrap;
  if (PIVOT_COLUMN_ALIGNS.includes(entry.align as PivotColumnAlign)) {
    config.align = entry.align;
  }
  if (typeof entry.label === "string" && entry.label !== "") {
    config.label = entry.label;
  }
  if (typeof entry.format_preset === "string" && entry.format_preset !== "") {
    config.format_preset = entry.format_preset;
  }
  if (typeof entry.format_d3 === "string" && entry.format_d3 !== "") {
    config.format_d3 = entry.format_d3;
  }
  for (const comparison of PIVOT_COMPARISONS) {
    const nested = sanitizeComparisonColumnConfig(entry[comparison]);
    if (nested) config[comparison] = nested;
  }
  return config;
}

function sanitizeComparisonColumnConfig(
  raw: unknown,
): PivotComparisonColumnConfig | undefined {
  if (typeof raw !== "object" || raw === null) return undefined;
  const width = (raw as { width?: unknown }).width;
  if (typeof width === "number" && Number.isFinite(width)) return { width };
  return undefined;
}

/** All entries of the component's field lists as objects, keyed by name. */
export function fieldConfigs(
  spec: PivotSpec | TableSpec,
): Record<string, PivotFieldConfig> {
  const lists =
    "columns" in spec
      ? [spec.columns]
      : [spec.measures, spec.row_dimensions, spec.col_dimensions];
  const configs: Record<string, PivotFieldConfig> = {};
  for (const list of lists) {
    if (!Array.isArray(list)) continue;
    for (const entry of list) {
      const name = fieldName(entry);
      if (name !== "" && !(name in configs)) {
        configs[name] = sanitizeFieldConfig(entry);
      }
    }
  }
  return configs;
}

/**
 * The list after the inspector changed the selected names: entries that
 * survive keep their object (and its overrides) in the new order, new names
 * are added as plain strings.
 */
export function mergeFieldList(
  prev: PivotFieldEntry[] | undefined | null,
  nextNames: string[],
): PivotFieldEntry[] {
  const prevByName = new Map<string, PivotFieldEntry>();
  if (Array.isArray(prev)) {
    for (const entry of prev) {
      const name = fieldName(entry);
      if (name !== "" && !prevByName.has(name)) prevByName.set(name, entry);
    }
  }
  return nextNames.map((name) => prevByName.get(name) ?? name);
}

/**
 * The list with one entry's overrides patched. Keys set to null are removed,
 * and an entry collapses back to a plain string when only its name is left.
 * A name that is not in the list is appended.
 */
export function setFieldConfig(
  list: PivotFieldEntry[] | undefined | null,
  name: string,
  patch: PivotFieldConfigPatch,
): PivotFieldEntry[] {
  const entries = Array.isArray(list) ? [...list] : [];
  const index = entries.findIndex((e) => fieldName(e) === name);
  const current = index === -1 ? { name } : sanitizeFieldConfig(entries[index]);
  const next: PivotFieldConfig = { ...current };
  for (const key of PRESENTATION_KEYS) {
    if (!(key in patch)) continue;
    const value = patch[key];
    if (value === null || value === undefined) {
      delete next[key];
    } else {
      // Each key keeps its own type; the cast only widens the union.
      (next as unknown as Record<string, unknown>)[key] = value;
    }
  }
  for (const comparison of PIVOT_COMPARISONS) {
    if (!(comparison in patch)) continue;
    const value = patch[comparison];
    if (value === null || value === undefined) {
      delete next[comparison];
      continue;
    }
    const merged: PivotComparisonColumnConfig = { ...next[comparison] };
    if ("width" in value) {
      if (value.width === null || value.width === undefined) {
        delete merged.width;
      } else {
        merged.width = value.width;
      }
    }
    if (Object.keys(merged).length) next[comparison] = merged;
    else delete next[comparison];
  }
  const collapsed = hasOverrides(next) ? next : next.name;
  if (index === -1) entries.push(collapsed);
  else entries[index] = collapsed;
  return entries;
}

/** Column dimensions are header groups, so they take no width, wrap, or alignment. */
export function stripColumnDimensionKeys(
  entry: PivotFieldEntry,
): PivotFieldEntry {
  if (typeof entry === "string") return entry;
  return setFieldConfig([entry], entry.name, {
    width: null,
    wrap: null,
    align: null,
  })[0];
}

/**
 * Row dimensions share one merged row-header column keyed by the first entry:
 * only it takes `width` and `wrap`, and row dimensions never take `align`.
 * Applied on every editor write so reordering never produces YAML the
 * runtime rejects.
 */
export function normalizeRowDimensionEntries(
  entries: PivotFieldEntry[],
): PivotFieldEntry[] {
  return entries.map((entry, index) => {
    if (typeof entry !== "object" || entry === null) return entry;
    return setFieldConfig(
      [entry],
      entry.name,
      index === 0 ? { align: null } : { align: null, width: null, wrap: null },
    )[0];
  });
}

/** Which overrides the inspector offers for a chip in a given list. */
export interface ColumnSettingsCapabilities {
  width: boolean;
  wrap: boolean;
  align: boolean;
  format: boolean;
  // Widths of the delta and percent-change columns: measures of the metrics
  // view only, as adhoc measures get no comparison columns.
  comparison: boolean;
}

export function columnSettingsCapabilities(
  listKey: PivotFieldListKey,
  index: number,
  isMeasure: boolean,
  isAdhocMeasure = false,
): ColumnSettingsCapabilities {
  const comparison = isMeasure && !isAdhocMeasure;
  switch (listKey) {
    case "columns":
      return {
        width: true,
        wrap: !isMeasure,
        align: true,
        format: isMeasure,
        comparison,
      };
    case "measures":
      return {
        width: true,
        wrap: false,
        align: true,
        format: true,
        comparison,
      };
    case "row_dimensions":
      // Only the first row dimension renders a column of its own.
      return {
        width: index === 0,
        wrap: index === 0,
        align: false,
        format: false,
        comparison: false,
      };
    case "col_dimensions":
      return {
        width: false,
        wrap: false,
        align: false,
        format: false,
        comparison: false,
      };
  }
}

export function wrapLinesFromSpec(spec: PivotSpec | TableSpec): number {
  const lines = spec.wrap_lines;
  if (typeof lines !== "number" || !Number.isFinite(lines)) {
    return WRAP_LINES_DEFAULT;
  }
  return Math.min(WRAP_LINES_MAX, Math.max(WRAP_LINES_MIN, Math.round(lines)));
}

/**
 * The renderer's per-column styles. Table: every column, keyed by its name.
 * Pivot: the measures (one shared width per measure across all column groups)
 * and the first row dimension (the merged row-header column); column
 * dimensions and deeper row dimensions have no column of their own.
 */
export function resolveColumnStyles(
  spec: PivotSpec | TableSpec,
  isMeasure: (name: string) => boolean,
): PivotColumnStyles {
  const styles: PivotColumnStyles = {};
  const add = (
    entry: PivotFieldEntry,
    role: ColumnWidthRole,
    allowAlign: boolean,
  ) => {
    if (typeof entry === "string") return;
    const config = sanitizeFieldConfig(entry);
    const style: PivotColumnStyles[string] = {};
    if (config.width !== undefined) {
      style.width = clampColumnWidth(role, config.width);
    }
    if (role === "dimension" && config.wrap !== undefined) {
      style.wrap = config.wrap;
    }
    if (allowAlign && config.align !== undefined) style.align = config.align;
    if (Object.keys(style).length) styles[config.name] = style;
    if (role !== "measure") return;
    // The comparison columns are keyed by the measure name plus a suffix.
    for (const comparison of PIVOT_COMPARISONS) {
      const width = config[comparison]?.width;
      if (width === undefined) continue;
      styles[`${config.name}${comparisonColumnSuffix(comparison)}`] = {
        width: clampColumnWidth("measure", width),
      };
    }
  };
  if ("columns" in spec) {
    for (const entry of spec.columns ?? []) {
      const name = fieldName(entry);
      if (name !== "") {
        add(entry, isMeasure(name) ? "measure" : "dimension", true);
      }
    }
  } else {
    for (const entry of spec.measures ?? []) add(entry, "measure", true);
    // The row header is a flex cell with expand chevrons; it takes no alignment.
    const first = spec.row_dimensions?.[0];
    if (first !== undefined) add(first, "dimension", false);
  }
  return styles;
}

/**
 * The metrics view measures and dimensions with the component's `label` and
 * number format overrides applied, so the shared pivot column definitions
 * pick them up unchanged. A format override replaces the measure's format
 * fields as a set.
 */
export function applyFieldConfigToMetricsView(
  measures: MetricsViewSpecMeasure[],
  dimensions: MetricsViewSpecDimension[],
  spec: PivotSpec | TableSpec,
): {
  measures: MetricsViewSpecMeasure[];
  dimensions: MetricsViewSpecDimension[];
} {
  const configs = fieldConfigs(spec);
  const hasOverrides = Object.values(configs).some(
    (config) =>
      config.label !== undefined ||
      config.format_preset !== undefined ||
      config.format_d3 !== undefined,
  );
  if (!hasOverrides) return { measures, dimensions };
  return {
    measures: measures.map((measure) => {
      const config = configs[measure.name ?? ""];
      if (!config) return measure;
      const next = { ...measure };
      if (config.label !== undefined) next.displayName = config.label;
      if (config.format_preset !== undefined) {
        next.formatPreset = config.format_preset;
        delete next.formatD3;
        delete next.formatD3Locale;
      } else if (config.format_d3 !== undefined) {
        next.formatD3 = config.format_d3;
        delete next.formatPreset;
      }
      return next;
    }),
    dimensions: dimensions.map((dimension) => {
      const config =
        configs[dimension.name ?? ""] ?? configs[dimension.column ?? ""];
      if (!config || config.label === undefined) return dimension;
      return { ...dimension, displayName: config.label };
    }),
  };
}

/** Names `sort_by` may reference: table columns; pivot measures and row dimensions. */
export function sortableFieldNames(spec: PivotSpec | TableSpec): string[] {
  if ("columns" in spec) return fieldNames(spec.columns);
  return [...fieldNames(spec.measures), ...fieldNames(spec.row_dimensions)];
}

export function defaultSortDir(
  sortBy: string,
  isMeasure: (name: string) => boolean,
): PivotSortDir {
  return isMeasure(sortBy) ? "desc" : "asc";
}

/**
 * The initial tanstack sorting state for the spec's `sort_by`, `sort_dir` and
 * `sort_comparison`. Empty when there is no (valid) sort, so the pivot's own
 * default applies. In a pivot any row dimension sorts the row axis, whose
 * column is keyed by the first row dimension. A comparison sort targets the
 * measure's delta or percent-change column, which only exists while the
 * widget compares (`comparisonEnabled`); otherwise the measure's own value is
 * sorted.
 */
export function sortingFromSpec(
  spec: PivotSpec | TableSpec,
  isMeasure: (name: string) => boolean,
  comparisonEnabled = false,
): SortingState {
  const sortBy = spec.sort_by;
  if (typeof sortBy !== "string" || sortBy === "") return [];
  if (!sortableFieldNames(spec).includes(sortBy)) return [];
  const dir = PIVOT_SORT_DIRS.includes(spec.sort_dir as PivotSortDir)
    ? (spec.sort_dir as PivotSortDir)
    : defaultSortDir(sortBy, isMeasure);
  const isMeasureSort =
    "columns" in spec
      ? isMeasure(sortBy)
      : fieldNames(spec.measures).includes(sortBy);
  let id = sortBy;
  if (!("columns" in spec) && !isMeasureSort) {
    id = fieldNames(spec.row_dimensions)[0] ?? sortBy;
  }
  const comparison = spec.sort_comparison;
  if (
    isMeasureSort &&
    comparisonEnabled &&
    PIVOT_COMPARISONS.includes(comparison as PivotComparison)
  ) {
    id = `${sortBy}${comparisonColumnSuffix(comparison as PivotComparison)}`;
  }
  return [{ id, desc: dir === "desc" }];
}

export interface PivotSortSpec {
  sort_by: string | undefined;
  sort_dir: PivotSortDir | undefined;
  sort_comparison: PivotComparison | undefined;
}

/**
 * The spec keys for a sorting state the user produced by clicking a header,
 * including a measure's delta or percent-change column. Undefined when the
 * sorted column cannot be expressed in YAML (a measure under one
 * column-dimension value in a nested pivot).
 */
export function sortingToSpec(
  sorting: SortingState,
  spec: PivotSpec | TableSpec,
): PivotSortSpec | undefined {
  const first = sorting[0];
  if (!first) {
    return {
      sort_by: undefined,
      sort_dir: undefined,
      sort_comparison: undefined,
    };
  }
  const split = splitComparisonColumnId(first.id);
  const sortBy = split ? split.measure : first.id;
  if (!sortableFieldNames(spec).includes(sortBy)) return undefined;
  return {
    sort_by: sortBy,
    sort_dir: first.desc ? "desc" : "asc",
    sort_comparison: split?.comparison,
  };
}

/** A patch for one comparison column's width; null clears it. */
export function comparisonWidthPatch(
  comparison: PivotComparison,
  width: number | null,
): PivotFieldConfigPatch {
  return comparison === "delta"
    ? { delta: { width } }
    : { percent_change: { width } };
}

/** The inspector's per-chip column settings for one field list. */
export interface ColumnSettings {
  listKey: PivotFieldListKey;
  configs: Record<string, PivotFieldConfig>;
  // Adhoc measures get no comparison columns, so their chips offer no
  // comparison widths.
  adhocNames: ReadonlySet<string>;
  onChange: (name: string, patch: PivotFieldConfigPatch) => void;
}
