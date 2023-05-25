import { NicelyFormattedTypes, humanizeDataType } from "../humanize-numbers";
import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  formatMeasurePercentageDifference,
  // humanizeDataType,
} from "../humanize-numbers";

// /**
//  * LeaderboardRenderGlobal contains items that affect all leaderboards,
//  * and may be relevant to the rendering of any individual leaderboard entries.
//  */
export type LeaderboardRenderGlobal = {
  // total number of rows captured by current set of filters
  totalFilteredRowCount: number;
  // whether the active measure is summable
  activeMeasureIsSummable: boolean;
  // the numeric formatting preset for the active measure
  formatPreset: NicelyFormattedTypes;
  // the active metricViewName
  metricViewName: string;
};

/**
 * PerLeaderboardData contains data that is specific to a single
 * leaderboard, but may be relevant to the rendering of all
 *  entries in that leaderboard.
 */
export type PerLeaderboardData = {
  // whether the leaderboard is currently in "exclude" mode.
  // true means that the leaderboard is in "exclude" mode,
  // false means that the leaderboard is in "include" mode.
  excludeMode: boolean;
  // whether there is at least one active (selected) value
  // in this leaderboard.
  atLeastOneActive: boolean;
};

/**
 * LeaderboardRenderValue is the data that an individual
 * leaderboard entry requires to be rendered.
 */
export type LeaderboardRenderValue = {
  // the label of the leaderboard entry
  label: string;
  // the numeric value of the leaderboard entry for the current time period
  value: number;
  // the formatted string value for the current leaderboard entry
  formattedValue: string;
  // the nmber of rows included in the current set of filters
  // for this leaderboard entry
  rowCount: number;

  // whether this leaderboard entry is currently *selected*, independent of whether it is included or excluded
  active: boolean;
  // whether this leaderboard entry should be excluded from
  // the current set of filters
  excluded: boolean;

  // whether the comparison value should be shown for this leaderboard entry
  showComparisonForThisValue: boolean;
  // the numeric value of the leaderboard entry for the comparison time period
  comparisonValue: number;
};

export function valuesToRenderValues(
  values,
  activeValues,
  comparisonMap,
  comparisonLabelToReveal,
  filterExcludeMode,
  atLeastOneActive,
  formatPreset
): LeaderboardRenderValue[] {
  return values.map((v) => {
    const active = activeValues.findIndex((value) => value === v.label) >= 0;
    const comparisonValue = comparisonMap.get(v.label);

    // Super important special case: if there is not at least one "active" (selected) value,
    // we need to set *all* items to be included, because by default if a user has not
    // selected any values, we assume they want all values included in all calculations.
    const excluded = atLeastOneActive
      ? (filterExcludeMode && active) || (!filterExcludeMode && !active)
      : false;

    return {
      ...v,
      active,
      excluded,
      comparisonValue,
      formattedValue: humanizeDataType(v.value, formatPreset),
      showComparisonForThisValue: comparisonLabelToReveal === v.label,
    };
  });
}

export function getFormatterValueForPercDiff(comparisonValue, value) {
  if (comparisonValue === 0) return PERC_DIFF.PREV_VALUE_ZERO;
  if (!comparisonValue) return PERC_DIFF.PREV_VALUE_NO_DATA;
  if (value === null || value === undefined)
    return PERC_DIFF.CURRENT_VALUE_NO_DATA;

  const percDiff = (value - comparisonValue) / comparisonValue;
  return formatMeasurePercentageDifference(percDiff);
}

export function isActiveMeasureSummable(activeMeasure) {
  return (
    activeMeasure?.expression.toLowerCase()?.includes("count(") ||
    activeMeasure?.expression?.toLowerCase()?.includes("sum(")
  );
}
