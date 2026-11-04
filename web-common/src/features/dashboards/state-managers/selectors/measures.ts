import { appendEphemeralSpecMeasures } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecMeasure,
  type V1MetricsViewSpec,
  V1TimeGrain,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationMeasureComputeExpression,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

/**
 * The subset of an aggregation measure that identifies the spec measure it derives from.
 * Both spec measures and aggregation measures satisfy it,
 * so selectors that only resolve the source measure can accept either.
 */
type AggregationMeasureRef = Pick<
  V1MetricsViewAggregationMeasure,
  | "name"
  | "comparisonDelta"
  | "comparisonValue"
  | "comparisonRatio"
  | "percentOfTotal"
> & {
  /**
   * Aggregation measures carry the ephemeral expression compute here, while spec measures carry
   * their SQL expression as a plain string. Only the object form marks an ephemeral measure.
   */
  expression?: V1MetricsViewAggregationMeasureComputeExpression | string;
};

export const allMeasures = ({
  validMetricsView,
  validExplore,
  dashboard,
}: Pick<DashboardDataSources, "validMetricsView" | "validExplore"> &
  Partial<
    Pick<DashboardDataSources, "dashboard">
  >): MetricsViewSpecMeasure[] => {
  if (!validMetricsView?.measures || !validExplore?.measures) return [];

  const specMeasures = validMetricsView.measures
    .filter((m) => validExplore.measures!.includes(m.name!))
    // Sort the filtered measures based on their order in validExplore.measures
    .sort(
      (a, b) =>
        validExplore.measures!.indexOf(a.name!) -
        validExplore.measures!.indexOf(b.name!),
    );
  // ephemeral measures behave like regular measures across the
  // explore; they get synthetic spec entries so labels/formatting resolve.
  return appendEphemeralSpecMeasures(
    specMeasures,
    dashboard?.ephemeralMeasures,
  );
};

export const visibleMeasures = ({
  validMetricsView,
  validExplore,
  dashboard,
}: DashboardDataSources): MetricsViewSpecMeasure[] => {
  if (!validMetricsView?.measures || !validExplore?.measures) return [];

  const all = allMeasures({ validMetricsView, validExplore, dashboard });
  return dashboard.visibleMeasures
    .map((mes) => all.find((m) => m.name === mes))
    .filter(Boolean) as MetricsViewSpecMeasure[];
};

export const getMeasureByName = (
  dashData: DashboardDataSources,
): ((name: string | undefined) => MetricsViewSpecMeasure | undefined) => {
  return (name: string | undefined) => {
    return allMeasures(dashData)?.find((measure) => measure.name === name);
  };
};

export const measureLabel = (
  dashData: DashboardDataSources,
): ((m: string) => string) => {
  return (measureName) => {
    const measure = allMeasures(dashData).find((d) => d.name === measureName);
    return measure?.displayName || measureName;
  };
};

export const isMeasureValidPercentOfTotal = ({
  validMetricsView,
}: DashboardDataSources): ((measureName: string) => boolean) => {
  return (measureName: string) => {
    const selectedMeasure = validMetricsView?.measures?.find(
      (m) => m.name === measureName,
    );
    return selectedMeasure?.validPercentOfTotal ?? false;
  };
};

export const filteredSimpleMeasures = ({
  validMetricsView,
  validExplore,
}: DashboardDataSources) => {
  return () =>
    getFilteredSimpleMeasures(
      validMetricsView?.measures ?? [],
      validExplore?.measures,
    );
};

export function getFilteredSimpleMeasures(
  allMeasures: MetricsViewSpecMeasure[],
  exploreMeasures: string[] | undefined,
) {
  if (!exploreMeasures) return [];

  return (
    allMeasures
      .filter((m) => exploreMeasures.includes(m.name!) && isSimpleMeasure(m))
      // Sort the filtered measures based on their order in validExplore.measures
      .sort(
        (a, b) =>
          exploreMeasures.indexOf(a.name!) - exploreMeasures.indexOf(b.name!),
      )
  );
}

/**
 * A simple measure can be aggregated by any set of dimensions, including none.
 * Window measures, time comparison measures and measures with required dimensions
 * can only be queried when specific dimensions are grouped by.
 */
export const isSimpleMeasure = (measure: MetricsViewSpecMeasure) =>
  !measure.window &&
  measure.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON &&
  !measure.requiredDimensions?.length;

/**
 * A totals query aggregates a measure to a single value with no dimensions,
 * so it cannot satisfy a measure's required dimensions. E.g. a rolling window
 * measure ordered by the time dimension produces one value per time bucket and
 * has no meaningful single total; the backend rejects a totals query for it
 * with "missing required dimension".
 */
export const measureSupportsTotalsQuery = (measure: MetricsViewSpecMeasure) =>
  !measure.requiredDimensions?.length;

/**
 * Selects measure valid for current dashboard selections. We filter out advanced measures that are,
 * 1. Of type MEASURE_TYPE_TIME_COMPARISON.
 * 2. Dependent on a time dimension with a defined grain and not equal to the current selected grain.
 * 3. Window measures and measures with required dimensions, unless the query groups by the time dimension
 *    and that is the only dimension they require. See isMeasureSupported.
 *
 * queryHasTimeDimension is true for queries that group by the time dimension at the selected grain (time series)
 * and false for queries with no dimensions (totals) or a single categorical dimension (leaderboards, dimension table).
 */
