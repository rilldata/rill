import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import { includeExcludeModeFromFilters } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import {
  mapV1TimeRangeToSelectedComparisonTimeRange,
  mapV1TimeRangeToSelectedTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils";
import { Filters } from "@rilldata/web-common/features/scheduled-reports/filters/Filters.ts";
import { MetricsViewData } from "@rilldata/web-common/features/scheduled-reports/filters/MetricsViewData.ts";
import { TimeControls } from "@rilldata/web-common/features/scheduled-reports/filters/TimeControls.ts";
import {
  getDayOfMonthFromCronExpression,
  getDayOfWeekFromCronExpression,
  getFrequencyFromCronExpression,
  getNextQuarterHour,
  getTimeIn24FormatFromDateTime,
  getTimeOfDayFromCronExpression,
  getTodaysDayOfWeek,
  ReportFrequency,
} from "@rilldata/web-common/features/scheduled-reports/time-utils";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  type V1AlertSpec,
  V1ExportFormat,
  type V1MetricsViewAggregationRequest,
  type V1Notifier,
  type V1Query,
  type V1ReportSpec,
  type V1TimeRange,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";

export type ReportValues = ReturnType<typeof getNewReportInitialFormValues>;

export function getQueryNameFromQuery(query: V1Query) {
  if (query.metricsViewAggregationRequest) {
    return "MetricsViewAggregation";
  } else {
    throw new Error(
      "Currently, only `MetricsViewAggregation` queries can be scheduled through the UI",
    );
  }
}

export function getQueryArgsJsonFromQuery(query: V1Query): string {
  if (query.metricsViewAggregationRequest) {
    return JSON.stringify(query.metricsViewAggregationRequest);
  } else {
    throw new Error(
      "Currently, only `MetricsViewAggregation` queries can be scheduled through the UI",
    );
  }
}

export function getNewReportInitialFormValues(userEmail: string | undefined) {
  return {
    title: "",
    frequency: ReportFrequency.Weekly,
    dayOfWeek: getTodaysDayOfWeek(),
    dayOfMonth: 1,
    timeOfDay: getTimeIn24FormatFromDateTime(getNextQuarterHour()),
    timeZone: getLocalIANA(),
    exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV as V1ExportFormat,
    exportLimit: "",
    exportIncludeHeader: false,
    ...extractNotification(undefined, userEmail, false),
  };
}

export function getExistingReportInitialFormValues(
  reportSpec: V1ReportSpec,
  userEmail: string | undefined,
) {
  return {
    title: reportSpec.displayName ?? "",
    frequency: getFrequencyFromCronExpression(
      reportSpec.refreshSchedule?.cron as string,
    ),
    dayOfWeek: getDayOfWeekFromCronExpression(
      reportSpec.refreshSchedule?.cron as string,
    ),
    dayOfMonth: getDayOfMonthFromCronExpression(
      reportSpec.refreshSchedule?.cron as string,
    ),
    timeOfDay: getTimeOfDayFromCronExpression(
      reportSpec.refreshSchedule?.cron as string,
    ),
    timeZone: reportSpec.refreshSchedule?.timeZone ?? getLocalIANA(),
    exportFormat:
      reportSpec?.exportFormat ?? V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED,
    exportLimit: reportSpec.exportLimit === "0" ? "" : reportSpec.exportLimit,
    exportIncludeHeader: reportSpec.exportIncludeHeader ?? false,
    ...extractNotification(reportSpec.notifiers, userEmail, true),
  };
}

export function getDashboardNameFromReport(reportSpec: V1ReportSpec): string {
  if (reportSpec.annotations?.explore) return reportSpec.annotations.explore;

  if (reportSpec.annotations?.web_open_path)
    return getExploreName(reportSpec.annotations.web_open_path);

  const queryArgsJson = JSON.parse(reportSpec.queryArgsJson!);

  return (
    queryArgsJson?.metrics_view_name ??
    queryArgsJson?.metricsViewName ??
    queryArgsJson?.metrics_view ??
    queryArgsJson?.metricsView ??
    ""
  );
}

export function getFiltersFormValuesAggregationRequest(
  instanceId: string,
  metricsViewName: string,
  exploreName: string,
  aggregationRequest: V1MetricsViewAggregationRequest,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  const timeRange = (aggregationRequest.timeRange as V1TimeRange) ?? {
    isoDuration: TimeRangePreset.ALL_TIME,
  };

  let selectedTimeRange: DashboardTimeControls | undefined = undefined;
  let selectedComparisonTimeRange: DashboardTimeControls | undefined =
    undefined;
  if (timeRangeSummary?.max) {
    selectedTimeRange = mapV1TimeRangeToSelectedTimeRange(
      timeRange,
      timeRangeSummary,
      timeRangeSummary.max,
    );
    if (aggregationRequest.comparisonTimeRange) {
      selectedComparisonTimeRange = mapV1TimeRangeToSelectedComparisonTimeRange(
        aggregationRequest.comparisonTimeRange,
        timeRangeSummary,
        timeRangeSummary.max,
      );
    }
  }

  const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
    aggregationRequest.where,
  );

  const metricsViewData = new MetricsViewData(
    instanceId,
    metricsViewName,
    exploreName,
  );
  const filters = new Filters(metricsViewData, {
    whereFilter: dimensionFilters,
    dimensionsWithInlistFilter: [],
    dimensionThresholdFilters: dimensionThresholdFilters,
    dimensionFilterExcludeMode: includeExcludeModeFromFilters(dimensionFilters),
  });
  const timeControls = new TimeControls(metricsViewData, {
    selectedTimeRange,
    selectedComparisonTimeRange,
    showTimeComparison: !!selectedComparisonTimeRange,
    selectedTimezone: timeRange?.timeZone ?? "UTC",
  });
  return { filters, timeControls };
}

function extractNotification(
  notifiers: V1Notifier[] | undefined,
  userEmail: string | undefined,
  isEdit: boolean,
) {
  const slackNotifier = notifiers?.find((n) => n.connector === "slack");
  const slackChannels = mapAndAddEmptyEntry(
    slackNotifier?.properties?.channels as string[],
  );
  const slackUsers = mapAndAddEmptyEntry(
    slackNotifier?.properties?.users as string[],
  );

  const emailNotifier = notifiers?.find((n) => n.connector === "email");
  const emailRecipients = mapAndAddEmptyEntry(
    emailNotifier?.properties?.recipients as string[],
  );

  if (userEmail && !isEdit) {
    slackUsers.unshift(userEmail);
    emailRecipients.unshift(userEmail);
  }

  return {
    enableSlackNotification: isEdit ? !!slackNotifier : false,
    slackChannels,
    slackUsers,

    enableEmailNotification: isEdit ? !!emailNotifier : true,
    emailRecipients,
  };
}

function mapAndAddEmptyEntry(entries: string[] | undefined) {
  const finalEntries = entries ? [...entries] : [];
  finalEntries.push("");
  return finalEntries;
}
