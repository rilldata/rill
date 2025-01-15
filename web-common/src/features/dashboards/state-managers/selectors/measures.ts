import {
  type MetricsViewSpecDimensionSelector,
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
  return (
    validMetricsView?.measures?.filter((m) =>
      validExplore?.measures?.includes(m.name ?? ""),
    ) ?? []
  );
};

export const leaderboardMeasureName = ({ dashboard }: DashboardDataSources) => {
  return dashboard.leaderboardMeasureName;
};

export const visibleMeasures = ({
  validMetricsView,
  dashboard,
}: DashboardDataSources): MetricsViewSpecMeasureV2[] => {
  return (
    validMetricsView?.measures?.filter(
      (m) => m.name && dashboard.visibleMeasureKeys.has(m.name),
    ) ?? []
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
    return (
      validMetricsView?.measures?.filter(
        (m) =>
          !m.window &&
          m.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON &&
          validExplore?.measures?.includes(m.name ?? ""),
      ) ?? []
    );
  };
};

/**
 * Selects measure valid for current dashboard selections.
 * Also includes additional dimensions needed for any advanced measures.
 */
export const getFilteredMeasuresAndDimensions = ({
  dashboard,
}: Pick<DashboardDataSources, "dashboard">) => {
  return (
    metricsViewSpec: V1MetricsViewSpec,
    measureNames: string[],
  ): {
    measures: string[];
    dimensions: MetricsViewSpecDimensionSelector[];
  } => {
    const dimensions = new Map<string, V1TimeGrain>();
    const measures = new Set<string>();
    measureNames.forEach((measureName) => {
      const measure = metricsViewSpec.measures?.find(
        (m) => m.name === measureName,
      );
      if (
        !measure ||
        measure.type === MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON
        // TODO: we need to send a single query for this support
        // (measure.type ===
        //   MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON &&
        //   (!dashboard.showTimeComparison ||
        //     !dashboard.selectedComparisonTimeRange))
      )
        return;

      let skipMeasure = false;
      measure.requiredDimensions?.forEach((reqDim) => {
        if (
          reqDim.timeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED &&
          reqDim.timeGrain !== dashboard.selectedTimeRange?.interval
        ) {
          // filter out measures with dependant dimensions not matching the selected grain
          skipMeasure = true;
          return;
        }
        if (!reqDim.name) return;

        const existingEntry = dimensions.get(reqDim.name);
        if (existingEntry) {
          if (existingEntry === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) {
            dimensions.set(
              reqDim.name,
              reqDim.timeGrain ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
            );
          } else {
            // mismatching measures are requested
            skipMeasure = true;
          }
          return;
        }

        dimensions.set(
          reqDim.name,
          reqDim.timeGrain ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        );
      });
      if (skipMeasure) return;
      measures.add(measureName);
      measure.referencedMeasures?.filter((refMes) => measures.add(refMes));
    });
    return {
      measures: [...measures],
      dimensions: [...dimensions.entries()].map(([name, timeGrain]) => ({
        name,
        timeGrain,
      })),
    };
  };
};

export const getIndependentMeasures = (
  metricsViewSpec: V1MetricsViewSpec,
  measureNames: string[],
) => {
  const measures = new Set<string>();
  measureNames.forEach((measureName) => {
    const measure = metricsViewSpec.measures?.find(
      (m) => m.name === measureName,
    );
    // temporary check for window measures until the PR to move to AggregationApi is merged
    if (!measure || measure.requiredDimensions?.length || !!measure.window)
      return;
    measures.add(measureName);
    measure.referencedMeasures?.filter((refMes) => measures.add(refMes));
  });
  return [...measures];
};

export const getFilteredSimpleMeasures = (
  measures: MetricsViewSpecMeasureV2[],
) => {
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

  getFilteredMeasuresAndDimensions,
  leaderboardMeasureName,
};
