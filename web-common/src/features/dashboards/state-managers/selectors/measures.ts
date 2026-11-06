import { appendEphemeralSpecMeasures } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecMeasure,
  type V1MetricsViewSpec,
  V1TimeGrain,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationMeasureComputeExpression,
  type V1MetricsViewAggregationSort,
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

export const isSimpleMeasure = (measure: MetricsViewSpecMeasure) =>
  !measure.window &&
  measure.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON;

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
 * The dimensions an explore query groups by.
 * An empty object is a query with no dimensions, such as totals.
 */
export type QueryGroupBy = {
  // Categorical dimensions the query groups by.
  dimensions?: string[];
  // Whether the query groups by the explore's time dimension at the selected grain, like the time series.
  time?: boolean;
};

/**
 * Whether a query grouped by `groupBy` satisfies the measure's required dimensions.
 * The runtime rejects a measure unless the query groups by every dimension in its required_dimensions,
 * and by a required time dimension at the required grain.
 */
export function queryGroupsByRequiredDimensions(
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  measure: MetricsViewSpecMeasure,
  groupBy: QueryGroupBy,
): boolean {
  // Time grouped queries use the explore's selected time dimension, falling back to the metrics view's primary one.
  const timeDimension =
    exploreState.selectedTimeDimension || metricsViewSpec.timeDimension;
  return (measure.requiredDimensions ?? []).every((reqDim) => {
    const hasGrain =
      !!reqDim.timeGrain &&
      reqDim.timeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
    if (groupBy.time && reqDim.name === timeDimension) {
      return (
        !hasGrain ||
        reqDim.timeGrain === exploreState.selectedTimeRange?.interval
      );
    }
    return (
      !hasGrain && !!reqDim.name && !!groupBy.dimensions?.includes(reqDim.name)
    );
  });
}

/**
 * Selects measure valid for current dashboard selections. We filter out advanced measures that are,
 * 1. Of type MEASURE_TYPE_TIME_COMPARISON.
 * 2. Window measures, unless the query groups by the time dimension.
 * 3. Measures whose required dimensions the query does not group by. See queryGroupsByRequiredDimensions.
 */
export const filterOutSomeAdvancedMeasures = (
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  measureNames: string[],
  groupBy: QueryGroupBy,
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
      groupBy,
    );
    if (!measureIsSupported || measuresSeen.has(measureName)) return false;

    measuresSeen.add(measureName);
    return true;
  });
};

/**
 * Selects measure valid for current dashboard selections. We filter out advanced measures that are,
 * 1. Of type MEASURE_TYPE_TIME_COMPARISON.
 * 2. Window measures, unless the query groups by the time dimension.
 * 3. Measures whose required dimensions the query does not group by. See queryGroupsByRequiredDimensions.
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
  groupBy: QueryGroupBy,
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
        groupBy,
      );
      if (!measureIsSupported) return false;
    }

    if (measuresSeen.has(measure.name!)) return false;

    measuresSeen.add(measure.name!);
    return true;
  });
};

/**
 * Measures and sort for a query that groups only by `dimensionName`,
 * such as the dimension table export and the TDD comparison table's top list.
 * Measures the query cannot satisfy (e.g. a rolling window ordered by the time dimension) are omitted,
 * and the query sorts by the dimension when the sort measure is one of them.
 * sortBy is the measure to sort by, or the dimension itself when sorting by dimension.
 */
export function getDimensionQueryMeasuresAndSort<
  T extends AggregationMeasureRef,
>(
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  measures: T[],
  dimensionName: string,
  sortBy: string | undefined,
): { measures: T[]; sort: V1MetricsViewAggregationSort[] } {
  const supportedMeasures = filterOutSomeAdvancedAggregationMeasures(
    exploreState,
    metricsViewSpec,
    measures,
    { dimensions: [dimensionName] },
  );
  const sortByIsSupported = supportedMeasures.some((m) => m.name === sortBy);
  return {
    measures: supportedMeasures,
    sort: [
      {
        desc: exploreState.sortDirection === SortDirection.DESCENDING,
        name: sortByIsSupported ? sortBy : dimensionName,
      },
    ],
  };
}

const isMeasureSupported = (
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  measure: MetricsViewSpecMeasure,
  groupBy: QueryGroupBy,
) =>
  measure.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON &&
  (!measure.window || !!groupBy.time) &&
  queryGroupsByRequiredDimensions(
    exploreState,
    metricsViewSpec,
    measure,
    groupBy,
  );

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
