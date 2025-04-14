import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";

const ShallowMergeKeys: (keyof MetricsExplorerEntity)[] = [
  "activePage",

  "visibleMeasures",
  "allMeasuresVisible",
  "visibleDimensions",
  "allDimensionsVisible",
  "leaderboardSortByMeasureName",
  "leaderboardMeasureCount",
  "dashboardSortType",
  "sortDirection",

  "whereFilter",
  "dimensionsWithInlistFilter",
  "dimensionThresholdFilters",

  "selectedScrubRange",
  "selectedTimezone",
  "showTimeComparison",

  "selectedDimensionName",
];
const OneLevelDeepShallowMergeKeys: (keyof MetricsExplorerEntity)[] = [
  "selectedTimeRange",
  "selectedComparisonTimeRange",
  "tdd",
  "pivot",
];

export function cascadingExploreStateMerge(
  exploreStatesInOrder: Partial<MetricsExplorerEntity>[],
) {
  const finalExplorePreset: Partial<MetricsExplorerEntity> = {};

  ShallowMergeKeys.forEach((key) => {
    const firstMatchingState = exploreStatesInOrder.find((o) => {
      const v = o[key];
      return v !== undefined && v !== null;
    });
    if (!firstMatchingState) return;

    (finalExplorePreset as any)[key] = firstMatchingState[key];
  });

  OneLevelDeepShallowMergeKeys.forEach((levelOneKey) => {
    const oneLevelDeepState = {} as any;

    // check if the 1st value present is undefined. this means it was an unset of the param
    const firstMatchingState = exploreStatesInOrder.find((o) => {
      return levelOneKey in o;
    });
    // none of the states has the key. do not set it in the final state
    if (!firstMatchingState) return;
    // if the first state containing the key had undefined then set undefined and return
    if (firstMatchingState[levelOneKey] === undefined) {
      finalExplorePreset[levelOneKey] = undefined;
      return;
    }

    // else merge them in reverse order so that the state higher in the array are merged last
    for (let i = exploreStatesInOrder.length - 1; i >= 0; i--) {
      if (!exploreStatesInOrder[i][levelOneKey]) continue;
      Object.assign(oneLevelDeepState, exploreStatesInOrder[i][levelOneKey]);
    }

    finalExplorePreset[levelOneKey] = oneLevelDeepState;
  });

  return finalExplorePreset;
}