export const filterOutSomeAdvancedMeasures = (
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  measureNames: string[],
  queryHasTimeDimension: boolean,
) => {
  const measuresSeen = new Set<string>();

  const ephemeralMeasureNames = new Set(
    exploreState.ephemeralMeasures?.map((def) => def.name) ?? [],
  );

  return measureNames.filter((measureName) => {
    // ephemeral measures are simple aggregations; always supported.
    if (ephemeralMeasureNames.has(measureName)) {
      if (measuresSeen.has(measureName)) return false;
      measuresSeen.add(measureName);
      return true;
    }
    const measureSpec = metricsViewSpec.measures?.find(
      (m) => m.name === measureName,
    );
    if (!measureSpec) return false;
    const measureIsSupported = isMeasureSupported(
      exploreState,
      metricsViewSpec,
      measureSpec,
      queryHasTimeDimension,
      true,
    );
    if (!measureIsSupported || measuresSeen.has(measureName)) return false;

    measuresSeen.add(measureName);
    return true;
  });
};

/**
 * Selects measure valid for current dashboard selections. We filter out advanced measures that are,
 * 1. Of type MEASURE_TYPE_TIME_COMPARISON.
 * 2. Dependent on a time dimension with a defined grain and not equal to the current selected grain.
 * 3. Window measures and measures with required dimensions, unless the query groups by the time dimension
 *    and that is the only dimension they require. See isMeasureSupported.
 *
 * This is a variant of the above but works on V1MetricsViewAggregationMeasure.
 * Once we move all queries to MetricsViewAggregation we dont need the above method.
 *
 * It is generic over the measure type so callers that pass spec measures get spec measures back.
 */
export const filterOutSomeAdvancedAggregationMeasures = <
  T extends AggregationMeasureRef,
>(
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  measures: T[],
  queryHasTimeDimension: boolean,
): T[] => {
  const measuresSeen = new Set<string>();

  const ephemeralMeasureNames = new Set(
    exploreState.ephemeralMeasures?.map((def) => def.name) ?? [],
  );

  return measures.filter((measure) => {
    const sourceMeasureName =
      measure.comparisonDelta?.measure ??
      measure.comparisonValue?.measure ??
      measure.comparisonRatio?.measure ??
      measure.percentOfTotal?.measure ??
      measure.name ??
      "";
    // Ephemeral expression measures are derived from measures already in the spec and have no spec
    // entry of their own, so there is no source measure to resolve or check for support.
    // The same applies to comparison measures derived from an ephemeral measure.
    const isEphemeral =
      (!!measure.expression && typeof measure.expression === "object") ||
      ephemeralMeasureNames.has(sourceMeasureName);
    if (!isEphemeral) {
      const measureSpec = metricsViewSpec.measures?.find(
        (m) => m.name === sourceMeasureName,
      );
      if (!measureSpec) return false;
      const measureIsSupported = isMeasureSupported(
        exploreState,
        metricsViewSpec,
        measureSpec,
        queryHasTimeDimension,
        false,
      );
      if (!measureIsSupported) return false;
    }

    if (measuresSeen.has(measure.name!)) return false;

    measuresSeen.add(measure.name!);
    return true;
  });
};

/**
 * Whether a measure can be queried with the dashboard's current selections.
 * The runtime rejects a query for a measure unless all of its required dimensions are grouped by,
 * and explore queries group by at most the time dimension.
 * So a measure that requires any other dimension is never supported,
 * and one that requires the time dimension is only supported when the query groups by it.
 */
const isMeasureSupported = (
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  measure: MetricsViewSpecMeasure,
  queryHasTimeDimension: boolean,
  allowTimeDependentMeasure: boolean,
) => {
  if (
    measure.type === MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON ||
    (!queryHasTimeDimension && measure.window)
  )
    return false;

  if (!measure.requiredDimensions?.length) return true;
  if (!queryHasTimeDimension) return false;

  // Time series queries group by the explore's selected time dimension, falling back to the metrics view's primary one.
  const timeDimension =
    exploreState.selectedTimeDimension || metricsViewSpec.timeDimension;
  return measure.requiredDimensions.every((reqDim) => {
    if (reqDim.name !== timeDimension) return false;

    const hasNoTimeGrain =
      !reqDim.timeGrain ||
      reqDim.timeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
    if (hasNoTimeGrain) return true;

    if (!allowTimeDependentMeasure) return false;
    return reqDim.timeGrain === exploreState.selectedTimeRange?.interval;
  });
};

export const measureSelectors = {
  /**
   * Get all measures in the dashboard.
   */
  allMeasures,

  /**
   * Returns a function that can be used to get a MetricsViewSpecMeasure
   * by name; this fn returns undefined if the dashboard has no measure with that name.
   */
  getMeasureByName,

  /**
   * Gets all visible measures in the dashboard.
   */
  visibleMeasures,
  /**
   * Get label for a measure by name
   */
  measureLabel,
  /**
   * Checks if the provided measure is a valid percent of total
   */
  isMeasureValidPercentOfTotal,

  filteredSimpleMeasures,
};
