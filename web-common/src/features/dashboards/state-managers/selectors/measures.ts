import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecMeasure,
  type V1MetricsViewSpec,
  V1TimeGrain,
  type V1MetricsViewAggregationMeasure,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

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
 */
export const filterOutSomeAdvancedAggregationMeasures = (
  exploreState: ExploreState,
  metricsViewSpec: V1MetricsViewSpec,
  measures: V1MetricsViewAggregationMeasure[],
  includeWindowMeasures: boolean,
) => {
  const measuresSeen = new Set<string>();

  return measures.filter((measure) => {
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
    if (!measureIsSupported || measuresSeen.has(measure.name!)) return false;

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
