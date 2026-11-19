import { Document } from "yaml";

/**
 * Pure gallery entry → component conversion: no runtime-client or I/O imports, so it can run under
 * tsx and in unit tests. The I/O side of the import lives in example-to-component.ts.
 *
 * The gallery is Flint's own example set, generated from the installed package by
 * scripts/generate-flint-catalog.ts — a few curated examples per chart type, each showing a channel
 * layout that type supports. Importing an entry produces a fully parameterized component:
 * each channel the example binds becomes a typed param named after its role, a required
 * metrics_view param is declared, and both the metrics_sql and the Flint encodings reference them.
 *
 * The example contributes its chart type, channel layout, and any presentation tuning. Its sample
 * data is only used to render the thumbnail at build time and never reaches the component.
 */

export type ParamRole = "measure" | "dimension" | "time_dimension";

export interface ChannelBinding {
  channel: string;
  param: string;
  role: ParamRole;
}

/** One example in the gallery. */
export interface GalleryEntry {
  /** Unique key, also the gallery list key. */
  name: string;
  /** Slug of the chart type, used as the generated component's file name. */
  chartSlug: string;
  chartType: string;
  category: string;
  /** Flint's own label for the example. Not shown on the card; it feeds search and the tooltip. */
  title: string;
  description: string;
  tags: string[];
  bindings: ChannelBinding[];
  chartProperties?: Record<string, unknown>;
  /** The rendered SVG as a data URI, ready to drop into an img src. */
  thumbnail: string;
  /**
   * Whether the chart plots individual observations, so its component selects a `detail`
   * dimension that sets the grain of the rows without being drawn.
   */
  rowLevel: boolean;
  defaultLimit: number;
}

/** The unencoded dimension a row-level chart selects to get one row per observation. */
export const DETAIL_PARAM = "detail";

export interface GalleryIndex {
  flintVersion: string;
  examples: GalleryEntry[];
}

/** Lazily loads the generated gallery. */
export async function loadGallery(): Promise<GalleryIndex> {
  const module = await import("./flint-charts-index.json");
  return module.default as unknown as GalleryIndex;
}

/**
 * Builds a parameterized component from a gallery entry.
 *
 * Ordering matters for the generated SQL: the sort is by the time dimension ascending for time
 * series, or by the first measure descending otherwise, so the LIMIT keeps the rows the chart
 * should show rather than an arbitrary slice.
 */
export function exampleComponentYAML(entry: GalleryEntry): string {
  const encodings: Record<string, unknown> = {};
  for (const binding of entry.bindings) {
    encodings[binding.channel] = { field: templateRef(binding.param) };
  }
  const doc = new Document({
    type: "component",
    display_name: entry.chartType,
    ...(entry.description ? { description: entry.description } : {}),
    params: exampleParams(entry),
    custom_chart: {
      metrics_sql: buildMetricsSQL(entry),
      spec: {
        chartType: entry.chartType,
        encodings,
        ...(entry.chartProperties
          ? { chartProperties: entry.chartProperties }
          : {}),
      },
    },
  });

  // lineWidth 0 disables plain-scalar folding, keeping the templated SQL on one line.
  return doc.toString({ lineWidth: 0 });
}

/**
 * The params a component built from the entry declares: a required metrics_view, one required
 * field param per bound channel (typed by its role), the detail dimension of a row-level chart,
 * and the row limit.
 */
export function exampleParams(entry: GalleryEntry): Record<string, unknown>[] {
  return [
    { name: "metrics_view", type: "metrics_view", required: true },
    ...entry.bindings.map((binding) => ({
      name: binding.param,
      type: binding.role,
      required: true,
      description: `Field shown on the ${binding.channel} channel.`,
    })),
    ...(entry.rowLevel
      ? [
          {
            name: DETAIL_PARAM,
            type: "dimension",
            required: true,
            description:
              "Field that splits the data into one row per observation; it is not drawn.",
          },
        ]
      : []),
    {
      name: "limit",
      type: "number",
      default: entry.defaultLimit,
      description: "Maximum rows returned.",
    },
  ];
}

/** The entry's params as the `params:` block of a component file. */
export function exampleParamsYAML(entry: GalleryEntry): string {
  return new Document({ params: exampleParams(entry) }).toString({
    lineWidth: 0,
  });
}

function templateRef(param: string): string {
  return `{{ .params.${param} }}`;
}

/**
 * The component's metrics_sql. The binding roles are known here, so the ordering is fixed rather
 * than templated on the bound field's type: a conditional ORDER BY is easy to get wrong in the
 * metrics SQL dialect, which has no CASE.
 */
export function buildMetricsSQL(entry: GalleryEntry): string {
  const selected = entry.bindings.map((binding) => templateRef(binding.param));
  if (entry.rowLevel) selected.push(templateRef(DETAIL_PARAM));

  const timeBinding = entry.bindings.find(
    (binding) => binding.role === "time_dimension",
  );
  const measureBinding = entry.bindings.find(
    (binding) => binding.role === "measure",
  );

  // Time series read left to right, so keep them in chronological order, largest series first
  // within each bucket so a LIMIT cut drops the smallest ones. Everything else is a ranking, where
  // the LIMIT should keep the largest rows.
  const orderBy = [
    ...(timeBinding ? [templateRef(timeBinding.param)] : []),
    ...(measureBinding ? [`${templateRef(measureBinding.param)} DESC`] : []),
  ];

  return [
    `SELECT ${dedupe(selected).join(", ")}`,
    "FROM {{ .params.metrics_view }}",
    orderBy.length ? `ORDER BY ${orderBy.join(", ")}` : "",
    "LIMIT {{ .params.limit }}",
  ]
    .filter(Boolean)
    .join(" ");
}

function dedupe(values: string[]): string[] {
  return [...new Set(values)];
}
