import { Document } from "yaml";

/**
 * Pure example → component conversion: no runtime-client or I/O imports, so the
 * snapshot script (scripts/snapshot-vega-examples.ts) can run it under tsx to
 * precompute each gallery entry's `parameterizes` flag. The I/O side of the
 * import lives in example-to-component.ts.
 *
 * Importing an example converts it into a parameterized component: the example's
 * sample data is discarded, each encoded field becomes a typed param (dimension,
 * measure, or time_dimension, inferred from the encoding's type), a required
 * metrics_view param is declared, and both the generated metrics_sql and the Vega
 * encodings are templated with {{ .params.<name> }} references. The example
 * contributes only its visual design.
 *
 * Specs the deterministic conversion can't handle (composite views, multiple
 * datasets, repeat operators) fall back to a verbatim import with the sample data
 * inlined; those can be connected to project data in the workspace, manually or
 * with AI.
 */

/**
 * A gallery entry as listed in the index: everything the dialog's grid needs,
 * without the spec. The spec is fetched per example on selection (see
 * loadExample), because inlined sample data dwarfs the rest of the entry.
 */
export interface GalleryEntry {
  name: string;
  title: string;
  description?: string;
  category: string;
  subcategory: string;
  // Static SVG thumbnail pre-rendered at snapshot time (scales via its viewBox).
  thumbnail?: string;
  // Whether the deterministic import parameterizes this example, or falls back
  // to a verbatim sample-data import. Precomputed at snapshot time so the grid
  // can show it without loading every spec.
  parameterizes: boolean;
}

/** A gallery entry with its Vega-Lite spec loaded; what the import paths consume. */
export interface GalleryExample extends GalleryEntry {
  // The Vega-Lite spec as a JSON string, with external data inlined as values.
  spec: string;
}

export interface GalleryIndex {
  license: string;
  examples: GalleryEntry[];
}

/** Lazily loads the gallery index (thumbnails and metadata, no specs). */
export async function loadGallery(): Promise<GalleryIndex> {
  const module = await import("./vega-examples-index.json");
  return module.default as unknown as GalleryIndex;
}

/**
 * Loads one example's spec on demand. Specs live in per-example files so the
 * gallery grid never downloads the inlined sample data of 90 examples at once;
 * the bundler code-splits them into individually fetched chunks.
 */
export async function loadExample(
  entry: GalleryEntry,
): Promise<GalleryExample> {
  const module = (await import(`./specs/${entry.name}.json`)) as {
    default: unknown;
  };
  return { ...entry, spec: JSON.stringify(module.default) };
}

/**
 * Converts a gallery example into component YAML. The spec is parameterized
 * deterministically when its encodings can be analyzed, and imported verbatim
 * (with sample data) otherwise. In both cases fixed sizes are normalized to fill
 * the container, since gallery examples assume a standalone fixed-size render.
 */
export function exampleComponentYAML(example: GalleryExample): string {
  const spec = JSON.parse(example.spec) as Record<string, unknown>;
  normalizeSize(spec);

  const parameterized = parameterizeExampleSpec(spec);

  const doc = new Document({
    type: "component",
    display_name: example.title,
    ...(example.description ? { description: example.description } : {}),
    ...(parameterized ? { params: parameterized.params } : {}),
    custom_chart: {
      ...(parameterized ? { metrics_sql: [parameterized.metricsSQL] } : {}),
      vega_spec: JSON.stringify(parameterized?.spec ?? spec, null, 2),
    },
  });
  // lineWidth 0 disables plain-scalar folding, keeping the templated SQL on one line.
  return doc.toString({ lineWidth: 0 });
}

type ExampleParamType = "dimension" | "measure" | "time_dimension";

interface ParameterizedExample {
  spec: Record<string, unknown>;
  params: Record<string, unknown>[];
  metricsSQL: string;
}

const COMPOSITE_SPEC_KEYS = [
  "facet",
  "repeat",
  "concat",
  "hconcat",
  "vconcat",
  "spec",
];

/**
 * Attempts the deterministic example → parameterized component conversion.
 * Returns null when the spec is out of scope (composite views, multiple datasets,
 * no analyzable encodings), in which case the caller imports it verbatim.
 */
