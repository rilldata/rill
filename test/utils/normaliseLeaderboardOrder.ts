import type { LeaderboardValues } from "$lib/application-state-stores/explorer-stores";

const FuzzyGroupLargeThreshold = 500;
const FuzzyGroupSmallThreshold = 0.25;
const FuzzyGroupSmallNumber = 10;

/**
 * Normalises leaderboard values for a list of columns.
 * Calls {@link normaliseArrayValues} for each column.
 */
export function normaliseLeaderboardOrder(
  leaderboard: Array<LeaderboardValues>,
  measureSqlName: string
): Array<[string, Array<string>]> {
  return leaderboard.map((l) => [
    l.dimensionName,
    normaliseArrayValues(
      l.values.map((val) => ({
        value: val[measureSqlName],
        label: val[l.dimensionName],
      }))
    ),
  ]);
}

/**
 * Normalises leaderboard values. Expects values to be sorted.
 * Finds sequence of labels with close values and sorts them alphabetically.
 * Retails these sequences in order.
 * EG: [L3: 4500, L1: 4300, L2: 2120, L0: 2090] => [L1, l3, L0, L2]
 */
function normaliseArrayValues(
  values: Array<{ value: number; label: string }>
): Array<string> {
  const normalisedValues = new Array<string>();

  let fuzzyGroupAvg = values[0].value;
  let fuzzyGroup = new Array<string>(values[0].label);

  // these thresholds are based on the data used.
  // they are in no way a generic threshold
  const FuzzyGroupThreshold =
    fuzzyGroupAvg <= FuzzyGroupSmallNumber
      ? FuzzyGroupSmallThreshold
      : FuzzyGroupLargeThreshold;

  const addToNormalValues = () => {
    normalisedValues.push(...fuzzyGroup.sort());
    fuzzyGroup = [];
  };

  for (let i = 1; i < values.length; i++) {
    if (Math.abs(fuzzyGroupAvg - values[i].value) > FuzzyGroupThreshold) {
      addToNormalValues();
      fuzzyGroupAvg = values[i].value;
    }
    fuzzyGroup.push(values[i].label);
  }

  addToNormalValues();

  return normalisedValues;
}
