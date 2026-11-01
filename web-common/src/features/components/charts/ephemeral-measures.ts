import { ephemeralSpecsToDefs } from "@rilldata/web-common/features/dashboards/ephemeral-measures/canvas";
import {
  applyExactMeasures,
  ephemeralMeasureToSpecMeasure,
  mapEphemeralMeasuresForRequest,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import type {
  MetricsViewSpecMeasure,
  V1MetricsViewAggregationMeasure,
} from "@rilldata/web-common/runtime-client";
import type { CommonChartProperties } from "./types";

type EphemeralMeasureCarrier = Pick<CommonChartProperties, "adhoc_measures">;
type MeasureRequestCarrier = Pick<
  CommonChartProperties,
  "adhoc_measures" | "exact_measures"
>;

/**
 * Attaches the `expression` compute to request measures whose name matches one
 * of the chart's ephemeral measures, and flags the chart's exact measures with
 * `exact: true`. Leaves other measures untouched.
 */
export function withEphemeralMeasures(
  config: MeasureRequestCarrier,
  measures: V1MetricsViewAggregationMeasure[],
): V1MetricsViewAggregationMeasure[] {
  return applyExactMeasures(
    mapEphemeralMeasuresForRequest(
      measures,
      ephemeralSpecsToDefs(config.adhoc_measures),
    ),
    config.exact_measures?.length ? new Set(config.exact_measures) : undefined,
  );
}

/**
 * Names of the chart's ephemeral measures, for quick membership checks.
 */
export function chartEphemeralMeasureNames(
  config: EphemeralMeasureCarrier,
): Set<string> {
  return new Set(
    ephemeralSpecsToDefs(config.adhoc_measures)?.map((def) => def.name) ?? [],
  );
}

/**
 * Synthesizes a metrics view spec measure for an ephemeral measure name so
 * labels, formatters and tooltips resolve like for any other measure. Returns
 * undefined when the name is not one of the chart's ephemeral measures.
 */
export function resolveEphemeralMeasureSpec(
  config: EphemeralMeasureCarrier,
  name: string,
): MetricsViewSpecMeasure | undefined {
  const def = ephemeralSpecsToDefs(config.adhoc_measures)?.find(
    (d) => d.name === name,
  );
  return def ? ephemeralMeasureToSpecMeasure(def) : undefined;
}
