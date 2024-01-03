import {
  V1MetricsViewComparisonMeasureType as ApiSortType,
  type V1MetricsViewComparisonRow,
  type V1MetricsViewComparisonValue,
} from "@rilldata/web-common/runtime-client";

import { SortType } from "../proto-state/derived-types";

/**
 * A `V1MetricsViewComparisonRow` basically represents a row of data
 * in the *dimension detail table*, NOT in the leaderboard. Therefore,
 * to convert to rows of leaderboard data, we need to extract a single
 * measure from the dimension table shaped data (namely, the active
 * measure in the leaderboard).
 */
export function getLabeledComparisonFromComparisonRow(
  row: V1MetricsViewComparisonRow,
  measureName: string | number
): ComparisonValueWithLabel {
  const measure = row.measureValues?.find((v) => v.measureName === measureName);
  if (!measure) {
    throw new Error(
      `Could not find measure ${measureName} in row ${JSON.stringify(row)}`
    );
  }
  return {
    dimensionValue: row.dimensionValue as string,
    ...measure,
  };
}

export type LeaderboardItemData = {
  /**
   *The dimension value label to be shown in the leaderboard
   */
  dimensionValue: string;

  /**
   *  main value to be shown in the leaderboard
   */
  value: number | null;

  /**
   * Percent of total for summable measures; null if not summable.
   * Note that this value will be between 0 and 1, not 0 and 100.
   */
  pctOfTotal: number | null;

  /**
   *  The value from the comparison period.
   * Techinally this might not be a "previous value" but
   * we use that name as a shorthand, since it's the most
   * common use case.
   */
  prevValue: number | null;
  /**
   *
   * the relative change from the previous value
   * note that this needs to be multiplied by 100 to get
   * the percentage change
   */
  deltaRel: number | null;

  /**
   *  the absolute change from the previous value
   */
  deltaAbs: number | null;

  /**
   *  This tracks the order in which an item was selected,
   * which is used to maintain a mapping between the color
   * of the line in the charts and the icon in the
   * leaderboard/dimension detail table.
   * Will be -1 if the item is not selected.
   * FIXME: this should be nullable rather than using -1 sentinel value!!!
   */
  selectedIndex: number;
};

function cleanUpComparisonValue(
  v: ComparisonValueWithLabel,
  total: number | null,
  selectedIndex: number
): LeaderboardItemData {
  if (!(Number.isFinite(v.baseValue) || v.baseValue === null)) {
    throw new Error(
      `Leaderboards only implemented for numeric baseValues or missing data (null). Got: ${JSON.stringify(
        v
      )}`
    );
  }
  const value = v.baseValue === null ? null : (v.baseValue as number);

  return {
    dimensionValue: v.dimensionValue,
    value,
    pctOfTotal: total !== null && value !== null ? value / total : null,
    prevValue: Number.isFinite(v.comparisonValue)
      ? (v.comparisonValue as number)
      : null,
    deltaRel: Number.isFinite(v.deltaRel) ? (v.deltaRel as number) : null,
    deltaAbs: Number.isFinite(v.deltaAbs) ? (v.deltaAbs as number) : null,
    selectedIndex,
  };
}

/**
 * A `V1MetricsViewComparisonValue` augmented with the dimension
 * value that it corresponds to.
 */
type ComparisonValueWithLabel = V1MetricsViewComparisonValue & {
  dimensionValue: string;
};

/**
 *
 * @param values
 * @param selectedValues
 * @param total: the total of the measure for the current period,
 * or null if the measure is not valid_percent_of_total
 * @returns
 */
