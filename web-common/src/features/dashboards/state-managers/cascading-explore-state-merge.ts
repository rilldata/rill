import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";

const ShallowMergeOneLevelDeepKeys = new Set<keyof ExploreState>([
  "selectedComparisonTimeRange",
  "tdd",
  "pivot",
]);

/**
 * Performs a cascading merge of provided explore states in order.
 * Applies a shallow merge for all keys except those in ShallowMergeOneLevelDeepKeys.
 * Keys in ShallowMergeOneLevelDeepKeys are shallow merged at their own level.
 *
 * Most state values are literals, so shallow/deep merging makes no difference for them.
 * Non-literal values like filters, visible measures/dimensions should come entirely from a single source.
 * For example: We cannot have a publisher filter from one source and a domain filter from another source.
 * These should come from the first source that contains filters.
 *
 * The exception is keys in ShallowMergeOneLevelDeepKeys which need to be shallow merged separately at their level.
 * This is necessary because these state values contain nested properties one level deep.
 * For example: selectedTimeRange contains both selected time range and grain.
 * tdd and pivot have their respective values underneath those keys.
 *
 * This would be avoided if each subsection of the state have their own classes.
 * Then we could offload the specific merging logic to the class methods.
 */
export function cascadingExploreStateMerge(
  exploreStatesInOrder: Partial<ExploreState>[],
) {
  const mergedExploreState: Partial<ExploreState> = {};

  const keyProcessed = new Set<string>();
  // Merge all keys not part of ShallowMergeOneLevelDeepKeys. This allows for future keys to be merged without changes.
  exploreStatesInOrder.forEach((state) => {
    Object.keys(state).forEach((key: keyof ExploreState) => {
      // Since the states are in order a key found 1st should only be merged once.
      // So ignore keys we have already seen
      const isKeyAlreadyProcessed = keyProcessed.has(key);

      // Ignore keys that are shallow merged a level deep, they are merged separately
      const isKeyForShallowMergeOneLevelDeep =
        ShallowMergeOneLevelDeepKeys.has(key);

      if (isKeyAlreadyProcessed || isKeyForShallowMergeOneLevelDeep) return;

      const value = state[key];
      // Skip undefined/null values these are values that are not set in the state,
      // but because of certain deserializers, and it's needing to be backwards compatible we need this check.
      if (value === undefined || value === null) return;

      keyProcessed.add(key);
      mergedExploreState[key] = value as any;
    });
  });

  // Merge certain keys that are one level deep, these are merged as a shallow merge but one level deep.
  ShallowMergeOneLevelDeepKeys.forEach((levelOneKey) => {
    const oneLevelDeepState = {};

    // check if the 1st value present is undefined. this means it was an unset of the param
    const firstMatchingState = exploreStatesInOrder.find((o) => {
      return levelOneKey in o;
    });

    // none of the states has the key. do not set it in the final state
    if (!firstMatchingState) return;

    // if the first state containing the key had undefined then set undefined and return
    if (firstMatchingState[levelOneKey] === undefined) {
      mergedExploreState[levelOneKey] = undefined;
      return;
    }

    // else merge them in reverse order so that the state higher in the array are merged last
    for (let i = exploreStatesInOrder.length - 1; i >= 0; i--) {
      if (!exploreStatesInOrder[i]?.[levelOneKey]) continue;
      Object.assign(oneLevelDeepState, exploreStatesInOrder[i][levelOneKey]);
    }

    mergedExploreState[levelOneKey] = oneLevelDeepState as any;
  });

  return mergedExploreState;
}
