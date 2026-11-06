import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";

/**
 * Precision of a measure on a dashboard surface.
 * - "approx": the measure has an exact variant, and the approximate default is shown.
 * - "exact": the viewer selected the measure, so its exact variant is computed.
 * - "partial": only on a chip that controls several measures at once (a canvas chart): some of them are exact.
 */
export type PrecisionState = "approx" | "exact" | "partial";

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
 * The precision of every measure of one surface, resolved once from the
 * measures selected in that surface so that its chips and requests agree.
 * Shared by the explore and canvas dashboards; see `ExactSelection`.
 */
export interface ExactMeasureContext {
  /** Measures selected as exact by the viewer in this surface. */
  selected: ReadonlySet<string>;
  /** Measures computed with their `exact_expression` on the surface: the selected measures that define one. */
  exact: ReadonlySet<string>;
  /**
   * Whether the measure defines an `exact_expression`. Only simple measures
   * do: derived and ad hoc measures have no exact variant and no chip, and
   * are computed from the other measures as the query computes them.
   */
  hasExactVariant: (name: string) => boolean;
}

export const EMPTY_SET: ReadonlySet<string> = new Set();

export const EMPTY_EXACT_CONTEXT: ExactMeasureContext = {
  selected: EMPTY_SET,
  exact: EMPTY_SET,
  hasExactVariant: () => false,
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

/**
 * The context of a surface: its selection, filtered to the measures that
 * define an `exact_expression`. A selection can outlive the metrics view it
 * was made against (the project is edited while the dashboard is open), so a
 * selected measure that lost its exact variant is ignored.
 */
export function createExactMeasureContext(
  selected: ReadonlySet<string>,
  specMeasures: MetricsViewSpecMeasure[] | undefined,
): ExactMeasureContext {
  const withExactVariant = new Set<string>();
  for (const measure of specMeasures ?? []) {
    if (measure.name && measure.exactExpression) {
      withExactVariant.add(measure.name);
    }
  }
  const exact = new Set<string>();
  for (const name of selected) {
    if (withExactVariant.has(name)) exact.add(name);
  }
  return {
    selected,
    exact: exact.size ? exact : EMPTY_SET,
    hasExactVariant: (name) => withExactVariant.has(name),
  };
}

/** Whether the measure can be computed exactly: it defines an `exact_expression`. */
export function hasExactVariant(
  name: string,
  ctx: ExactMeasureContext,
): boolean {
  return ctx.hasExactVariant(name);
}

/** The given measures that have an exact variant; see `hasExactVariant`. */
export function withExactVariant(
  names: Iterable<string>,
  ctx: ExactMeasureContext,
): string[] {
  const result: string[] = [];
  for (const name of names) {
    if (ctx.hasExactVariant(name)) result.push(name);
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
  if (!ctx.hasExactVariant(name)) return null;
  return ctx.exact.has(name) ? "exact" : "approx";
}

/**
 * The requested measures to flag `exact` in a request: the ones the surface
 * computes exactly. `applyExactMeasures` flags the compute entries
 * (comparisons, percent of total) of these measures as well, so a measure has
 * the same value in every request of the surface.
 */
export function requestExactNames(
  requested: Iterable<string>,
  ctx: ExactMeasureContext,
): Set<string> {
  const result = new Set<string>();
  if (!ctx.exact.size) return result;
  for (const name of requested) {
    if (ctx.exact.has(name)) result.add(name);
  }
  return result;
}
