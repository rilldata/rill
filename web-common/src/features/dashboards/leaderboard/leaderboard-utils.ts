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
  selected: boolean;
};

export function prepareLeaderboardItemData(
  values: { value: number; label: string | number }[],
  selectedValues: (string | number)[],
  comparisonMap: Map<string | number, number>
): LeaderboardItemData[] {
  return values.map((v) => {
    const selected =
      selectedValues.findIndex((value) => value === v.label) >= 0;
    const comparisonValue = comparisonMap.get(v.label);

    return {
      ...v,
      selected,
      comparisonValue,
    };
  });
}

export const CONTEXT_COLUMN_WIDTH = 44;

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
  let formattedValue = "";

  if (contextType === LeaderboardContextColumn.DELTA_PERCENT) {
    formattedValue = getFormatterValueForPercDiff(
      value && comparisonValue ? value - comparisonValue : null,
      comparisonValue
    );
  } else if (contextType === LeaderboardContextColumn.PERCENT) {
    formattedValue = getFormatterValueForPercDiff(value, unfilteredTotal);
  } else if (contextType === LeaderboardContextColumn.DELTA_ABSOLUTE) {
    formattedValue = humanizeDataType(
      value && comparisonValue ? value - comparisonValue : null,
      formatPreset
    );
  } else {
    formattedValue = "";
  }
  return formattedValue;
}
