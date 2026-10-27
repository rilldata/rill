import {
  createExactMeasureContext,
  EMPTY_SET,
  selectedInSurface,
  type ExactMeasureContext,
} from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
import { getQueryContextKey } from "@rilldata/web-common/features/dashboards/exact-measures/query-context-key";
import type { DashboardDataSources } from "./types";

/**
 * Measures the viewer selected as exact in a surface (see
 * `exploreExactSurface`) for the current query context. Empty when the stored
 * selection belongs to a different context (the store prunes such selections,
 * this is a second guard for derived readers).
 */
export const selectedExactMeasureNames = (
  { dashboard }: Pick<DashboardDataSources, "dashboard">,
  surface: string,
): ReadonlySet<string> => {
  // Skip serializing the query context in the common case of no selection;
  // every surface calls this on every dashboard update.
  if (!dashboard.exactMeasures) return EMPTY_SET;
  return selectedInSurface(
    dashboard.exactMeasures,
    getQueryContextKey(dashboard),
    surface,
  );
};

/**
 * The context that decides the precision of every measure of a surface and
 * flags the surface's request measures; see `exact-measures.ts`.
 */
export const exactMeasureContextFor = (
  {
    dashboard,
    validMetricsView,
  }: Pick<DashboardDataSources, "dashboard" | "validMetricsView">,
  surface: string,
): ExactMeasureContext =>
  createExactMeasureContext(
    selectedExactMeasureNames({ dashboard }, surface),
    validMetricsView?.measures,
  );

/**
 * Store form of `exactMeasureContextFor`: a function of the surface, with the
 * contexts cached for the current dashboard state.
 */
export const exactMeasureContext = (
  args: Pick<DashboardDataSources, "dashboard" | "validMetricsView">,
): ((surface: string) => ExactMeasureContext) => {
  const contexts = new Map<string, ExactMeasureContext>();
  return (surface) => {
    let ctx = contexts.get(surface);
    if (!ctx) {
      ctx = exactMeasureContextFor(args, surface);
      contexts.set(surface, ctx);
    }
    return ctx;
  };
};

/**
 * Whether any measure of the metrics view defines an `exact_expression`.
 * Surfaces skip all precision work when this is false.
 */
export const hasAnyExactVariant = ({
  validMetricsView,
}: Pick<DashboardDataSources, "validMetricsView">): boolean =>
  validMetricsView?.measures?.some((m) => !!m.exactExpression) ?? false;

// `selectedExactMeasureNames` and `exactMeasureContextFor` take a surface, so
// they are plain functions; the store-backed selectors take the data sources only.
export const exactMeasureSelectors = {
  exactMeasureContext,
  hasAnyExactVariant,
};
