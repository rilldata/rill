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
}: Pick<
  DashboardDataSources,
  "validMetricsView" | "validExplore"
>): MetricsViewSpecMeasure[] => {
  if (!validMetricsView?.measures || !validExplore?.measures) return [];

  return (
    validMetricsView.measures
      .filter((m) => validExplore.measures!.includes(m.name!))
      // Sort the filtered measures based on their order in validExplore.measures
      .sort(
        (a, b) =>
          validExplore.measures!.indexOf(a.name!) -
          validExplore.measures!.indexOf(b.name!),
      )
  );
};

export const visibleMeasures = ({
  validMetricsView,
  validExplore,
  dashboard,
}: DashboardDataSources): MetricsViewSpecMeasure[] => {
  if (!validMetricsView?.measures || !validExplore?.measures) return [];

  return dashboard.visibleMeasures
    .map((mes) => validMetricsView.measures?.find((m) => m.name === mes))
    .filter(Boolean) as MetricsViewSpecMeasure[];
};

export const getMeasureByName = (
  dashData: DashboardDataSources,
): ((name: string | undefined) => MetricsViewSpecMeasure | undefined) => {
  return (name: string | undefined) => {
    return allMeasures(dashData)?.find((measure) => measure.name === name);
  };
};

export const measureLabel = ({
  validMetricsView,
}: DashboardDataSources): ((m: string) => string) => {
  return (measureName) => {
    const measure = validMetricsView?.measures?.find(
      (d) => d.name === measureName,
    );
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
  return () => {
    if (!validMetricsView?.measures || !validExplore?.measures) return [];

    return (
      validMetricsView.measures
        .filter(
          (m) => validExplore.measures!.includes(m.name!) && isSimpleMeasure(m),
        )
        // Sort the filtered measures based on their order in validExplore.measures
        .sort(
          (a, b) =>
            validExplore.measures!.indexOf(a.name!) -
            validExplore.measures!.indexOf(b.name!),
        )
    );
  };
};

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
 * Selects measure valid for current dashboard selections. We filter out advanced measures that are,
 * 1. Of type MEASURE_TYPE_TIME_COMPARISON.
 * 2. Dependent on a time dimension with a defined grain and not equal to the current selected grain.
 * 3. Window measures if includeWindowMeasures=false. Right now totals query does not support these.
 */
export const filterOutSomeAdvancedMeasures = (
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  measureNames: string[],
  includeWindowMeasures: boolean,
) => {
  const measuresSeen = new Set<string>();

  return measureNames.filter((measureName) => {
    const measureSpec = metricsViewSpec.measures?.find(
      (m) => m.name === measureName,
    );
    if (!measureSpec) return false;
    const measureIsSupported = isMeasureSupported(
      exploreState,
      measureSpec,
      includeWindowMeasures,
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
 * 3. Window measures if includeWindowMeasures=false. Right now totals query does not support these.
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
  includeWindowMeasures: boolean,
): T[] => {
  const measuresSeen = new Set<string>();

  return measures.filter((measure) => {
    // Ephemeral expression measures are derived from measures already in the spec and have no spec
    // entry of their own, so there is no source measure to resolve or check for support.
    const isEphemeralExpression =
      !!measure.expression && typeof measure.expression === "object";
    if (!isEphemeralExpression) {
      const sourceMeasureName =
        measure.comparisonDelta?.measure ??
        measure.comparisonValue?.measure ??
        measure.comparisonRatio?.measure ??
        measure.percentOfTotal?.measure ??
        measure.name ??
        "";
      const measureSpec = metricsViewSpec.measures?.find(
        (m) => m.name === sourceMeasureName,
      );
      if (!measureSpec) return false;
      const measureIsSupported = isMeasureSupported(
        exploreState,
        measureSpec,
        includeWindowMeasures,
        false,
      );
      if (!measureIsSupported) return false;
    }

    if (measuresSeen.has(measure.name!)) return false;

    measuresSeen.add(measure.name!);
    return true;
  });
};

const isMeasureSupported = (
  exploreState: ExploreState,
  measure: MetricsViewSpecMeasure,
  allowWindowMeasure: boolean,
  allowTimeDependentMeasure: boolean,
) => {
  if (
    measure.type === MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON ||
    (!allowWindowMeasure && measure.window)
  )
    return false;

  const allDependentDimensionsAllowed =
    measure.requiredDimensions?.every((reqDim) => {
      const hasNoTimeGrain =
        !reqDim.timeGrain ||
        reqDim.timeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
      if (hasNoTimeGrain) return true;

      if (!allowTimeDependentMeasure) return false;
      return reqDim.timeGrain === exploreState.selectedTimeRange?.interval;
    }) ?? true;
  return allDependentDimensionsAllowed;
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