export function parameterizeExampleSpec(
  spec: Record<string, unknown>,
): ParameterizedExample | null {
  if (COMPOSITE_SPEC_KEYS.some((key) => key in spec)) return null;
  if (spec.datasets) return null;
  if (countDataNodes(spec) > 1) return null;

  // Collect encoded fields from the top-level encoding and single-level layers.
  const encodings: Record<string, unknown>[] = [];
  if (spec.encoding) encodings.push(spec.encoding as Record<string, unknown>);
  const layers = (spec.layer as Record<string, unknown>[]) ?? [];
  // Transforms (calculate/stack/filter/...) encode logic written against the
  // example's data values (specific field names, literal category strings), which
  // cannot be remapped onto arbitrary user data. Such specs import verbatim instead.
  if (hasTransforms(spec) || layers.some(hasTransforms)) return null;
  for (const layer of layers) {
    if (layer.layer || layer.data) return null; // Nested layers / per-layer data: out of scope.
    if (layer.encoding)
      encodings.push(layer.encoding as Record<string, unknown>);
  }

  // Map each encoded field to the param that will replace it. Params are named
  // after the chart role (the encoding channel), not the example's column names.
  const fieldInfo = new Map<
    string,
    { type: ExampleParamType; channel: string }
  >();
  for (const encoding of encodings) {
    for (const [channelName, channelDef] of Object.entries(encoding)) {
      for (const def of Array.isArray(channelDef) ? channelDef : [channelDef]) {
        if (!def || typeof def !== "object") continue;
        const channel = def as Record<string, unknown>;
        const field = channel.field;
        if (typeof field !== "string") continue;
        const paramType = classifyEncodedField(channel);
        // First channel names the param; when channels disagree on the same
        // field's type, measure/time_dimension win over plain dimension reads.
        const existing = fieldInfo.get(field);
        if (!existing) {
          fieldInfo.set(field, { type: paramType, channel: channelName });
        } else if (existing.type === "dimension" && paramType !== "dimension") {
          existing.type = paramType;
        }
      }
    }
  }
  if (fieldInfo.size === 0) return null;

  // Declare params: a required metrics_view plus one typed param per encoded field.
  // No defaults: bindings are chosen per dashboard (and auto-filled in previews).
  const params: Record<string, unknown>[] = [
    { name: "metrics_view", type: "metrics_view", required: true },
  ];
  const usedNames = new Set(["metrics_view"]);
  const fieldToParam = new Map<string, string>();
  for (const [field, info] of fieldInfo) {
    const paramName = uniqueParamName(
      channelParamName(info.channel),
      usedNames,
    );
    usedNames.add(paramName);
    params.push({
      name: paramName,
      type: info.type,
      required: true,
      description: `Field for the "${info.channel}" encoding (the example used "${field}")`,
    });
    fieldToParam.set(field, paramName);
  }

  // Rewrite the spec: drop sample data, template all field references.
  const rewritten = rewriteSpec(spec, fieldToParam) as Record<string, unknown>;
  rewritten.data = { name: "query1" };

  // Generate the templated query over the declared params. Time series order by
  // time (structural: the line scrambles otherwise); everything else sorts by a
  // customizable order_by param that the editor defaults to the measure. The row
  // limit is a customizable number param whose default is sized to the chart's
  // visual density (dense marks stay readable with more rows).
  const columns = [...fieldToParam.values()].map(
    (param) => `{{ .params.${param} }}`,
  );
  const timeParam = params.find((param) => param.type === "time_dimension");
  const defaultLimit = timeParam
    ? 2000
    : DENSE_MARKS.has(markType(spec) ?? "")
      ? 100
      : 10;
  let orderClause: string;
  if (timeParam) {
    orderClause = `ORDER BY {{ .params.${timeParam.name} }}`;
  } else {
    const orderByParam = uniqueParamName("order_by", usedNames);
    usedNames.add(orderByParam);
    params.push({
      name: orderByParam,
      type: "string",
      required: true,
      description: "Field to sort the query by (defaults to the measure)",
    });
    orderClause = `ORDER BY {{ .params.${orderByParam} }} DESC`;
  }
  const limitParam = uniqueParamName("limit", usedNames);
  usedNames.add(limitParam);
  params.push({
    name: limitParam,
    type: "number",
    default: defaultLimit,
    description: "Maximum number of rows to plot",
  });
  const metricsSQL = `SELECT ${columns.join(", ")} FROM {{ .params.metrics_view }} ${orderClause} LIMIT {{ .params.${limitParam} }}`;

  return { spec: rewritten, params, metricsSQL };
}

function hasTransforms(node: Record<string, unknown>): boolean {
  return Array.isArray(node.transform) && node.transform.length > 0;
}

// Marks that draw one small glyph per row and stay readable with many rows;
// they get a Top-100 limit instead of the Top-10 used for bars/arcs/text.
const DENSE_MARKS = new Set(["point", "circle", "square", "rect", "tick"]);

function markType(spec: Record<string, unknown>): string | undefined {
  const mark =
    spec.mark ?? (spec.layer as Record<string, unknown>[])?.[0]?.mark;
  if (typeof mark === "string") return mark;
  const type = (mark as Record<string, unknown> | undefined)?.type;
  return typeof type === "string" ? type : undefined;
}

// Param names by encoding channel, describing the field's role in the chart.
const CHANNEL_PARAM_NAMES: Record<string, string> = {
  x: "x_axis",
  y: "y_axis",
  x2: "x_axis_end",
  y2: "y_axis_end",
  theta: "angle",
  theta2: "angle_end",
  radius: "radius",
  color: "color",
  fill: "color",
  stroke: "stroke",
  size: "size",
  shape: "shape",
  opacity: "opacity",
  text: "label",
  detail: "detail",
  order: "order",
  tooltip: "tooltip",
  row: "row",
  column: "column",
  xOffset: "x_offset",
  yOffset: "y_offset",
};

