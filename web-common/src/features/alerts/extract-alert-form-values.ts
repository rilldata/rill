import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import {
  getEmptyMeasureFilterEntry,
  mapExprToMeasureFilter,
  type MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  V1Operation,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";

export type AlertFormValuesSubset = Pick<
  AlertFormValues,
  | "metricsViewName"
  | "whereFilter"
  | "dimensionThresholdFilters"
  | "timeRange"
  | "comparisonTimeRange"
  | "measure"
  | "splitByDimension"
  | "criteria"
  | "criteriaOperation"
>;

export function extractAlertFormValues(
  queryArgs: V1MetricsViewAggregationRequest,
  metricsViewSpec: V1MetricsViewSpec,
  allTimeRange: V1MetricsViewTimeRangeResponse,
): AlertFormValuesSubset {
  if (!queryArgs) return {} as AlertFormValuesSubset;

  const measures = queryArgs.measures as V1MetricsViewAggregationMeasure[];
  const dimensions =
    queryArgs.dimensions as V1MetricsViewAggregationDimension[];

  const timeRange = (queryArgs.timeRange as V1TimeRange) ?? {
    isoDuration: metricsViewSpec.defaultTimeRange ?? TimeRangePreset.ALL_TIME,
  };
  if (!timeRange.end && allTimeRange.timeRangeSummary?.max) {
    // alerts only have duration optionally offset, end is added during execution by reconciler
    // so, we add end here to get a valid query
    timeRange.end = allTimeRange.timeRangeSummary?.max;
  }

  const comparisonTimeRange = queryArgs.comparisonTimeRange;
  if (
    comparisonTimeRange &&
    !comparisonTimeRange.end &&
    allTimeRange.timeRangeSummary?.max
  ) {
    // alerts only have duration and offset, end is added during execution by reconciler
    // so, we add end here to get a valid query
    comparisonTimeRange.end = allTimeRange.timeRangeSummary?.max;
  }

  const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
    queryArgs.where,
  );

  return {
    measure: measures[0]?.name ?? "",
    splitByDimension: dimensions[0]?.name ?? "",

    criteria: (queryArgs.having?.cond?.exprs?.map(
      mapExprToMeasureFilter,
    ) as MeasureFilterEntry[]) ?? [getEmptyMeasureFilterEntry()],
    criteriaOperation: queryArgs.having?.cond?.op ?? V1Operation.OPERATION_AND,

    // These are not part of the form, but are used to track the state of the form
    metricsViewName: queryArgs.metricsView as string,
    whereFilter: dimensionFilters,
    dimensionThresholdFilters,
    timeRange,
    comparisonTimeRange,
  };
}

export type AlertNotificationValues = Pick<
  AlertFormValues,
  | "enableSlackNotification"
  | "slackChannels"
  | "slackUsers"
  | "enableEmailNotification"
  | "emailRecipients"
>;
