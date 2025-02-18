import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecMeasureV2,
  type V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

export const allMeasures = ({
  validMetricsView,
  validExplore,
}: Pick<
  DashboardDataSources,
  "validMetricsView" | "validExplore"
>): MetricsViewSpecMeasureV2[] => {
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

export const leaderboardMeasureName = ({ dashboard }: DashboardDataSources) => {
  return dashboard.leaderboardMeasureName;
};

export const visibleMeasures = ({
  validMetricsView,
  validExplore,
  dashboard,
}: DashboardDataSources): MetricsViewSpecMeasureV2[] => {
  if (!validMetricsView?.measures || !validExplore?.measures) return [];

  return (
    validMetricsView.measures
      .filter((m) => dashboard.visibleMeasureKeys.has(m.name!))
      // Sort the filtered measures based on their order in validExplore.measures
      .sort(
        (a, b) =>
          validExplore.measures!.indexOf(a.name!) -
          validExplore.measures!.indexOf(b.name!),
      )
  );
};

export const getMeasureByName = (
  dashData: DashboardDataSources,
): ((name: string | undefined) => MetricsViewSpecMeasureV2 | undefined) => {
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
          (m) =>
            validExplore.measures!.includes(m.name!) &&
            !m.window &&
            m.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON,
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

/**
 * Selects measure valid for current dashboard selections. We filter out advanced measures that are,
 * 1. Of type MEASURE_TYPE_TIME_COMPARISON.
 * 2. Dependent on a time dimension with a defined grain and not equal to the current selected grain.
 * 3. Window measures if includeWindowMeasures=false. Right now totals query does not support these.
 */
export const filterSomeAdvancedMeasures = (
  exploreState: MetricsExplorerEntity,
  metricsViewSpec: V1MetricsViewSpec,
  measureNames: string[],
  includeWindowMeasures: boolean,
) => {
  return measureNames.filter((measureName) => {
    const measure = metricsViewSpec.measures?.find(
      (m) => m.name === measureName,
    );
    if (
      !measure ||
      measure.type ===
        MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON ||
      (!includeWindowMeasures && measure.window)
      // TODO: we need to send a single query for this support
      // (measure.type ===
      //   MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON &&
      //   (!dashboard.showTimeComparison ||
      //     !dashboard.selectedComparisonTimeRange))
    )
      return false;

    let skipMeasure = false;
    measure.requiredDimensions?.forEach((reqDim) => {
      if (
        reqDim.timeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED &&
        reqDim.timeGrain !== exploreState.selectedTimeRange?.interval
      ) {
        // filter out measures with dependant dimensions not matching the selected grain
        skipMeasure = true;
        return;
      }
    });
    return !skipMeasure;
  });
};

export const getSimpleMeasures = (measures: MetricsViewSpecMeasureV2[]) => {
  return (
    measures?.filter(
      (m) =>
        !m.window &&
        m.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON,
    ) ?? []
  );
};

export const measureSelectors = {
  /**
   * Get all measures in the dashboard.
   */
  allMeasures,

  /**
   * Returns a function that can be used to get a MetricsViewSpecMeasureV2
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

  leaderboardMeasureName,
};
