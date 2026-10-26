import {
  createExactMeasureContext,
  precisionStateFor,
  withExactVariant,
  type ExactMeasureContext,
  type PrecisionChipProps,
  type PrecisionState,
} from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import { derived, type Readable } from "svelte/store";

/**
 * The exact-selection surface of a canvas component: the component itself, or
 * one leaderboard of a leaderboard component. Every surface is toggled on its
 * own, so a click only re-runs that component's (or that leaderboard's)
 * queries.
 */
export function canvasExactSurface(
  componentId: string,
  dimensionName?: string,
): string {
  return dimensionName ? `${componentId}:${dimensionName}` : componentId;
}

/**
 * The exact-measure context of a canvas surface: its selection, resolved over
 * the metrics view's measures.
 */
export function canvasExactMeasureContext(
  canvasEntity: CanvasEntity,
  surface: string,
  metricsViewName: string,
): Readable<ExactMeasureContext> {
  return derived(
    [
      canvasEntity.selectedExactMeasures(surface),
      canvasEntity.metricsView.getMeasuresForMetricView(metricsViewName),
    ],
    ([selected, specMeasures]) =>
      createExactMeasureContext(selected, specMeasures),
  );
}

/**
 * The measures of a canvas surface to request with `exact: true`: the ones in
 * its selection. The selection only ever holds measures with an exact variant
 * (see `CanvasEntity.toggleExactSurface`), and the runtime ignores the flag on
 * any other measure, so the metrics view spec is not consulted here.
 */
export function canvasExactRequestNames(
  measureNames: string[],
  selected: ReadonlySet<string>,
): string[] {
  if (!selected.size) return [];
  return measureNames.filter((name) => selected.has(name));
}

/**
 * The measure names a chart spec uses: the `field` (or `fields`) of every
 * input param the chart declares as a measure field.
 */
export function chartMeasureNames(
  chartInputParams: Record<string, ComponentInputParam> | undefined,
  spec: object,
): string[] {
  const names = new Set<string>();
  for (const [key, param] of Object.entries(chartInputParams ?? {})) {
    if (param.meta?.chartFieldInput?.type !== "measure") continue;
    const value = (spec as Record<string, unknown>)[key] as
      | { field?: string; fields?: string[] }
      | undefined;
    if (!value || typeof value !== "object") continue;
    if (value.field) names.add(value.field);
    for (const field of value.fields ?? []) names.add(field);
  }
  return [...names];
}

/**
 * Builds the precision chip of a measure in a canvas component (a KPI, a table
 * or pivot column), or undefined when the measure has no exact variant.
 */
export function canvasPrecisionChip(
  measureName: string,
  measureLabel: string,
  ctx: ExactMeasureContext,
  loading: boolean,
  error: string | undefined,
  onToggle: () => void,
): PrecisionChipProps | undefined {
  const state = precisionStateFor(measureName, ctx);
  if (!state) return undefined;
  return chipProps(state, measureLabel, loading, error, onToggle);
}

/**
 * Builds the single precision chip of a canvas chart, which switches every
 * measure of the chart with an exact variant at once (see
 * `CanvasEntity.toggleExactSurface`), or undefined when no measure has one.
 * The chip is exact when all of those measures are exact and partial when only
 * some are, which happens when the chart's measures change after the
 * selection was made.
 */
export function canvasChartPrecisionChip(
  measureNames: string[],
  chartTitle: string,
  ctx: ExactMeasureContext,
  loading: boolean,
  error: string | undefined,
  onToggle: () => void,
): PrecisionChipProps | undefined {
  const capable = withExactVariant(measureNames, ctx);
  if (!capable.length) return undefined;
  const exactCount = capable.filter(
    (name) => precisionStateFor(name, ctx) !== "approx",
  ).length;
  const state: PrecisionState =
    exactCount === capable.length ? "exact" : exactCount ? "partial" : "approx";
  return chipProps(state, chartTitle, loading, error, onToggle);
}

// Loading and errors only concern exact computations: an approximate surface
// that is still fetching shows a plain chip.
function chipProps(
  state: PrecisionState,
  measureLabel: string,
  loading: boolean,
  error: string | undefined,
  onToggle: () => void,
): PrecisionChipProps {
  return {
    state,
    measureLabel,
    loading: state !== "approx" && loading,
    error: state !== "approx" ? error : undefined,
    onToggle,
  };
}