function channelParamName(channel: string): string {
  return CHANNEL_PARAM_NAMES[channel] ?? sanitizeParamName(channel);
}

/** Maps a Vega-Lite encoding channel to the param type of its field. */
function classifyEncodedField(
  channel: Record<string, unknown>,
): ExampleParamType {
  // A timeUnit always reads a temporal field, whatever the declared axis type
  // (calendar heatmaps encode ordinal day/month axes derived from one date field).
  if (channel.timeUnit) return "time_dimension";
  switch (channel.type) {
    case "quantitative":
      return "measure";
    case "temporal":
      return "time_dimension";
    case "nominal":
    case "ordinal":
      return "dimension";
    default:
      // No explicit type: aggregated or binned channels read numeric fields.
      return channel.aggregate || channel.bin ? "measure" : "dimension";
  }
}

// Param names the parser rejects (common component properties and Vega reserved words).
const RESERVED_PARAM_NAMES = new Set([
  "metrics_view",
  "title",
  "description",
  "datum",
  "item",
  "event",
  "parent",
]);

function sanitizeParamName(name: string): string {
  const sanitized = name
    .toLowerCase()
    .replace(/\W/g, "_")
    .replace(/^(\d)/, "_$1");
  if (!sanitized) return "field";
  return RESERVED_PARAM_NAMES.has(sanitized) ? `${sanitized}_field` : sanitized;
}

function uniqueParamName(name: string, used: Set<string>): string {
  if (!used.has(name)) return name;
  for (let i = 2; ; i++) {
    const candidate = `${name}_${i}`;
    if (!used.has(candidate)) return candidate;
  }
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

/** Counts data definitions anywhere in the spec (multiple ⇒ out of scope). */
function countDataNodes(node: unknown): number {
  if (Array.isArray(node)) {
    return node.reduce<number>((sum, child) => sum + countDataNodes(child), 0);
  }
  if (!node || typeof node !== "object") return 0;
  let count = 0;
  for (const [key, value] of Object.entries(node)) {
    if (key === "data" && value && typeof value === "object") count++;
    count += countDataNodes(value);
  }
  return count;
}

/**
 * Recursively rewrites the spec, replacing references to example fields with
 * {{ .params.<name> }} placeholders: exact string matches (field definitions,
 * sort fields, titles) and `datum.<field>` references inside expressions.
 */
function rewriteSpec(
  node: unknown,
  fieldToParam: Map<string, string>,
  parentKey?: string,
): unknown {
  if (typeof node === "string") {
    // Values of these keys are Vega-Lite keywords, never field references; an
    // example field named e.g. "date" must not template `"timeUnit": "date"`.
    if (parentKey && KEYWORD_VALUED_KEYS.has(parentKey)) return node;
    const param = fieldToParam.get(node);
    if (param) return `{{ .params.${param} }}`;
    let result = node;
    for (const [field, paramName] of fieldToParam) {
      const datumRef = new RegExp(
        `datum(\\.${escapeRegExp(field)}\\b|\\['${escapeRegExp(field)}'\\])`,
        "g",
      );
      result = result.replace(datumRef, `datum['{{ .params.${paramName} }}']`);
    }
    return result;
  }
  if (Array.isArray(node)) {
    return node.map((child) => rewriteSpec(child, fieldToParam, parentKey));
  }
  if (node && typeof node === "object") {
    return Object.fromEntries(
      Object.entries(node).map(([key, value]) => [
        key,
        rewriteSpec(value, fieldToParam, key),
      ]),
    );
  }
  return node;
}

// Keys whose string values are Vega-Lite keywords rather than field references.
const KEYWORD_VALUED_KEYS = new Set([
  "timeUnit",
  "type",
  "aggregate",
  "op",
  "format",
  "formatType",
  "scheme",
  "interpolate",
  "order",
  "stack",
  "align",
  "baseline",
  "orient",
]);

const COMPOSITE_KEYS = [
  "hconcat",
  "vconcat",
  "concat",
  "facet",
  "repeat",
  "spec",
];

function normalizeSize(spec: Record<string, unknown>) {
  // Composite specs manage their own layout; leave them untouched.
  if (COMPOSITE_KEYS.some((key) => key in spec)) return;

  // Step-based sizes are part of the design (e.g. grouped bars); keep them.
  if (spec.width === undefined || typeof spec.width === "number") {
    spec.width = "container";
  }
  if (spec.height === undefined || typeof spec.height === "number") {
    spec.height = "container";
  }
  if (spec.width === "container" && spec.height === "container") {
    spec.autosize = { type: "fit", contains: "padding" };
  }
}
