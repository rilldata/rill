import { getDimensionFilterWithSearch } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
import { calculateEffectiveRowLimit } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import { NUM_ROWS_PER_PAGE } from "@rilldata/web-common/features/dashboards/pivot/pivot-infinite-scroll";
import type {
  V1Expression,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationResponseDataItem,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import {
  getPivotConfigKey,
  getSortFilteredMeasureBody,
  getSortForAccessor,
} from "./pivot-utils";
import type { PivotDataStoreConfig } from "./types";

export interface PivotBaseQueryPlan {
  anchorDimension: string;
  configKey: string;
  displayTotalsRow: boolean;
  effectiveOutermostLimit: number | undefined;
  isMeasureSortAccessor: boolean;
  measureBody: V1MetricsViewAggregationMeasure[];
  rowOffset: number;
  rowPage: number;
  sortAccessor: string | undefined;
  sortFilteredMeasureBody: V1MetricsViewAggregationMeasure[];
  sortPivotBy: V1MetricsViewAggregationSort[];
  timeRange: TimeRangeString;
  whereFilter: V1Expression;
  rowAxisLimitToQuery: string;
}

export interface LimitedRowAxes {
  axesRowTotals: V1MetricsViewAggregationResponseDataItem[];
  hasMoreRows: boolean;
  rowDimensionValues: string[];
}

export function createPivotBaseQueryPlan(
  config: PivotDataStoreConfig,
  columnDimensionAxesData: Record<string, string[]> | undefined,
): PivotBaseQueryPlan {
  const anchorDimension = config.rowDimensionNames[0];
  const measureBody = config.measureNames.map((m) => ({ name: m }));
  const rowPage = config.pivot.rowPage;
  const rowOffset = (rowPage - 1) * NUM_ROWS_PER_PAGE;

  let whereFilter = config.whereFilter;
  if (config.searchText) {
    whereFilter =
      getDimensionFilterWithSearch(
        whereFilter,
        config.searchText,
        anchorDimension,
      ) || config.whereFilter;
  }

  const {
    where: measureWhere,
    sortPivotBy,
    timeRange,
  } = getSortForAccessor(anchorDimension, config, columnDimensionAxesData);

  const { sortFilteredMeasureBody, isMeasureSortAccessor, sortAccessor } =
    getSortFilteredMeasureBody(measureBody, sortPivotBy, measureWhere);

  const effectiveOutermostLimit =
    config.pivot.outermostRowLimit ?? config.pivot.rowLimit;
  const rowAxisLimitToQuery = getOutermostRowAxisLimit(
    config,
    rowOffset,
    effectiveOutermostLimit,
  );

  return {
    anchorDimension,
    configKey: getPivotConfigKey(config),
    displayTotalsRow: Boolean(
      config.rowDimensionNames.length && config.measureNames.length,
    ),
    effectiveOutermostLimit,
    isMeasureSortAccessor,
    measureBody,
    rowOffset,
    rowPage,
    sortAccessor,
    sortFilteredMeasureBody,
    sortPivotBy,
    timeRange,
    whereFilter,
    rowAxisLimitToQuery,
  };
}

function getOutermostRowAxisLimit(
  config: PivotDataStoreConfig,
  rowOffset: number,
  effectiveOutermostLimit: number | undefined,
) {
  const isExplicitOutermostLimit = config.pivot.outermostRowLimit !== undefined;
  const limitToApply = calculateEffectiveRowLimit(
    effectiveOutermostLimit,
    rowOffset,
    NUM_ROWS_PER_PAGE,
    !isExplicitOutermostLimit,
  );

  if (effectiveOutermostLimit === undefined) return limitToApply;
  return (parseInt(limitToApply) + 1).toString();
}

export function applyOutermostRowLimit(
  config: PivotDataStoreConfig,
  plan: PivotBaseQueryPlan,
  rowDimensionValues: string[],
  axesRowTotals: V1MetricsViewAggregationResponseDataItem[],
): LimitedRowAxes {
  if (config.isFlat || plan.effectiveOutermostLimit === undefined) {
    return {
      axesRowTotals,
      hasMoreRows: false,
      rowDimensionValues,
    };
  }

  const isExplicitOutermostLimit = config.pivot.outermostRowLimit !== undefined;
  const limitToApply = calculateEffectiveRowLimit(
    plan.effectiveOutermostLimit,
    plan.rowOffset,
    NUM_ROWS_PER_PAGE,
    !isExplicitOutermostLimit,
  );
  const actualLimit = parseInt(limitToApply);

  if (rowDimensionValues.length <= actualLimit) {
    return {
      axesRowTotals,
      hasMoreRows: false,
      rowDimensionValues,
    };
  }

  return {
    axesRowTotals: axesRowTotals.slice(0, actualLimit),
    hasMoreRows: true,
    rowDimensionValues: rowDimensionValues.slice(0, actualLimit),
  };
}
