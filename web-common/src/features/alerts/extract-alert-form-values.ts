import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import {
  getEmptyMeasureFilterEntry,
  mapExprToMeasureFilter,
  type MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  type V1AlertSpec,
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
  | "dimensionsWithInlistFilter"
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
  partialExploreState: Partial<ExploreState>,
): AlertFormValuesSubset {
  if (!queryArgs) return {} as AlertFormValuesSubset;

  const measures = queryArgs.measures as V1MetricsViewAggregationMeasure[];
  const dimensions =
    queryArgs.dimensions as V1MetricsViewAggregationDimension[];

  const timeRange = (queryArgs.timeRange as V1TimeRange) ?? {
    isoDuration: TimeRangePreset.ALL_TIME,
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
    dimensionsWithInlistFilter:
      partialExploreState.dimensionsWithInlistFilter ?? [],
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

export function extractAlertNotification(
  alertSpec: V1AlertSpec,
): AlertNotificationValues {
  const slackNotifier = alertSpec.notifiers?.find(
    (n) => n.connector === "slack",
  );
  const slackChannels = slackNotifier?.properties?.channels as
    | string[]
    | undefined;
  const slackUsers = slackNotifier?.properties?.users as string[] | undefined;

  const emailNotifier = alertSpec.notifiers?.find(
    (n) => n.connector === "email",
  );
  const emailRecipients = emailNotifier?.properties?.recipients as
    | string[]
    | undefined;

  return {
    enableSlackNotification: !!slackNotifier,
    slackChannels: mapAndAddEmptyEntry(slackChannels, "channel"),
    slackUsers: mapAndAddEmptyEntry(slackUsers, "email"),

    enableEmailNotification: !!emailNotifier,
    emailRecipients: mapAndAddEmptyEntry(emailRecipients, "email"),
  };
}

function mapAndAddEmptyEntry<R>(entries: string[] | undefined, key: string): R {
  const mappedEntries = entries?.map((e) => ({ [key]: e })) ?? [];
  mappedEntries.push({ [key]: "" });
  return mappedEntries as R;
}
