import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  FormatPreset,
  formatMeasurePercentageDifference,
  humanizeDataType,
} from "../humanize-numbers";
import { LeaderboardContextColumn } from "../leaderboard-context-column";

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
  // main value to be shown in the leaderboard
  value: number;
  // the comparison value, which may be either the previous value
  // (used to calculate the absolute or percentage change) or
  // the measure total (used to calculate the percentage of total)
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
  excludeMode: boolean,
  initalCount = 0
): LeaderboardItemData[] {
  let count = initalCount;

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

/**
 * Returns the formatted value for the context column
 * given the
 * accounting for the context column type.
 */
export function formatContextColumnValue(
  itemData: LeaderboardItemData,
  unfilteredTotal: number,
  contextType: LeaderboardContextColumn,
  formatPreset: FormatPreset
): string {
  const { value, comparisonValue } = itemData;

  switch (contextType) {
    case LeaderboardContextColumn.DELTA_ABSOLUTE: {
      const delta = value && comparisonValue ? value - comparisonValue : null;
      let formattedValue = humanizeDataType(delta, formatPreset);
      if (delta && delta > 0) {
        formattedValue = "+" + formattedValue;
      }
      return formattedValue;
    }
    case LeaderboardContextColumn.DELTA_PERCENT:
      return getFormatterValueForPercDiff(
        value && comparisonValue ? value - comparisonValue : null,
        comparisonValue
      );
    case LeaderboardContextColumn.PERCENT:
      return getFormatterValueForPercDiff(value, unfilteredTotal);
    case LeaderboardContextColumn.HIDDEN:
      return "";
    default:
      throw new Error("Invalid context column, all cases must be handled");
  }
}