export function prepareLeaderboardItemData(
  values: ComparisonValueWithLabel[],
  numberAboveTheFold: number,
  selectedValues: string[],
  total: number | null
): {
  aboveTheFold: LeaderboardItemData[];
  selectedBelowTheFold: LeaderboardItemData[];
  noAvailableValues: boolean;
  showExpandTable: boolean;
} {
  const aboveTheFold: LeaderboardItemData[] = [];
  const selectedBelowTheFold: LeaderboardItemData[] = [];

  // we keep a copy of the selected values array to keep
  // track of values that the user has selected but that
  // are not included in the latest filtered results returned
  // by the API. We'll filter this list as we encounter
  // selected values that _are_ in the API results.
  //
  // We also need to retain the original selection indices
  const selectedButNotInAPIResults = new Map<string, number>();
  selectedValues.map((v, i) => selectedButNotInAPIResults.set(v, i));

  values.forEach((v, i) => {
    const selectedIndex = selectedValues.findIndex(
      (value) => value === v.dimensionValue
    );
    // if we have found this selected value in the API results,
    // remove it from the selectedButNotInAPIResults array
    if (selectedIndex > -1) selectedButNotInAPIResults.delete(v.dimensionValue);

    const cleanValue = cleanUpComparisonValue(v, total, selectedIndex);

    if (i < numberAboveTheFold) {
      aboveTheFold.push(cleanValue);
    } else if (selectedIndex > -1) {
      // Note: if selectedIndex is > -1, it represents the
      // selected value must be included in the below-the-fold list.
      selectedBelowTheFold.push(cleanValue);
    }
  });

  // FIXME: note that it is possible for some values to be selected
  // but not included in the results returned by the API, for example
  // if a dimension value is selected and then a filter is applied
  // that pushes it out of the top N. In that case, we will follow
  // the previous strategy, and just push a dummy value with only
  // the dimension value and nulls for all measure values.
  for (const [dimensionValue, selectedIndex] of selectedButNotInAPIResults) {
    selectedBelowTheFold.push({
      dimensionValue,
      selectedIndex,
      value: null,
      pctOfTotal: null,
      prevValue: null,
      deltaRel: null,
      deltaAbs: null,
    });
  }

  const noAvailableValues = values.length === 0;
  const showExpandTable = values.length > numberAboveTheFold;

  return {
    aboveTheFold,
    selectedBelowTheFold,
    noAvailableValues,
    showExpandTable,
  };
}

/**
 * This returns the "default selection" item labels that
 * will be used when a leaderboard has a comparison active
 * but no items have been directly selected *and included*
 * by the user.
 *
 * Thus, there are three cases:
 * - the leaderboard is in include mode, and there is
 * a selection, we DO NOT return a _default selection_,
 * because the user has made an _explicit selection_.
 *
 * - the leaderboard is in include mode, and there is
 * _no selection_, we return the first three items.
 *
 * - the leaderboard is in exclude mode, we return the
 * first three items that are not selected.
 */
export function getComparisonDefaultSelection(
  values: ComparisonValueWithLabel[],
  selectedValues: (string | number)[],
  excludeMode: boolean
): (string | number)[] {
  if (!excludeMode) {
    if (selectedValues.length > 0) {
      return [];
    }
    return values.slice(0, 3).map((value) => value.dimensionValue);
  }

  return values
    .filter((value) => !selectedValues.includes(value.dimensionValue))
    .map((value) => value.dimensionValue)
    .slice(0, 3);
}

export function getQuerySortType(sortType: SortType) {
  return (
    {
      [SortType.VALUE]:
        ApiSortType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,

      [SortType.DELTA_ABSOLUTE]:
        ApiSortType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,

      [SortType.DELTA_PERCENT]:
        ApiSortType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA,

      // NOTE: sorting by percent-of-total has the same effect
      // as sorting by base value
      [SortType.PERCENT]:
        ApiSortType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,

      // NOTE: UNSPECIFIED is not actually a valid sort type,
      // but it is required by protobuf serialization
      [SortType.UNSPECIFIED]:
        ApiSortType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,

      // FIXME: sort by dimension value is not yet implemented,
      // for now fall back to sorting by base value
      [SortType.DIMENSION]:
        ApiSortType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
    }[sortType] || ApiSortType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE
  );
}
