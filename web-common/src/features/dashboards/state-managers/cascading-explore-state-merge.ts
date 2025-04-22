import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";

const OneLevelDeepShallowMergeKeys = new Set<keyof MetricsExplorerEntity>([
  "selectedTimeRange",
  "selectedComparisonTimeRange",
  "tdd",
  "pivot",
]);

export function cascadingExploreStateMerge(
  exploreStatesInOrder: Partial<MetricsExplorerEntity>[],
) {
  const finalExploreState: Partial<MetricsExplorerEntity> = {};

  const shallowKeyProcessed = new Set<string>();
  // Merge all keys not part of OneLevelDeepShallowMergeKeys. This allows for future keys to be merged without changes.
  exploreStatesInOrder.forEach((state) => {
    Object.keys(state).forEach((key: keyof MetricsExplorerEntity) => {
      // Since the states are in order a key found 1st should only be merged once.
      // So ignore keys we have already seen
      const isKeyAlreadyProcessed = shallowKeyProcessed.has(key);

      // Ignore one level deep merges, they are merged separately
      const isKeyForDeepMerge = OneLevelDeepShallowMergeKeys.has(key);

      if (isKeyAlreadyProcessed || isKeyForDeepMerge) return;

      const value = state[key];
      // Skip undefined/null values these are values that are not set in the state,
      // but because of certain deserializers, and it's needing to be backwards compatible we need this check.
      if (value === undefined || value === null) return;

      shallowKeyProcessed.add(key);
      finalExploreState[key] = value as any;
    });
  });

  // Merge keys that are one level deep, these are merged as a shallow merge but one level deep.
  OneLevelDeepShallowMergeKeys.forEach((levelOneKey) => {
    const oneLevelDeepState = {};

    // check if the 1st value present is undefined. this means it was an unset of the param
    const firstMatchingState = exploreStatesInOrder.find((o) => {
      return levelOneKey in o;
    });
    // none of the states has the key. do not set it in the final state
    if (!firstMatchingState) return;
    // if the first state containing the key had undefined then set undefined and return
    if (firstMatchingState[levelOneKey] === undefined) {
      finalExploreState[levelOneKey] = undefined;
      return;
    }

    // else merge them in reverse order so that the state higher in the array are merged last
    for (let i = exploreStatesInOrder.length - 1; i >= 0; i--) {
      if (!exploreStatesInOrder[i]?.[levelOneKey]) continue;
      Object.assign(oneLevelDeepState, exploreStatesInOrder[i][levelOneKey]);
    }

    finalExploreState[levelOneKey] = oneLevelDeepState as any;
  });

  return finalExploreState;
}
