import { parseMeasureExpression } from "@rilldata/web-common/features/dashboards/ephemeral-measures/expression-parser";
import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";

/**
 * Precision of a measure on a dashboard surface.
 * - "approx": the measure (or a measure it references) has an exact variant, and the approximate default is shown.
 * - "exact": the viewer selected the measure, so its exact variant is computed.
 * - "inherited": the measure is computed exactly because of another exact measure: a measure it is calculated from
 *   is exact, or a measure calculated from it is exact (see `resolveExactMeasures`).
 * - "partial": only on a chip that controls several measures at once (a canvas chart): some of them are exact.
 */
export type PrecisionState = "approx" | "exact" | "inherited" | "partial";

/**
 * Props of the precision chip rendered next to a measure name. Surfaces that
 * render headers from column definitions (dimension table, pivot) carry them
 * on the column.
 */
export interface PrecisionChipProps {
  state: PrecisionState;
  /** The measure's label, or the chart's title for a chip that controls a whole chart; used in the aria label. */
  measureLabel: string;
  /** The surface's query for this measure is running an exact computation. */
  loading?: boolean;
  /** The surface's exact query failed with this message. */
  error?: string;
  onToggle?: () => void;
  /**
   * Called when the pointer or focus enters and leaves the chip. Hosts with a
   * tooltip of their own hide it while the chip is hovered, so that only the
   * chip's tooltip shows.
   */
  onHoverChange?: (hovered: boolean) => void;
}

/**
 * How the measures of a metrics view reference each other, and which of them
 * define an exact variant. Built from the metrics view spec and the ephemeral
 * measures.
 */
export interface MeasureGraph {
  /** Measures referenced by a measure: a derived measure's `requires`, or an ephemeral expression's refs. */
  refs: (name: string) => readonly string[];
  /** Whether the measure defines its own `exact_expression`. */
  hasOwnExact: (name: string) => boolean;
}

/**
 * The precision of every measure of one surface, resolved once from the
 * measures selected in that surface so that its chips and requests agree.
 * Shared by the explore and canvas dashboards; see `ExactSelection`.
 */
export interface ExactMeasureContext extends MeasureGraph {
  /** Measures selected as exact by the viewer in this surface. */
  selected: ReadonlySet<string>;
  /** Measures computed with their `exact_expression` on the dashboard; see `resolveExactMeasures`. */
  exact: ReadonlySet<string>;
}

const EMPTY_SET: ReadonlySet<string> = new Set();
const NO_REFS: readonly string[] = [];

export const EMPTY_EXACT_CONTEXT: ExactMeasureContext = {
  selected: EMPTY_SET,
  exact: EMPTY_SET,
  refs: () => NO_REFS,
  hasOwnExact: () => false,
};

/**
 * The exact selection of a dashboard: the measures the viewer switched to
 * exact, per surface, tied to the query context (filters, time controls,
 * view) they were selected in. A surface is one query target with its own
 * chips: the big-number row, one leaderboard, the dimension table, a pivot, a
 * chart, a KPI grid. Surfaces are independent, so a click only re-runs the
 * queries of the surface it is in.
 */
export interface ExactSelection {
  contextKey: string;
  surfaces: Record<string, string[]>;
}

/**
 * The measures selected as exact in a surface, empty when the selection
 * belongs to another query context.
 */
export function selectedInSurface(
  selection: ExactSelection | undefined,
  contextKey: string,
  surface: string,
): ReadonlySet<string> {
  if (!selection || selection.contextKey !== contextKey) return EMPTY_SET;
  const names = selection.surfaces[surface];
  return names?.length ? new Set(names) : EMPTY_SET;
}

/**
 * The selection after switching a measure of a surface between approximate and
 * exact, for the given query context. A selection made in another context is
 * discarded. Returns undefined when nothing is selected any more.
 */
export function toggleInSelection(
  selection: ExactSelection | undefined,
  contextKey: string,
  surface: string,
  measureName: string,
): ExactSelection | undefined {
  const surfaces =
    selection?.contextKey === contextKey ? { ...selection.surfaces } : {};
  const current = surfaces[surface] ?? [];
  const names = current.includes(measureName)
    ? current.filter((name) => name !== measureName)
    : [...current, measureName];
  if (names.length) surfaces[surface] = names;
  else delete surfaces[surface];
  return Object.keys(surfaces).length ? { contextKey, surfaces } : undefined;
}

/**
 * The selection after switching a whole surface between approximate and
 * exact, for surfaces with one chip for all their measures (canvas charts):
 * selects every given measure, or clears the surface when all of them are
 * selected already. A selection made in another context is discarded. Returns
 * undefined when nothing is selected any more.
 */
export function toggleSurfaceInSelection(
  selection: ExactSelection | undefined,
  contextKey: string,
  surface: string,
  measureNames: readonly string[],
): ExactSelection | undefined {
  const surfaces =
    selection?.contextKey === contextKey ? { ...selection.surfaces } : {};
  const current = new Set(surfaces[surface] ?? []);
  // `every` holds for an empty list, which clears the surface as well.
  if (measureNames.every((name) => current.has(name))) delete surfaces[surface];
  else surfaces[surface] = [...measureNames];
  return Object.keys(surfaces).length ? { contextKey, surfaces } : undefined;
}

