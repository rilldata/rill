import { ephemeralSpecsToDefs } from "@rilldata/web-common/features/dashboards/ephemeral-measures/canvas";
import {
  ephemeralMeasureToSpecMeasure,
  mapEphemeralMeasuresForRequest,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import type {
  MetricsViewSpecMeasure,
  V1MetricsViewAggregationMeasure,
} from "@rilldata/web-common/runtime-client";
import type { CommonChartProperties } from "./types";

type EphemeralMeasureCarrier = Pick<
  CommonChartProperties,
  "ephemeral_measures"
>;

/**
 * Attaches the `expression` compute to request measures whose name matches one
 * of the chart's ephemeral measures. Leaves other measures untouched.
 */
export function withEphemeralMeasures(
  config: EphemeralMeasureCarrier,
  measures: V1MetricsViewAggregationMeasure[],
): V1MetricsViewAggregationMeasure[] {
  return mapEphemeralMeasuresForRequest(
    measures,
    ephemeralSpecsToDefs(config.ephemeral_measures),
  );
}

/**
 * Names of the chart's ephemeral measures, for quick membership checks.
 */
export function chartEphemeralMeasureNames(
  config: EphemeralMeasureCarrier,
): Set<string> {
  return new Set(
    ephemeralSpecsToDefs(config.ephemeral_measures)?.map((def) => def.name) ??
      [],
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
  const def = ephemeralSpecsToDefs(config.ephemeral_measures)?.find(
    (d) => d.name === name,
  );
  return def ? ephemeralMeasureToSpecMeasure(def) : undefined;
}
