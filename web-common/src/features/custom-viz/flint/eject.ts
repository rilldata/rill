import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import type { ComponentParamType } from "@rilldata/web-common/features/custom-viz/params";

/**
 * Ejects a compiled chart spec into a parameterized Vega-Lite spec.
 *
 * A chart spec states a chart type and which field goes on which channel; the compiler turns that
 * into Vega-Lite by reading the data and the metrics view's semantics. Ejecting hands the compiled
 * Vega-Lite back to the author so they can edit it directly, at the cost of losing the compiler's
 * data-driven adaptation.
 *
 * The compiled spec has the component's param bindings resolved into it: field names in encodings,
 * transforms and expressions; display names in axis and legend titles; measure formatter names in
 * `formatType`. Writing that verbatim into a component file would hard-code one dashboard's
 * bindings into a component that is meant to be reused, so the resolved values are turned back into
 * template references before the spec is written:
 *
 *   field name     -> {{ .params.<param> }}
 *   display name   -> {{ .fields.<param>.display_name }}
 *   formatter name -> {{ .fields.<param>.format_type }}
 *
 * Field names are matched anywhere in a string, since Vega expressions embed them
 * (`datum['bid_price_sum']`), while display names are only matched as a whole title value whose
 * encoding binds the matching field, so a hard-coded label that happens to read like a display
 * name is left alone.
 */

/** The name the renderer registers a component's query result under. */
export const EJECTED_DATA_NAME = "query1";

const VEGA_LITE_SCHEMA = "https://vega.github.io/schema/vega-lite/v5.json";

/**
 * Delimits a parked match while substituting (see `substitute`). A private-use code point can
 * occur in neither a field name, a display name, nor any literal the compiler emits, so expanding
 * a placeholder back cannot catch unrelated text.
 */
const PARK_DELIMITER = "\uE000";
const PARKED_PATTERN = new RegExp(
  `${PARK_DELIMITER}(\\d+)${PARK_DELIMITER}`,
  "g",
);

/** Object keys whose value is a label rather than data, and so may hold a display name. */
const TITLE_KEYS = new Set(["title", "text"]);

/**
 * Multi-view specs size their subplots from the composition rather than from the enclosing view,
 * so their baked view size is left alone.
 */
const MULTI_VIEW_KEYS = [
  "facet",
  "repeat",
  "concat",
  "hconcat",
  "vconcat",
] as const;

/** A field-typed param binding to turn back into a template reference. */
export interface EjectFieldBinding {
  /** Name of the param the value is bound through. */
  param: string;
  /** The metrics view field the param is bound to. */
  field: string;
  /** The field's display name, as it was handed to the compiler. */
  displayName?: string;
  /** The param's declared type; only measures have a Rill formatter registered for them. */
  type: ComponentParamType;
}

export function ejectToVegaSpec(
  compiled: Record<string, unknown>,
  bindings: EjectFieldBinding[],
): string {
  const replacements = buildReplacements(bindings);
  const spec: Record<string, unknown> = {};

  for (const [key, value] of Object.entries(compiled)) {
    // The compiler inlines the rows it read, which the renderer supplies as a named dataset instead.
    // Skipping it also keeps cell values that happen to match a field name from being templated.
    if (key === "data") continue;
    spec[key] = templatize(value, undefined, key, bindings, replacements);
  }

  // Ordered so the spec opens with the keys a reader orients by, matching the base specs the
  // native charts are built from (see features/components/charts/builder.ts).
  return JSON.stringify(
    {
      $schema: VEGA_LITE_SCHEMA,
      ...takeContainerSizing(spec),
      data: { name: EJECTED_DATA_NAME },
      ...spec,
    },
    null,
    2,
  );
}

/**
 * Builds the field-name and formatter-name substitutions, longest needle first.
 *
 * Ordering matters because one needle can contain another: a measure's formatter name
 * (`rill_bid_price`) contains its field name, and one field name can be a prefix of another
 * (`price`, `price_sum`). Matching the longest first, and never rescanning what was already
 * substituted, keeps the shorter needle from splitting a longer match.
 */
