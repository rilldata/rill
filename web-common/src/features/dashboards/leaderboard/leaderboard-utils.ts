import { PERC_DIFF } from "../../../components/data-types/type-utils";
import { formatMeasurePercentageDifference } from "../humanize-numbers";

export function getFormatterValueForPercDiff(numerator, denominator) {
  if (denominator === 0) return PERC_DIFF.PREV_VALUE_ZERO;
  if (!denominator) return PERC_DIFF.PREV_VALUE_NO_DATA;
  if (numerator === null || numerator === undefined)
    return PERC_DIFF.CURRENT_VALUE_NO_DATA;

  const percDiff = numerator / denominator;
  return formatMeasurePercentageDifference(percDiff);
}

export type LeaderboardItemData = {
  label: string | number;
  value: number;
  comparisonValue: number;
  // selection is not enough to determine if the item is included
  // or excluded; for that we need to know the leaderboard's
  // include/exclude state
  selectedIndex: number;
  defaultComparedIndex: number;
};

export function prepareLeaderboardItemData(
  values: { value: number; label: string | number }[],
  selectedValues: (string | number)[],
  comparisonMap: Map<string | number, number>,
  excludeMode: boolean
): LeaderboardItemData[] {
  let count = 0;

  return values.map((v) => {
    const selectedIndex = selectedValues.findIndex(
      (value) => value === v.label
    );
    const comparisonValue = comparisonMap.get(v.label);

    // Tag values which will be compared by default
    let defaultComparedIndex = -1;
    if (!excludeMode && count < 3 && !selectedValues.length) {
      defaultComparedIndex = count;
      count = count + 1;
    } else if (excludeMode && count < 3) {
      if (selectedIndex === -1) {
        defaultComparedIndex = count;
        count += 1;
      }
    }
    return {
      ...v,
      selectedIndex,
      comparisonValue,
      defaultComparedIndex,
    };
  });
}

export const CONTEXT_COLUMN_WIDTH = 44;
