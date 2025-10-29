import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { URI_DIMENSION_SUFFIX } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  QueryServiceMetricsViewAggregationBody,
  V1Expression,
  V1MetricsViewAggregationMeasure,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { SortType } from "./proto-state/derived-types";
import type { TimeControlState } from "./time-controls/time-control-store";

const countRegex = /count(?=[^(]*\()/i;
const sumRegex = /sum(?=[^(]*\()/i;

export function isSummableMeasure(measure: MetricsViewSpecMeasure): boolean {
  if (measure.validPercentOfTotal) return true;

  const expression = measure.expression?.toLowerCase();

  // Check if expression contains SUM or COUNT
  const hasSumOrCount = !!(
    expression?.match(countRegex) || expression?.match(sumRegex)
  );

  if (hasSumOrCount) {
    // If it has SUM/COUNT and contains division, verify both sides are summable
    if (expression?.includes("/")) {
      // Check if this is a ratio expression: sum/sum, count/sum, sum/count, count/count
      // Split by division and check if both parts contain sum or count
      const parts = expression.split("/").map((p) => p.trim());
      if (parts.length === 2) {
        const leftHasSumOrCount = !!(
          parts[0].match(countRegex) || parts[0].match(sumRegex)
        );
        const rightHasSumOrCount = !!(
          parts[1].match(countRegex) || parts[1].match(sumRegex)
        );

        return leftHasSumOrCount && !rightHasSumOrCount;
      }
    }
    return true;
  }
  return false;
}

export function isPercentageMeasure(measure: MetricsViewSpecMeasure): boolean {
  return (
    measure.formatPreset === FormatPreset.PERCENTAGE ||
    !!measure.formatD3?.endsWith("%")
  );
}

/**
 * Returns a sanitized column name appropriate for use in e.g. filters.
 *
 * Even though this is a one-liner, we externalize it as a function
 * becuase it is used in a few places and we want to make sure we
 * are consistent in how we handle this.
 */
export function getDimensionColumn(
  dimension: MetricsViewSpecDimension,
): string {
  return (dimension?.column || dimension?.name) as string;
}

export function prepareSortedQueryBody(
  dimensionName: string,
  measureNames: string[],
  timeControls: TimeControlState,
  // Note: sortMeasureName may be null if we are sorting by dimension values
  sortMeasureName: string | null,
  sortType: SortType,
  sortAscending: boolean,
  whereFilterForDimension: V1Expression,
  limit: number,
): QueryServiceMetricsViewAggregationBody {
  const measures = measureNames.map(
    (n) =>
      <V1MetricsViewAggregationMeasure>{
        name: n,
      },
  );

  let apiSortName = sortMeasureName;
  if (sortType === SortType.DIMENSION || sortMeasureName === null) {
    apiSortName = dimensionName;
  }

  if (
    timeControls.showTimeComparison &&
    !!timeControls.selectedComparisonTimeRange &&
    sortMeasureName
  ) {
    // insert beside the correct measure
    measures.splice(
      measures.findIndex((m) => m.name === sortMeasureName),
      0,
      ...getComparisonRequestMeasures(sortMeasureName),
    );
    if (apiSortName === sortMeasureName) {
      // only update if the sort was on measure
      switch (sortType) {
        case DashboardState_LeaderboardSortType.DELTA_ABSOLUTE:
          apiSortName += ComparisonDeltaAbsoluteSuffix;
          break;
        case DashboardState_LeaderboardSortType.DELTA_PERCENT:
          apiSortName += ComparisonDeltaRelativeSuffix;
          break;
      }
    }
  }

  return {
    dimensions: [
      {
        name: dimensionName,
      },
    ],
    measures,
    timeRange: {
      start: timeControls.timeStart,
      end: timeControls.timeEnd,
    },
    ...(timeControls.selectedComparisonTimeRange &&
    timeControls.showTimeComparison
      ? {
          comparisonTimeRange: {
            start: timeControls.comparisonTimeStart,
            end: timeControls.comparisonTimeEnd,
          },
        }
      : {}),
    sort: apiSortName
      ? [
          {
            desc: !sortAscending,
            name: apiSortName,
          },
        ]
      : [],
    where: sanitiseExpression(whereFilterForDimension, undefined),
    limit: limit.toString(),
    offset: "0",
  };
}

/**
 * Gets comparison based measures used in MetricsViewAggregationRequest
 */
export function getComparisonRequestMeasures(
  measureName: string,
): V1MetricsViewAggregationMeasure[] {
  return [
    {
      name: measureName + ComparisonDeltaPreviousSuffix,
      comparisonValue: {
        measure: measureName,
      },
    },
    {
      name: measureName + ComparisonDeltaAbsoluteSuffix,
      comparisonDelta: {
        measure: measureName,
      },
    },
    {
      name: measureName + ComparisonDeltaRelativeSuffix,
      comparisonRatio: {
        measure: measureName,
      },
    },
  ];
}

export function getURIRequestMeasure(
  dimensionName: string,
): V1MetricsViewAggregationMeasure {
  return {
    name: dimensionName + URI_DIMENSION_SUFFIX,
    uri: {
      dimension: dimensionName,
    },
  };
}

export function getBreadcrumbOptions(
  exploreResources: V1Resource[],
  canvasResources: V1Resource[],
): Map<string, PathOption> {
  const exploreOptions = exploreResources.reduce((map, exploreResource) => {
    const name = exploreResource.meta?.name?.name ?? "";
    const label =
      exploreResource.explore?.state?.validSpec?.displayName || name;

    if (label && name)
      map.set(name.toLowerCase(), { label, section: "explore", depth: 0 });

    return map;
  }, new Map<string, PathOption>());

  const canvasOptions = canvasResources.reduce((map, canvasResource) => {
    const name = canvasResource.meta?.name?.name ?? "";
    const label = canvasResource?.canvas?.spec?.displayName || name;

    if (label && name)
      map.set(name.toLowerCase(), { label, section: "canvas", depth: 0 });

    return map;
  }, new Map<string, PathOption>());

  return new Map([...exploreOptions, ...canvasOptions]);
}
