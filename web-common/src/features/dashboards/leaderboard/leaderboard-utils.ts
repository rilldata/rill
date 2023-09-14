import {
  V1MetricsViewComparisonSortType as ApiSortType,
  type V1MetricsViewComparisonRow,
  type V1MetricsViewComparisonValue,
} from "@rilldata/web-common/runtime-client";
import { PERC_DIFF } from "../../../components/data-types/type-utils";
import {
  FormatPreset,
  formatMeasurePercentageDifference,
  formatProperFractionAsPercent,
  humanizeDataType,
} from "../humanize-numbers";
import { LeaderboardContextColumn } from "../leaderboard-context-column";
import { SortType } from "../proto-state/derived-types";
import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";

/**
 * A `V1MetricsViewComparisonRow` basically represents a row of data
 * in the *dimension detail table*, NOT in the leaderboard. Therefore,
 * to convert to rows of leaderboard data, we need to extract a single
 * measure from the dimension table shaped data (namely, the active
 * measure in the leaderboard).
 * @param params
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
    dimensionValue: row.dimensionValue as string | number,
    ...measure,
  };
}

export type LeaderboardItemData = {
  // The dimension value label to be shown in the leaderboard
  dimensionValue: string | number;

  // main value to be shown in the leaderboard
  value: number | null;

  // percent of total for summable measures; null if not summable
  pctOfTotal: number | null;

  // The value from the comparison period.
  // Techinally this might not be a "previous value" but
  // we use that name as a shorthand, since it's the most
  // common use case.
  prevValue: number | null;

  // the relative change from the previous value
  // note that this needs to be multiplied by 100 to get
  // the percentage change
  deltaRel: number | null;

  // the absolute change from the previous value
  deltaAbs: number | null;

  // selection is not enough to determine if the item is included
  // or excluded; for that we need to know the leaderboard's
  // include/exclude state
  selectedIndex: number;
  defaultComparedIndex: number;
};

function cleanUpComparisonValue(
  v: ComparisonValueWithLabel,
  total: number | null,
  selected: boolean
): LeaderboardItemData {
  if (!(Number.isFinite(v.baseValue) || v.baseValue === null)) {
    throw new Error(
      `Leaderboards only implemented for numeric baseValues or missing data (null). Got: ${JSON.stringify(
        v
      )}`
    );
  }
  const value = v.baseValue as number;

  return {
    dimensionValue: v.dimensionValue,
    value,
    pctOfTotal: total && value ? value / total : null,
    prevValue: Number.isFinite(v.comparisonValue)
      ? (v.comparisonValue as number)
      : null,
    deltaRel: Number.isFinite(v.deltaRel) ? (v.deltaRel as number) : null,
    deltaAbs: Number.isFinite(v.deltaAbs) ? (v.deltaAbs as number) : null,

    selected,
  };
}

/**
 * A `V1MetricsViewComparisonValue` augmented with the dimension
 * value that it corresponds to.
 */
type ComparisonValueWithLabel = V1MetricsViewComparisonValue & {
  dimensionValue: string | number;
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
  selectedValues: (string | number)[],
  total: number | null
): {
  aboveTheFold: LeaderboardItemData[];
  selectedBelowTheFold: LeaderboardItemData[];
  noAvailableValues: boolean;
  showExpandTable: boolean;
} {
  const aboveTheFold: LeaderboardItemData[] = [];
  const selectedBelowTheFold: LeaderboardItemData[] = [];
  let selectedValuesCopy = [...selectedValues];
  values.forEach((v, i) => {
    const selected =
      selectedValuesCopy.findIndex((value) => value === v.dimensionValue) >= 0;
    // drop the value from the selectedValues array so that we'll
    // have any left over values that were selected but not included
    // in the results returned by the API
    if (selected)
      selectedValuesCopy = selectedValuesCopy.filter(
        (value) => value !== v.dimensionValue
      );

    if (!excludeMode && count < 3 && !selectedValues.length) {
      defaultComparedIndex = count;
      count = count + 1;
    } else if (excludeMode && count < 3) {
      if (selectedIndex === -1) {
        defaultComparedIndex = count;
        count += 1;
      }
    }

    if (i < numberAboveTheFold) {
      aboveTheFold.push(cleanUpComparisonValue(v, total, selected));
    } else if (selected) {
      selectedBelowTheFold.push(cleanUpComparisonValue(v, total, selected));
    }
  });

  // FIXME: note that it is possible for some values to be selected
  // but not included in the results returned by the API, for example
  // if a dimension value is selected and then a filter is applied
  // that pushes it out of the top N. In that case, we will follow
  // the previous strategy, and just push a dummy value with only
  // the dimension value and nulls for all measure values.
  selectedValuesCopy.forEach((v) => {
    selectedBelowTheFold.push({
      dimensionValue: v,
      selected: true,
      value: null,
      pctOfTotal: null,
      prevValue: null,
      deltaRel: null,
      deltaAbs: null,
    });
  });

  const noAvailableValues = values.length === 0;
  const showExpandTable = values.length > numberAboveTheFold;

  return {
    aboveTheFold,
    selectedBelowTheFold,
    noAvailableValues,
    showExpandTable,
  };
}

export function prepareLeaderboardItemData_dhiraj(
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
 * accounting for the context column type.
 */
export function formatContextColumnValue(
  itemData: LeaderboardItemData,
  contextType: LeaderboardContextColumn,
  formatPreset: FormatPreset
): string | NumberParts | PERC_DIFF.PREV_VALUE_NO_DATA {
  switch (contextType) {
    case LeaderboardContextColumn.DELTA_ABSOLUTE: {
      return humanizeDataType(itemData.deltaAbs, formatPreset);
    }
    case LeaderboardContextColumn.DELTA_PERCENT:
      if (itemData.deltaRel === null || itemData.deltaRel === undefined)
        return PERC_DIFF.PREV_VALUE_NO_DATA;
      return formatMeasurePercentageDifference(itemData.deltaRel);
    case LeaderboardContextColumn.PERCENT:
      return formatProperFractionAsPercent(itemData.pctOfTotal);
    case LeaderboardContextColumn.HIDDEN:
      return "";
    default:
      throw new Error("Invalid context column, all cases must be handled");
  }
}
export const contextColumnWidth = (
  contextType: LeaderboardContextColumn
): string => {
  switch (contextType) {
    case LeaderboardContextColumn.DELTA_ABSOLUTE:
    case LeaderboardContextColumn.DELTA_PERCENT:
      return "56px";
    case LeaderboardContextColumn.PERCENT:
      return "44px";
    case LeaderboardContextColumn.HIDDEN:
      return "0px";
    default:
      throw new Error("Invalid context column, all cases must be handled");
  }
};

export function getQuerySortType(sortType: SortType) {
  return (
    {
      [SortType.VALUE]:
        ApiSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,

      [SortType.DELTA_ABSOLUTE]:
        ApiSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,

      [SortType.DELTA_PERCENT]:
        ApiSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_REL_DELTA,

      // NOTE: sorting by percent-of-total has the same effect
      // as sorting by base value
      [SortType.PERCENT]:
        ApiSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,

      // NOTE: UNSPECIFIED is not actually a valid sort type,
      // but it is required by protobuf serialization
      [SortType.UNSPECIFIED]:
        ApiSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,

      // FIXME: sort by dimension value is not yet implemented,
      // for now fall back to sorting by base value
      [SortType.DIMENSION]:
        ApiSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,
    }[sortType] || ApiSortType.METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE
  );
}
