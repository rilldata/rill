import { chartEphemeralMeasureNames } from "@rilldata/web-common/features/components/charts/ephemeral-measures";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { ControlKind, type AppliedControl } from "./controls";

/**
 * Applies field controls (measure and dimension swaps, slot controls) to a component's spec.
 * The result is the spec viewers see; the YAML spec is never modified.
 *
 * The runtime validates swaps with the same rules in runtime/canvas/controls.go (SwapField);
 * keep them in sync.
 */

// Renderer property keys whose string value is a field name.
const SWAP_SCALAR_KEYS = new Set([
  "field",
  "name",
  "measure",
  "dimension",
  "sort_by",
]);

// Renderer property keys whose value is a list of field names or field objects.
const SWAP_LIST_KEYS = new Set([
  "fields",
  "measures",
  "dimensions",
  "columns",
  "row_dimensions",
  "col_dimensions",
]);

function isFieldPosition(key: string, inFieldList: boolean): boolean {
  return inFieldList || SWAP_SCALAR_KEYS.has(key);
}

function swapFields(value: unknown, swaps: Map<string, string>): unknown {
  // All swaps are applied in one pass over the original spec, so A→B and B→A don't chain.
  const walk = (v: unknown, key: string, inFieldList: boolean): unknown => {
    if (typeof v === "string") {
      return isFieldPosition(key, inFieldList) && swaps.has(v)
        ? swaps.get(v)
        : v;
    }
    if (Array.isArray(v)) {
      const isFieldList = SWAP_LIST_KEYS.has(key);
      return v.map((child) => walk(child, key, isFieldList));
    }
    if (v && typeof v === "object") {
      return Object.fromEntries(
        Object.entries(v).map(([k, child]) => [k, walk(child, k, false)]),
      );
    }
    return v;
  };
  return walk(value, "", false);
}

/** Whether a component's spec shows the field somewhere a measure or dimension swap would change. */
export function specUsesField(spec: unknown, field: string): boolean {
  const visit = (
    value: unknown,
    key: string,
    inFieldList: boolean,
  ): boolean => {
    if (typeof value === "string")
      return value === field && isFieldPosition(key, inFieldList);
    if (Array.isArray(value)) {
      return value.some((child) => visit(child, key, SWAP_LIST_KEYS.has(key)));
    }
    if (value && typeof value === "object") {
      return Object.keys(value).some((key) =>
        visit((value as Record<string, unknown>)[key], key, false),
      );
    }
    return false;
  };
  return visit(spec, "", false);
}

/**
 * Whether a component's spec shows a field twice: in one list of fields (a table column or a KPI,
 * say), on both axes of a chart (x and y, or y1 and y2), or in both the rows and the columns of
 * a pivot. The runtime implements the same rules (showsFieldTwice in runtime/canvas/controls.go);
 * keep them in sync.
 */
export function showsFieldTwice(spec: unknown, field: string): boolean {
  const nameOf = (entry: unknown) =>
    entry && typeof entry === "object" && !Array.isArray(entry)
      ? (entry as Record<string, unknown>).name
      : entry;
  const visit = (value: unknown, key: string): boolean => {
    if (Array.isArray(value)) {
      const isFieldList = SWAP_LIST_KEYS.has(key);
      let count = 0;
      for (const child of value as unknown[]) {
        if (isFieldList && nameOf(child) === field) count++;
        if (visit(child, key)) return true;
      }
      return count > 1;
    }
    if (value && typeof value === "object") {
      return Object.entries(value).some(([k, child]) => visit(child, k));
    }
    return false;
  };
  if (visit(spec, "")) return true;
  if (!spec || typeof spec !== "object") return false;

  const props = spec as Record<string, unknown>;
  const fieldAt = (key: string) => {
    const value = props[key];
    return value && typeof value === "object" && !Array.isArray(value)
      ? (value as Record<string, unknown>).field
      : undefined;
  };
  for (const [a, b] of [
    ["x", "y"],
    ["y1", "y2"],
  ]) {
    if (fieldAt(a) === field && fieldAt(b) === field) return true;
  }
  const inList = (key: string) =>
    Array.isArray(props[key]) &&
    (props[key] as unknown[]).some((entry) => nameOf(entry) === field);
  return inList("row_dimensions") && inList("col_dimensions");
}

