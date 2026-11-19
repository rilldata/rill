import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import { DETAIL_PARAM } from "@rilldata/web-common/features/custom-viz/examples/example-spec";
import {
  MetricsViewSpecDimensionType,
  type V1ComponentParam,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";

/**
 * Utilities for working with the declared params of standalone (custom viz) components.
 * A component declares typed params in its YAML (`params:`), and canvas items bind values
 * to them (`params:` on the item). See runtime/parser/parse_component.go for the source
 * of truth on param types and validation.
 */

/** Project folder where custom viz component files are created by convention. */
export const CUSTOM_VIZ_FOLDER = "viz";

export type ComponentParamType =
  | "string"
  | "number"
  | "boolean"
  | "metrics_view"
  | "measure"
  | "dimension"
  | "time_dimension";

/**
 * Reads the declared params from a component resource.
 * This is the single accessor for the generated-client field so consumers stay
 * insulated from the resource shape.
 */
export function getDeclaredParams(
  resource: V1Resource | null | undefined,
  allowUnvalidatedSpec = false,
): V1ComponentParam[] {
  const component = resource?.component;
  const spec =
    component?.state?.validSpec ??
    (allowUnvalidatedSpec ? component?.spec : undefined);
  return spec?.params ?? [];
}

/** Whether a component resource declares params (i.e. is a parameterized custom viz). */
export function hasDeclaredParams(
  resource: V1Resource | null | undefined,
  allowUnvalidatedSpec = false,
): boolean {
  return getDeclaredParams(resource, allowUnvalidatedSpec).length > 0;
}

/** Converts a snake_case param name into a human-friendly label. */
export function prettyParamLabel(name = ""): string {
  return name.replace(/_/g, " ").replace(/(^|\s)\w/g, (c) => c.toUpperCase());
}

/**
 * Name convention for the sort-field param that generated components declare:
 * a string param named "order_by" is rendered as a dropdown of the field values
 * currently bound to the component's other params, and defaults to the measure.
 */
export const ORDER_BY_PARAM_NAME = "order_by";

const FIELD_PARAM_TYPES = new Set<ComponentParamType>([
  "measure",
  "dimension",
  "time_dimension",
]);

/** Whether a declared param references a field of a metrics view. */
export function isFieldParam(param: V1ComponentParam): boolean {
  return FIELD_PARAM_TYPES.has(param.type as ComponentParamType);
}

/**
 * The metrics view bound to the component's first metrics_view param,
 * falling back to its default.
 */
export function boundMetricsViewName(
  params: V1ComponentParam[],
  args: Record<string, unknown>,
): string | undefined {
  const param = params.find((candidate) => candidate.type === "metrics_view");
  if (!param?.name) return undefined;
  const name = args[param.name] ?? param.default;
  return typeof name === "string" && name ? name : undefined;
}

/** Whether a declared param is the by-convention sort-field selector. */
export function isOrderByParam(param: V1ComponentParam): boolean {
  return (
    param.name === ORDER_BY_PARAM_NAME &&
    (!param.type || param.type === "string")
  );
}

/**
 * Choices for the order_by param: the field values currently bound to the
 * component's dimension/measure/time_dimension params.
 */
export function orderByOptions(
  params: V1ComponentParam[],
  args: Record<string, unknown>,
): { value: string; label: string }[] {
  const options: { value: string; label: string }[] = [];
  const seen = new Set<string>();
  for (const param of params) {
    if (!param.name || !isFieldParam(param)) continue;
    const value = args[param.name];
    if (typeof value !== "string" || !value || seen.has(value)) continue;
    seen.add(value);
    options.push({
      value,
      label: `${value} (${prettyParamLabel(param.name)})`,
    });
  }
  return options;
}

/**
 * Maps a declared param to the inspector input widget that edits it.
 * The mapping targets the existing ParamMapper switchboard, so generated
 * param forms render with the same widgets as built-in components.
 */
export function paramToInputParam(
  param: V1ComponentParam,
  bindings?: { params: V1ComponentParam[]; args: Record<string, unknown> },
): ComponentInputParam {
  const common = {
    label: prettyParamLabel(param.name),
    optional: !param.required,
    description: param.description,
  };
  // The by-convention order_by param selects among the currently bound field values
  // instead of taking free text.
  if (bindings && isOrderByParam(param)) {
    return {
      type: "select",
      ...common,
      meta: { options: orderByOptions(bindings.params, bindings.args) },
    };
  }
  switch (param.type as ComponentParamType) {
    case "metrics_view":
      return { type: "metrics", ...common };
    case "measure":
      return { type: "measure", ...common };
    case "dimension":
      return { type: "dimension", ...common };
    case "time_dimension":
      // A dimension picker restricted to the metrics view's time fields:
      // the primary time dimension plus any time-typed dimensions.
      return {
        type: "dimension",
        ...common,
        meta: { includeTime: true, timeFieldsOnly: true },
      };
    case "boolean":
      return {
        type: "boolean",
        ...common,
        meta: { defaultValue: param.default ?? false },
      };
    case "string":
    case "number":
      if (param.options?.length) {
        return {
          type: "select",
          ...common,
          meta: {
            options: param.options.map((option) => ({
              value: String(option),
              label: String(option),
            })),
            default: param.default,
            paramType: param.type,
          },
        };
      }
      return { type: param.type === "number" ? "number" : "text", ...common };
    default:
      return { type: "text", ...common };
  }
}

/**
 * The chart channel each param is drawn on, read from the `{{ .params.<name> }}` references in a
 * custom chart's spec encodings. Params not drawn on a channel are absent.
 */
export function paramChannels(
  rendererProperties: Record<string, unknown> | undefined,
): Record<string, string> {
  const spec = rendererProperties?.spec as
    | { encodings?: Record<string, unknown> }
    | undefined;
  const channels: Record<string, string> = {};
  for (const [channel, encoding] of Object.entries(spec?.encodings ?? {})) {
    for (const match of JSON.stringify(encoding).matchAll(
      /\{\{\s*\.params\.(\w+)\s*\}\}/g,
    )) {
      channels[match[1]] ??= channel;
    }
  }
  return channels;
}

/** Returns the first non-empty Metrics SQL query from the supported scalar/list shapes. */
export function normalizeMetricsSQL(value: unknown): string | undefined {
  if (typeof value === "string") return value.trim() ? value : undefined;
  if (Array.isArray(value)) {
    return value.find(
      (entry): entry is string =>
        typeof entry === "string" && entry.trim().length > 0,
    );
  }
  return undefined;
}

/** Channels that split a chart into series or panels; each value becomes a color or a panel. */
const SERIES_CHANNELS = new Set(["color", "group", "row", "column"]);

/** Channels whose measure is read against the chart's primary measure, on the same scale. */
const PAIRED_MEASURE_CHANNELS = new Set([
  "x2",
  "y2",
  "goal",
  "open",
  "high",
  "low",
  "close",
]);

/** Channels holding the two ends or sides of each mark, by chart type. */
const PAIR_CHANNELS: Record<string, Set<string>> = {
  "Ranged Dot Plot": new Set(["color"]),
  "Pyramid Chart": new Set(["color"]),
  // The two periods a slope connects.
  "Slope Chart": new Set(["x"]),
};

/** Number of rows a cardinality probe reads; a count at this cap means "at least this many". */
export const VALUE_COUNT_PROBE_LIMIT = 51;

/**
 * What a dimension param is for, which decides how many distinct values suit it: a `series` splits
 * the chart into a handful of colors or panels, a `pair` into exactly the two ends or sides of each
 * mark (or the two periods of a slope chart), a `category` lays out an axis, and the `detail` dimension of a row-level chart sets the
 * grain of its rows.
 */
export type DimensionRole = "detail" | "series" | "pair" | "category";

export function dimensionRole(
  paramName: string | undefined,
  channels: Record<string, string>,
  chartType: string | undefined,
): DimensionRole {
  if (paramName === DETAIL_PARAM) return "detail";
  const channel = channels[paramName ?? ""] ?? "";
  if (PAIR_CHANNELS[chartType ?? ""]?.has(channel)) return "pair";
  return SERIES_CHANNELS.has(channel) ? "series" : "category";
}

/** Distinct values that suit each role, as an inclusive [min, max]. */
const ROLE_VALUE_RANGE: Record<
  Exclude<DimensionRole, "detail">,
  [number, number]
> = {
  series: [2, 12],
  pair: [2, 2],
  category: [3, 50],
};

export interface DefaultArgsOptions {
  /** The chart channel each param is drawn on, keyed by param name (see paramChannels). */
  channels?: Record<string, string>;
  /** The component's chart type, for channels whose needs depend on it. */
  chartType?: string;
  /** Distinct values per dimension, capped at VALUE_COUNT_PROBE_LIMIT. */
  valueCounts?: Map<string, number>;
}

/**
 * Computes best-guess default bindings for a component's declared params against a
 * metrics view, mirroring the smart defaults used when adding built-in widgets:
 * metrics_view params get the given metrics view, time_dimension params its time
 * dimension, and scalar params their declared default.
 *
 * Measure and dimension params get the field that suits the channel they are drawn on.
 * A series channel (color, group, a facet) needs a handful of values, or the chart drowns in
 * series and its row limit fills up on the first category; a category axis wants a few dozen;
 * the unencoded detail dimension of a row-level chart wants as many as possible, one row per
 * observation. Without value counts, fields are taken in declaration order. A measure read
 * against the primary measure (a range end, a target, OHLC) prefers one in the same format, and
 * reuses a measure rather than leaving a required param unbound when the metrics view runs out.
 */
export function buildDefaultArgs(
  params: V1ComponentParam[],
  metricsViewName: string | undefined,
  metricsViewSpec: V1MetricsViewSpec | undefined,
  options: DefaultArgsOptions = {},
): Record<string, unknown> {
  const { channels = {}, chartType, valueCounts } = options;
  const args: Record<string, unknown> = {};
  const usedFields = new Set<string>();
  const timeDimension = metricsViewSpec?.timeDimension;
  const dimensions = (metricsViewSpec?.dimensions ?? []).filter(
    (candidate) =>
      candidate.name &&
      candidate.name !== timeDimension &&
      candidate.type !== MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
  );
  const measures = (metricsViewSpec?.measures ?? []).filter(
    (candidate) => candidate.name,
  );

  const splitsIntoSeries = Object.values(channels).some((channel) =>
    SERIES_CHANNELS.has(channel),
  );

  const pickDimension = (param: V1ComponentParam) => {
    const candidates = dimensions.filter(
      (candidate) => !usedFields.has(candidate.name!),
    );
    const countOf = (name: string) => valueCounts?.get(name) ?? 0;
    if (!valueCounts) return candidates[0]?.name;

    const role = dimensionRole(param.name, channels, chartType);
    if (role === "detail") {
      return candidates.reduce<string | undefined>(
        (best, candidate) =>
          !best || countOf(candidate.name!) > countOf(best)
            ? candidate.name
            : best,
        undefined,
      );
    }

    // An axis whose categories are each split into series has room for fewer of them.
    const [min, max] =
      role === "category" && splitsIntoSeries
        ? [3, 12]
        : ROLE_VALUE_RANGE[role];
    const fit = candidates.find(
      (candidate) =>
        countOf(candidate.name!) >= min && countOf(candidate.name!) <= max,
    );
    if (fit) return fit.name;
    // Nothing in range: take the nearest field that still varies.
    const distance = (count: number) =>
      count < min ? min - count : count - max;
    return candidates
      .filter((candidate) => countOf(candidate.name!) >= 2)
      .sort((a, b) => distance(countOf(a.name!)) - distance(countOf(b.name!)))
      .concat(candidates)[0]?.name;
  };

  const formatOf = (name: unknown) => {
    const measure = measures.find((candidate) => candidate.name === name);
    return `${measure?.formatPreset ?? ""}|${measure?.formatD3 ?? ""}`;
  };
  let primaryMeasure: string | undefined;

  const pickMeasure = (param: V1ComponentParam) => {
    const unused = measures.filter(
      (candidate) => !usedFields.has(candidate.name!),
    );
    const paired =
      !!primaryMeasure &&
      PAIRED_MEASURE_CHANNELS.has(channels[param.name ?? ""] ?? "");
    const choice =
      (paired
        ? unused.find(
            (candidate) =>
              formatOf(candidate.name) === formatOf(primaryMeasure),
          )
        : undefined) ?? unused[0];
    const name = choice?.name ?? primaryMeasure ?? measures[0]?.name;
    if (!paired && !primaryMeasure) primaryMeasure = name;
    return name;
  };

  // The most constrained dimensions pick first, so an axis does not take the only field with few
  // enough values to split series by.
  const pickOrder: Record<DimensionRole, number> = {
    pair: 0,
    series: 1,
    category: 2,
    detail: 3,
  };
  const priority = (param: V1ComponentParam) =>
    param.type === "dimension"
      ? pickOrder[dimensionRole(param.name, channels, chartType)]
      : pickOrder.category;
  const ordered = [...params].sort((a, b) => priority(a) - priority(b));

  const fieldNames = new Set([
    ...dimensions.map((dimension) => dimension.name),
    ...measures.map((measure) => measure.name),
    timeDimension,
  ]);

  for (const param of ordered) {
    if (!param.name) continue;
    // A field param's default names a field of whichever metrics view it was written against; it
    // only applies when this metrics view has that field.
    const usableDefault =
      param.default !== undefined &&
      param.default !== null &&
      (!isFieldParam(param) || fieldNames.has(param.default as string));
    if (usableDefault) {
      args[param.name] = param.default;
      continue;
    }
    switch (param.type as ComponentParamType) {
      case "metrics_view":
        if (metricsViewName) args[param.name] = metricsViewName;
        break;
      case "measure": {
        const name = pickMeasure(param);
        if (name) {
          args[param.name] = name;
          usedFields.add(name);
        }
        break;
      }
      case "dimension": {
        const name = pickDimension(param);
        if (name) {
          args[param.name] = name;
          usedFields.add(name);
        }
        break;
      }
      case "time_dimension":
        if (timeDimension) args[param.name] = timeDimension;
        break;
      default:
        // Scalar params without a default stay unset.
        break;
    }
  }

  // The by-convention order_by param defaults to the bound measure (falling
  // back to any bound field) so generated Top-N queries sort sensibly.
  const orderBy = params.find(isOrderByParam);
  if (orderBy?.name && args[orderBy.name] === undefined) {
    const value = defaultOrderByValue(params, args);
    if (value) args[orderBy.name] = value;
  }

  return args;
}

/**
 * Keeps the by-convention order_by binding in step when another param is
 * re-bound: order_by follows the changed field when it pointed at its previous
 * value, and otherwise is re-seeded from the measure when it no longer matches
 * any bound field. Mutates and returns args.
 */
export function reconcileOrderByArg(
  params: V1ComponentParam[],
  args: Record<string, unknown>,
  changed?: { name: string; previousValue: unknown },
): Record<string, unknown> {
  const orderBy = params.find(isOrderByParam);
  if (!orderBy?.name) return args;
  // The user picked order_by explicitly; nothing to reconcile.
  if (changed?.name === orderBy.name) return args;
  const current = args[orderBy.name];

  // Follow a re-bound field param whose previous value order_by pointed at.
  if (changed) {
    const changedParam = params.find((param) => param.name === changed.name);
    if (
      changedParam &&
      isFieldParam(changedParam) &&
      current !== undefined &&
      current === changed.previousValue &&
      typeof args[changed.name] === "string"
    ) {
      args[orderBy.name] = args[changed.name];
      return args;
    }
  }

  // Otherwise make sure it still matches a bound field; re-seed from the measure.
  const valid = orderByOptions(params, args).map((option) => option.value);
  if (typeof current !== "string" || !valid.includes(current)) {
    const reseeded = defaultOrderByValue(params, args);
    if (reseeded) args[orderBy.name] = reseeded;
    else delete args[orderBy.name];
  }
  return args;
}

/**
 * The default value for the order_by param given the currently bound field
 * params: the measure's value, falling back to any other bound field.
 */
export function defaultOrderByValue(
  params: V1ComponentParam[],
  args: Record<string, unknown>,
): string | undefined {
  const boundFieldOf = (type: ComponentParamType) =>
    params.find(
      (param) =>
        param.type === type &&
        param.name &&
        typeof args[param.name] === "string",
    );
  const source =
    boundFieldOf("measure") ??
    boundFieldOf("time_dimension") ??
    boundFieldOf("dimension");
  return source?.name ? (args[source.name] as string) : undefined;
}

const TEMPLATE_ARG_REGEX = /\{\{\s*\.(?:params|args)\.(\w+)\s*\}\}/g;

/**
 * Client-side substitution of `{{ .params.x }}` (and its `.args.x` alias) placeholders.
 * Only used for optimistic rendering before the server-resolved properties arrive;
 * ResolveComponent is the authoritative resolution path.
 */
export function substituteArgs(
  template: string,
  args: Record<string, unknown>,
): string {
  return template.replace(TEMPLATE_ARG_REGEX, (match, name: string) => {
    const value = args[name];
    return value === undefined || value === null ? match : String(value);
  });
}

/**
 * Applies substituteArgs to every string nested in the given value.
 * Used to optimistically resolve a component's renderer properties client-side.
 */
export function substituteArgsRecursively(
  value: unknown,
  args: Record<string, unknown>,
): unknown {
  if (typeof value === "string") return substituteArgs(value, args);
  if (Array.isArray(value)) {
    return value.map((entry) => substituteArgsRecursively(entry, args));
  }
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value).map(([k, v]) => [
        k,
        substituteArgsRecursively(v, args),
      ]),
    );
  }
  return value;
}

