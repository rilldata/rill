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
  defaultLimit: number;
}

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
  const params: Record<string, unknown>[] = [
    { name: "metrics_view", type: "metrics_view", required: true },
  ];

  const encodings: Record<string, unknown> = {};
  const selected: string[] = [];

  for (const binding of entry.bindings) {
    params.push({
      name: binding.param,
      type: binding.role,
      required: true,
      description: `Field shown on the ${binding.channel} channel.`,
    });
    encodings[binding.channel] = { field: templateRef(binding.param) };
    selected.push(templateRef(binding.param));
  }

  params.push({ name: "limit", type: "number", default: entry.defaultLimit });

  const doc = new Document({
    type: "component",
    display_name: entry.chartType,
    ...(entry.description ? { description: entry.description } : {}),
    params,
    custom_chart: {
      metrics_sql: buildMetricsSQL(entry, selected),
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

function templateRef(param: string): string {
  return `{{ .params.${param} }}`;
}

function buildMetricsSQL(entry: GalleryEntry, selected: string[]): string {
  const timeBinding = entry.bindings.find(
    (binding) => binding.role === "time_dimension",
  );
  const measureBinding = entry.bindings.find(
    (binding) => binding.role === "measure",
  );

  // Time series read left to right, so keep them in chronological order. Everything else is a
  // ranking, where the LIMIT should keep the largest rows.
  const orderBy = timeBinding
    ? `ORDER BY ${templateRef(timeBinding.param)}`
    : measureBinding
      ? `ORDER BY ${templateRef(measureBinding.param)} DESC`
      : "";

  return [
    `SELECT ${dedupe(selected).join(", ")}`,
    "FROM {{ .params.metrics_view }}",
    orderBy,
    "LIMIT {{ .params.limit }}",
  ]
    .filter(Boolean)
    .join(" ");
}

function dedupe(values: string[]): string[] {
  return [...new Set(values)];
}
