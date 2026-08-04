import { assembleVegaLite } from "flint-chart/vegalite";
import type { VisualizationSpec } from "svelte-vega";
import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import type { FlintFields } from "./semantic-types";

/**
 * Compiles a Flint chart spec into a Vega-Lite spec.
 *
 * Flint is a compiler, not a renderer: it derives scales, axes, formats, sorting, stacking and
 * layout from the data and the declared semantics, then emits a Vega-Lite spec. That spec goes
 * through the same VegaLiteRenderer, theme and measure formatters as the rest of Rill's charts.
 *
 * Flint's types are intentionally not imported: only values are re-exported from the
 * `flint-chart/vegalite` subpath, and the package is pre-1.0 and moving quickly, so the narrow
 * shapes this module actually depends on are declared locally.
 */

/** The complete Flint encoding vocabulary. */
export interface FlintEncoding {
  field?: string;
  type?: "quantitative" | "nominal" | "ordinal" | "temporal";
  aggregate?: "count" | "sum" | "average" | "mean";
  sortOrder?: "ascending" | "descending";
  sortBy?: string;
  scheme?: string;
}

/** A channel binds to a field name, an encoding object, or an array of either. */
export type FlintEncodingValue =
  | string
  | FlintEncoding
  | (string | FlintEncoding)[];

/** A Flint chart spec, as authored under `custom_chart.spec` in a component file. */
export interface FlintChartSpec {
  chartType?: string;
  encodings?: Record<string, FlintEncodingValue>;
  chartProperties?: Record<string, unknown>;
}

export interface FlintWarning {
  severity: "info" | "warning" | "error";
  code: string;
  message: string;
  channel?: string;
  field?: string;
}

export interface CompileOptions {
  spec: FlintChartSpec;
  rows: Record<string, unknown>[];
  fields: FlintFields;
  /** Measured size of the container, used as Flint's hard layout ceiling. */
  size?: { width: number; height: number };
  /** Measure names to route through Rill's `rill_<measure>` Vega expression formatters. */
  measureFields?: string[];
}

export interface CompileResult {
  spec: VisualizationSpec | null;
  warnings: FlintWarning[];
  error: string | null;
}

export function compileFlintSpec({
  spec,
  rows,
  fields,
  size,
  measureFields = [],
}: CompileOptions): CompileResult {
  if (!spec?.chartType) {
    return { spec: null, warnings: [], error: "No chart type provided" };
  }

  let assembled: Record<string, unknown>;
  try {
    assembled = assembleVegaLite({
      data: { values: rows },
      semantic_types: fields.semantic_types,
      field_display_names: fields.field_display_names,
      chart_spec: {
        chartType: spec.chartType,
        encodings: spec.encodings ?? {},
        // Flint treats baseSize as a target and canvasSize as a ceiling it may not exceed.
        // Passing the measured box as both pins the chart to its canvas slot.
        ...(size ? { baseSize: size, canvasSize: size } : {}),
        ...(spec.chartProperties
          ? { chartProperties: spec.chartProperties }
          : {}),
      },
      // Off by default in Flint.
      options: { addTooltips: true },
    }) as Record<string, unknown>;
  } catch (e) {
    return {
      spec: null,
      warnings: [],
      error: e instanceof Error ? e.message : String(e),
    };
  }

  const warnings = (assembled._warnings as FlintWarning[] | undefined) ?? [];
  const vegaSpec = stripInternalKeys(assembled);
  applyMeasureFormatters(vegaSpec, measureFields);

  return { spec: vegaSpec as VisualizationSpec, warnings, error: null };
}

/**
 * Removes Flint's non-standard bookkeeping keys (`_warnings`, `_width`, `_height`, `_options`,
 * `_transform`, `_pivot`). Vega-Lite rejects unknown top-level properties, so these must not reach
 * the renderer. Matching on the underscore prefix rather than a fixed list keeps this working as
 * Flint adds more of them.
 */
function stripInternalKeys(
  assembled: Record<string, unknown>,
): Record<string, unknown> {
  return Object.fromEntries(
    Object.entries(assembled).filter(([key]) => !key.startsWith("_")),
  );
}

/**
 * Routes measure axis and tooltip labels through Rill's registered `rill_<measure>` formatters.
 *
 * Flint picks a number format from the semantic type alone, so it has no way to honour a measure's
 * `format_d3` or currency/percent preset. Rill registers a Vega expression function per measure
 * (see Chart.svelte); pointing `formatType` at it restores measure formatting. Flint's own `format`
 * is removed where we override, since Vega-Lite would otherwise pass it to the formatter as an argument.
 */
function applyMeasureFormatters(
  spec: Record<string, unknown>,
  measureFields: string[],
): void {
  if (!measureFields.length) return;
  const measures = new Set(measureFields);

  const visit = (node: unknown): void => {
    if (Array.isArray(node)) {
      node.forEach(visit);
      return;
    }
    if (!node || typeof node !== "object") return;

    const record = node as Record<string, unknown>;
    if (
      typeof record.field === "string" &&
      record.type === "quantitative" &&
      measures.has(record.field)
    ) {
      const formatType = sanitizeFieldName(record.field);
      const axis = record.axis as Record<string, unknown> | undefined;
      if (axis) {
        delete axis.format;
        axis.formatType = formatType;
      }
      delete record.format;
      record.formatType = formatType;
    }

    Object.values(record).forEach(visit);
  };

  visit(spec);
}