/**
 * Optimistically resolves a component's renderer properties client-side, so the chart can render
 * before ResolveComponent returns (and at all for a draft that has yet to reconcile).
 *
 * Returns undefined when the result still holds a placeholder. Only `.params` (and its `.args`
 * alias) can be substituted here; `.fields`, `.env` and `.user` resolve on the server against the
 * metrics view, the project variables and the caller's claims. Half-resolved properties are worse
 * than none: a leftover placeholder reaches the renderer as a field name, or inside a Vega
 * expression that then fails to parse. So the fallback is all-or-nothing, and the caller waits for
 * the server resolution instead.
 */
export function optimisticRendererProps(
  rendererProperties: unknown,
  args: Record<string, unknown>,
): Record<string, unknown> | undefined {
  if (!rendererProperties) return undefined;
  const props = substituteArgsRecursively(rendererProperties, args);
  if (hasTemplatePlaceholder(props)) return undefined;
  return props as Record<string, unknown>;
}

/** Whether any string nested in the value still contains a template placeholder. */
function hasTemplatePlaceholder(value: unknown): boolean {
  if (typeof value === "string") return value.includes("{{");
  if (Array.isArray(value)) return value.some(hasTemplatePlaceholder);
  if (value && typeof value === "object") {
    return Object.values(value).some(hasTemplatePlaceholder);
  }
  return false;
}