export function createExactMeasureContext(
  selected: ReadonlySet<string>,
  specMeasures: MetricsViewSpecMeasure[] | undefined,
  ephemeralMeasures?: EphemeralMeasureDef[],
): ExactMeasureContext {
  const names: string[] = [];
  const refs = new Map<string, readonly string[]>();
  const hasOwnExact = new Set<string>();
  for (const measure of specMeasures ?? []) {
    if (!measure.name) continue;
    names.push(measure.name);
    if (measure.referencedMeasures?.length) {
      refs.set(measure.name, measure.referencedMeasures);
    }
    if (measure.exactExpression) hasOwnExact.add(measure.name);
  }
  for (const def of ephemeralMeasures ?? []) {
    names.push(def.name);
    const parsed = parseMeasureExpression(def.expression);
    if (parsed.refs.length) refs.set(def.name, parsed.refs);
  }
  const graph: MeasureGraph = {
    refs: (name) => refs.get(name) ?? NO_REFS,
    hasOwnExact: (name) => hasOwnExact.has(name),
  };
  return {
    ...graph,
    selected,
    exact: resolveExactMeasures(selected, names, graph),
  };
}

/**
 * The measures a surface computes with their `exact_expression`, given the
 * measures selected in it.
 *
 * This mirrors the runtime, which resolves the exact set of a query once: the
 * closure of the measures flagged in the query, following referenced measures,
 * filtered to the measures that define an `exact_expression`. Selecting a
 * ratio therefore makes the unique counts it is calculated from exact.
 *
 * The surface flags a requested measure when its closure meets this set
 * (`requestExactNames`), so the set is also closed under that rule: if a
 * flagged derived measure would make another measure exact in the queries
 * that contain it, that measure is part of the set and is exact in every
 * query of the surface. Surfaces are independent of each other: the same
 * measure can be exact in one leaderboard and approximate in the next, so a
 * click only re-runs the queries of the surface it is in.
 */
export function resolveExactMeasures(
  selected: ReadonlySet<string>,
  measureNames: Iterable<string>,
  graph: MeasureGraph,
): ReadonlySet<string> {
  if (!selected.size) return EMPTY_SET;
  let exact = ownExactIn(closureOf(selected, graph), graph);
  if (!exact.size) return EMPTY_SET;
  for (;;) {
    const flagged = new Set(selected);
    for (const name of measureNames) {
      if (intersects(transitiveRefs(name, graph), exact)) flagged.add(name);
    }
    // The flagged set only grows between rounds, so the exact set does too;
    // an unchanged size means the fixpoint is reached.
    const next = ownExactIn(closureOf(flagged, graph), graph);
    if (next.size === exact.size) return exact;
    exact = next;
  }
}

/**
 * The measure itself plus every measure it references, transitively.
 * Guards against reference cycles.
 */
export function transitiveRefs(name: string, graph: MeasureGraph): Set<string> {
  const result = new Set<string>();
  const stack = [name];
  while (stack.length) {
    const current = stack.pop()!;
    if (result.has(current)) continue;
    result.add(current);
    for (const ref of graph.refs(current)) stack.push(ref);
  }
  return result;
}

function closureOf(names: Iterable<string>, graph: MeasureGraph): Set<string> {
  const result = new Set<string>();
  for (const name of names) {
    for (const ref of transitiveRefs(name, graph)) result.add(ref);
  }
  return result;
}

function ownExactIn(names: Iterable<string>, graph: MeasureGraph): Set<string> {
  const result = new Set<string>();
  for (const name of names) {
    if (graph.hasOwnExact(name)) result.add(name);
  }
  return result;
}

function intersects(
  names: Iterable<string>,
  set: ReadonlySet<string>,
): boolean {
  for (const name of names) {
    if (set.has(name)) return true;
  }
  return false;
}

/**
 * Whether the measure can be computed exactly: it, or a measure it
 * references, defines an `exact_expression`.
 */
export function hasExactVariant(name: string, graph: MeasureGraph): boolean {
  for (const ref of transitiveRefs(name, graph)) {
    if (graph.hasOwnExact(ref)) return true;
  }
  return false;
}

/** The given measures that have an exact variant; see `hasExactVariant`. */
export function withExactVariant(
  names: Iterable<string>,
  graph: MeasureGraph,
): string[] {
  const result: string[] = [];
  for (const name of names) {
    if (hasExactVariant(name, graph)) result.push(name);
  }
  return result;
}

/**
 * The precision to show for a measure, or null when the measure has no exact
 * variant and should render without a precision chip.
 */
export function precisionStateFor(
  name: string,
  ctx: ExactMeasureContext,
): PrecisionState | null {
  if (!hasExactVariant(name, ctx)) return null;
  if (ctx.selected.has(name)) return "exact";
  if (intersects(transitiveRefs(name, ctx), ctx.exact)) return "inherited";
  return "approx";
}

/**
 * The requested measures to flag `exact` in a request: every requested measure
 * that is, or is calculated from, an exact measure. The runtime resolves the
 * same exact set from these flags (see `resolveExactMeasures`), so a measure
 * has the same value in every request of the surface.
 */
export function requestExactNames(
  requested: Iterable<string>,
  ctx: ExactMeasureContext,
): Set<string> {
  const result = new Set<string>();
  if (!ctx.exact.size) return result;
  for (const name of requested) {
    if (intersects(transitiveRefs(name, ctx), ctx.exact)) result.add(name);
  }
  return result;
}
