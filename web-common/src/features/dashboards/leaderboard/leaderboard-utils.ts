import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  V1MetricsViewComparisonMeasureType as ApiSortType,
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewComparisonValue,
} from "@rilldata/web-common/runtime-client";
import { SortType } from "../proto-state/derived-types";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";

export type LeaderboardItemData = {
  /**
   *The dimension value label to be shown in the leaderboard
   */
  dimensionValue: string;

  uri: string | null;

  /**
   *  main values to be shown in the leaderboard
   */
  values: (number | null)[];

  /**
   * Percent of total for summable measures; null if not summable.
   * Note that this value will be between 0 and 1, not 0 and 100.
   */
  pctOfTotal: number | null;

  /**
   *  The values from the comparison period.
   * Techinally this might not be a "previous value" but
   * we use that name as a shorthand, since it's the most
   * common use case.
   */
  prevValues: (number | null)[];

  /**
   *  the relative changes from the previous values
   * note that this needs to be multiplied by 100 to get
   * the percentage change
   */
  deltaRels: (number | null)[];

  /**
   *  the absolute changes from the previous values
   */
  deltaAbs: (number | null)[];

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

export const URI_DIMENSION_SUFFIX = "__rill_uri";

const finiteOrNull = (v: unknown): number | null =>
  Number.isFinite(v) ? (v as number) : null;

export function cleanUpComparisonValue(
  v: V1MetricsViewAggregationResponseDataItem,
  dimensionName: string,
  measureNames: string[],
  total: number | null,
  selectedIndex: number,
): LeaderboardItemData {
  const values = measureNames.map((name) => {
    const measureValue = v[name];
    if (!(Number.isFinite(measureValue) || measureValue === null)) {
      console.warn(
        `Leaderboards only implemented for numeric baseValues or missing data (null). Got: ${JSON.stringify(
          v,
        )}`,
      );
    }
    return finiteOrNull(measureValue);
  });

  const prevValues = measureNames.map((name) =>
    finiteOrNull(v[name + ComparisonDeltaPreviousSuffix]),
  );

  const deltaRels = measureNames.map((name) =>
    finiteOrNull(v[name + ComparisonDeltaRelativeSuffix]),
  );

  const deltaAbs = measureNames.map((name) =>
    finiteOrNull(v[name + ComparisonDeltaAbsoluteSuffix]),
  );

  return {
    dimensionValue: v[dimensionName],
    uri: v[dimensionName + URI_DIMENSION_SUFFIX] || null,
    values,
    pctOfTotal: total !== null && values[0] !== null ? values[0] / total : null,
    prevValues,
    deltaRels,
    deltaAbs,
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

function getApiSortName(activeMeasureName: string, sortType: SortType) {
  switch (sortType) {
    case DashboardState_LeaderboardSortType.DELTA_ABSOLUTE:
      return (activeMeasureName += ComparisonDeltaAbsoluteSuffix);

    case DashboardState_LeaderboardSortType.DELTA_PERCENT:
      return (activeMeasureName += ComparisonDeltaRelativeSuffix);

    default:
      return activeMeasureName;
  }
}

export function getSort(
  ascending: boolean,
  type: SortType,
  activeMeasureName: string,
  dimensionName: string,
  timeComparison: boolean,
) {
  return [
    {
      desc: !ascending,
      name:
        type === SortType.DIMENSION || !activeMeasureName
          ? dimensionName
          : timeComparison
            ? getApiSortName(activeMeasureName, type)
            : activeMeasureName || dimensionName,
    },
  ];
}

export function prepareLeaderboardItemData(
  values: V1MetricsViewAggregationResponseDataItem[] | undefined,
  dimensionName: string,
  measureNames: string[],
  numberAboveTheFold: number,
  selectedValues: string[],
  // The total of the measure for the current period,
  // or null if the measure is not valid_percent_of_total
  total: number | null,
) {
  if (values?.length === 0 || !values) {
    return {
      aboveTheFold: [],
      belowTheFoldValues: [],
      noAvailableValues: true,
      showExpandTable: false,
    };
  }

  const aboveTheFold: LeaderboardItemData[] = [];
  const belowTheFoldValues = new Set(selectedValues);

  for (const value of values) {
    if (aboveTheFold.length === numberAboveTheFold) break;

    const dimensionValue = value[dimensionName] as string;

    belowTheFoldValues.delete(dimensionValue);

    const selectedIndex = selectedValues.findIndex((value) =>
      compareLeaderboardValues(value, dimensionValue),
    );

    const cleanValue = cleanUpComparisonValue(
      value,
      dimensionName,
      measureNames,
      total,
      selectedIndex,
    );

    aboveTheFold.push(cleanValue);
  }

  return {
    aboveTheFold,
    belowTheFoldValues: Array.from(belowTheFoldValues),
    noAvailableValues: values.length === 0,
    showExpandTable: values.length > numberAboveTheFold,
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
  excludeMode: boolean,
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

const QuerySortTypeMap: Record<SortType, ApiSortType> = {
  [SortType.VALUE]: ApiSortType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,

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
};
export function getQuerySortType(sortType: SortType) {
  return (
    QuerySortTypeMap[sortType] ||
    ApiSortType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE
  );
}

const QuerySortTypeReverseMap: Record<ApiSortType, SortType> = {} as Record<
  ApiSortType,
  SortType
>;
for (const k in QuerySortTypeMap) {
  QuerySortTypeReverseMap[QuerySortTypeMap[k]] = Number(k);
}
export function getSortType(apiSortType: ApiSortType) {
  return QuerySortTypeReverseMap[apiSortType] || SortType.VALUE;
}

// Backwards compatibility fix for older filters that converted all non-null values to string
export function compareLeaderboardValues(selected: string, value: any) {
  if (selected === null || value === null) {
    return selected === value;
  }
  if (typeof selected === typeof value) {
    return selected === value;
  }
  switch (typeof value) {
    case "boolean":
      return (selected.toLowerCase() === "true") === value;

    case "number":
      return Number(selected) === value;

    default:
      return selected === value;
  }
}

// uri template or "true" string literal or undefined
export function makeHref(
  uriTemplateOrBoolean: string | boolean | null,
  dimensionValue: string,
) {
  if (!uriTemplateOrBoolean) {
    return undefined;
  }

  // temporary fix where uriTemplateOrBoolean is coming in as 0/1 instead of false/true
  if (typeof uriTemplateOrBoolean === "number") {
    uriTemplateOrBoolean = Boolean(uriTemplateOrBoolean);
  }

  // TODO: what should the value be if uriTemplateOrBoolean=false?
  let uri = dimensionValue;
  if (typeof uriTemplateOrBoolean === "string") {
    uri = uriTemplateOrBoolean.replace(/\s/g, "");
  }

  const hasProtocol = /^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(uri);

  if (!hasProtocol) {
    return "https://" + uri;
  } else {
    return uri;
  }
}