function buildReplacements(
  bindings: EjectFieldBinding[],
): { pattern: RegExp; template: string }[] {
  const needles: { needle: string; template: string }[] = [];
  for (const { param, field, type } of bindings) {
    if (!field) continue;
    if (type === "measure") {
      needles.push({
        needle: sanitizeFieldName(field),
        template: `{{ .fields.${param}.format_type }}`,
      });
    }
    needles.push({ needle: field, template: `{{ .params.${param} }}` });
  }
  needles.sort((a, b) => b.needle.length - a.needle.length);

  return needles.map(({ needle, template }) => ({
    // Bounded on both sides so a field name is not matched inside a longer identifier.
    // The leading boundary is a capture rather than a lookbehind for broader browser support.
    pattern: new RegExp(
      `(^|[^A-Za-z0-9_$])${escapeRegExp(needle)}(?![A-Za-z0-9_$])`,
      "g",
    ),
    template,
  }));
}

function templatize(
  value: unknown,
  scopeField: string | undefined,
  key: string | undefined,
  bindings: EjectFieldBinding[],
  replacements: { pattern: RegExp; template: string }[],
): unknown {
  if (typeof value === "string") {
    if (key && TITLE_KEYS.has(key)) {
      const binding = bindings.find(
        (candidate) =>
          candidate.field === scopeField && candidate.displayName === value,
      );
      if (binding) return `{{ .fields.${binding.param}.display_name }}`;
    }
    return substitute(value, replacements);
  }

  if (Array.isArray(value)) {
    return value.map((entry) =>
      templatize(entry, scopeField, key, bindings, replacements),
    );
  }

  if (value && typeof value === "object") {
    const record = value as Record<string, unknown>;
    // An encoding's channel properties (axis, legend, header) nest below the object that names the
    // field, so the field in scope is inherited until a nested object names its own.
    const nextScope =
      typeof record.field === "string" ? record.field : scopeField;
    return Object.fromEntries(
      Object.entries(record).map(([childKey, childValue]) => [
        childKey,
        templatize(childValue, nextScope, childKey, bindings, replacements),
      ]),
    );
  }

  return value;
}

/**
 * Applies the substitutions to a string in one pass.
 *
 * Each match is parked behind a placeholder that no later pattern can match, so a template
 * reference is never itself substituted: `{{ .params.publisher }}` inserted for the field
 * `publisher` would otherwise be rewritten again by a param of the same name.
 */
function substitute(
  value: string,
  replacements: { pattern: RegExp; template: string }[],
): string {
  const parked: string[] = [];
  let res = value;
  for (const { pattern, template } of replacements) {
    res = res.replace(pattern, (_, boundary: string) => {
      parked.push(template);
      return `${boundary}${PARK_DELIMITER}${parked.length - 1}${PARK_DELIMITER}`;
    });
  }
  return res.replace(
    PARKED_PATTERN,
    (match, index: string) => parked[Number(index)] ?? match,
  );
}

/**
 * Replaces the compiler's fixed sizing with container sizing, matching the base specs Rill's native
 * charts are built from: `width`/`height` of "container" and `autosize: {type: "fit"}`.
 *
 * The compiler sizes charts itself: handed the box it should draw into, it writes back a plot area
 * that already leaves room for axes and legends. A static spec cannot redo that arithmetic when the
 * container changes, so it hands the fitting to Vega-Lite instead. `autosize: "fit"` is the part
 * that matters — it makes the given dimensions the chart's *total* size rather than its plot area,
 * which keeps axis and legend decorations inside the container. Without it the Rill theme's
 * "fit-x" default applies, which fits the width only and lets the chart run past the bottom of
 * its box.
 *
 * Multi-view specs (facet, repeat, concat) size their subplots from the composition, and Vega-Lite
 * rejects container sizing for them, so theirs is left as the compiler wrote it.
 *
 * Mutates the spec to drop the sizing it replaces, and returns the keys to write in its place.
 */
function takeContainerSizing(
  spec: Record<string, unknown>,
): Record<string, unknown> {
  if (MULTI_VIEW_KEYS.some((key) => key in spec)) return {};

  // Includes a bar chart's `{step: 40}` band width: the renderer already sizes over it, and the
  // point of ejecting is a chart that fills whatever container it lands in.
  delete spec.width;
  delete spec.height;

  const view = (spec.config as Record<string, unknown> | undefined)?.view as
    | Record<string, unknown>
    | undefined;
  if (view) {
    delete view.continuousWidth;
    delete view.continuousHeight;
    if (Object.keys(view).length === 0) {
      delete (spec.config as Record<string, unknown>).view;
    }
  }

  return { width: "container", height: "container", autosize: { type: "fit" } };
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
