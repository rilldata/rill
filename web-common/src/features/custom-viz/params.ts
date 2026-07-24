import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
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
export const CUSTOM_VIZ_FOLDER = "viz_library";

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
    if (!param.name || !FIELD_PARAM_TYPES.has(param.type as ComponentParamType))
      continue;
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
): ComponentInputParam {
  const common = {
    label: prettyParamLabel(param.name),
    optional: !param.required,
    description: param.description,
  };
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
              value: option,
              label: String(option),
            })),
            default: param.default,
          },
        };
      }
      return { type: param.type === "number" ? "number" : "text", ...common };
    default:
      return { type: "text", ...common };
  }
}

/**
 * Computes best-guess default bindings for a component's declared params against a
 * metrics view, mirroring the smart defaults used when adding built-in widgets:
 * metrics_view params get the given metrics view, time_dimension params its time
 * dimension, measure/dimension params the first unused field, and scalar params
 * their declared default.
 */
export function buildDefaultArgs(
  params: V1ComponentParam[],
  metricsViewName: string | undefined,
  metricsViewSpec: V1MetricsViewSpec | undefined,
): Record<string, unknown> {
  const args: Record<string, unknown> = {};
  const usedFields = new Set<string>();
  const timeDimension = metricsViewSpec?.timeDimension;
  const dimensions = metricsViewSpec?.dimensions ?? [];
  const measures = metricsViewSpec?.measures ?? [];

  for (const param of params) {
    if (!param.name) continue;
    if (param.default !== undefined && param.default !== null) {
      args[param.name] = param.default;
      continue;
    }
    switch (param.type as ComponentParamType) {
      case "metrics_view":
        if (metricsViewName) args[param.name] = metricsViewName;
        break;
      case "measure": {
        const measure = measures.find(
          (candidate) => candidate.name && !usedFields.has(candidate.name),
        );
        if (measure?.name) {
          args[param.name] = measure.name;
          usedFields.add(measure.name);
        }
        break;
      }
      case "dimension": {
        const dimension = dimensions.find(
          (candidate) =>
            candidate.name &&
            !usedFields.has(candidate.name) &&
            candidate.name !== timeDimension &&
            candidate.type !== MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
        );
        if (dimension?.name) {
          args[param.name] = dimension.name;
          usedFields.add(dimension.name);
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
      FIELD_PARAM_TYPES.has(changedParam.type as ComponentParamType) &&
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