/**
 * Whether swapping `from` for `to` in a component's spec would show a field twice (see
 * showsFieldTwice). The runtime rejects such a control on the item itself; an item under a
 * control of an enclosing level keeps its own field instead (see applyFieldControls).
 */
export function swapRepeatsField(
  spec: unknown,
  from: string,
  to: string,
): boolean {
  return showsFieldTwice(swapFields(spec, new Map([[from, to]])), to);
}

function setPath(
  spec: Record<string, unknown>,
  path: string,
  value: unknown,
): Record<string, unknown> {
  const [head, ...rest] = path.split(".");
  if (rest.length === 0) {
    const next = { ...spec };
    if (value === undefined) delete next[head];
    else next[head] = value;
    return next;
  }
  const child = spec[head];
  return {
    ...spec,
    [head]: setPath(
      child && typeof child === "object" && !Array.isArray(child)
        ? (child as Record<string, unknown>)
        : {},
      rest.join("."),
      value,
    ),
  };
}

/**
 * Applies the controls to a component's spec. A swap leaves the component on its own field when
 * the component's metrics view lacks the chosen field (`hasField`, when known) or when the swaps
 * would show the chosen field twice in it, the same as the runtime accepts for a control of an
 * enclosing level (see ValidateControls).
 */
/** The fields a component's spec sets at a slot's path: the names in a list, or the one field. */
export function slotFields(spec: unknown, path: string): string[] {
  let value: unknown = spec;
  for (const key of path.split(".")) {
    value =
      value && typeof value === "object" && !Array.isArray(value)
        ? (value as Record<string, unknown>)[key]
        : undefined;
  }
  const list = Array.isArray(value) ? (value as unknown[]) : [value];
  return list
    .map((entry) =>
      entry && typeof entry === "object" && !Array.isArray(entry)
        ? (entry as Record<string, unknown>).name
        : entry,
    )
    .filter((name): name is string => typeof name === "string" && !!name);
}

export function applyFieldControls<T>(
  spec: T,
  controls: AppliedControl[],
  hasField?: (kind: "measure" | "dimension", name: string) => boolean,
): T {
  if (!spec || controls.length === 0) return spec;

  const swaps = new Map<string, string>();
  for (const { control, values } of controls) {
    if (
      control.kind !== ControlKind.Measure &&
      control.kind !== ControlKind.Dimension
    ) {
      continue;
    }
    const to = values?.[0];
    if (
      control.target &&
      typeof to === "string" &&
      to &&
      to !== control.target &&
      (!hasField || hasField(control.kind, to))
    ) {
      swaps.set(control.target, to);
    }
  }

  let result = spec as Record<string, unknown>;
  // Drop the swaps whose field the result would show twice, until none does.
  // Swaps are checked together, so exchanging two fields is fine.
  while (swaps.size > 0) {
    const swapped = swapFields(result, swaps) as Record<string, unknown>;
    const repeated = [...swaps].filter(([, to]) =>
      showsFieldTwice(swapped, to),
    );
    if (repeated.length === 0) {
      result = swapped;
      break;
    }
    for (const [from] of repeated) swaps.delete(from);
  }

  for (const { control, values } of controls) {
    if (control.kind !== ControlKind.Slot || !control.target || !values) {
      continue;
    }
    result = setPath(
      result,
      control.target,
      control.multiple ? values : values[0],
    );
  }

  return result as T;
}

/**
 * A `hasField` for applyFieldControls: whether a component's metrics view has the measure (or
 * the component defines it as an adhoc measure) or the dimension. Undefined while the metrics
 * view is unknown, so swaps apply as chosen until it loads.
 */
export function componentFields(
  metricsView: V1MetricsViewSpec | undefined,
  spec: unknown,
): ((kind: "measure" | "dimension", name: string) => boolean) | undefined {
  if (!metricsView) return undefined;
  const adhoc = chartEphemeralMeasureNames(
    (spec ?? {}) as Parameters<typeof chartEphemeralMeasureNames>[0],
  );
  const timeDimension = metricsView.timeDimension;
  return (kind, name) =>
    kind === "measure"
      ? adhoc.has(name) || !!metricsView.measures?.some((m) => m.name === name)
      : !!metricsView.dimensions?.some((d) => d.name === name) ||
        (!!timeDimension &&
          (name === timeDimension ||
            name.startsWith(`${timeDimension}_rill_`)));
}
