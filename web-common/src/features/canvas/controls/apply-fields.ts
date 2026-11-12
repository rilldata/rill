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

function swapFields(value: unknown, swaps: Map<string, string>): unknown {
  // All swaps are applied in one pass over the original spec, so A→B and B→A don't chain.
  const walk = (v: unknown, key: string, inFieldList: boolean): unknown => {
    if (typeof v === "string") {
      return (inFieldList || SWAP_SCALAR_KEYS.has(key)) && swaps.has(v)
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
  let used = false;
  const swaps = new Map([[field, field]]);
  swaps.get = (key: string) => {
    if (key === field) used = true;
    return field;
  };
  swapFields(spec, swaps);
  return used;
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

export function applyFieldControls<T>(spec: T, controls: AppliedControl[]): T {
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
    if (control.target && to && to !== control.target) {
      swaps.set(control.target, to);
    }
  }

  let result = spec as Record<string, unknown>;
  if (swaps.size > 0) {
    result = swapFields(result, swaps) as Record<string, unknown>;
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
