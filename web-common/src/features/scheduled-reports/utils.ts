import { splitDimensionsAndMeasuresAsRowsAndColumns } from "@rilldata/web-common/features/dashboards/aggregation-request-utils.ts";
import { getDimensionNameFromAggregationDimension } from "@rilldata/web-common/features/dashboards/aggregation-request/dimension-utils.ts";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import { includeExcludeModeFromFilters } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { ExploreMetricsViewMetadata } from "@rilldata/web-common/features/dashboards/stores/ExploreMetricsViewMetadata.ts";
import { Filters } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import {
  mapV1TimeRangeToSelectedComparisonTimeRange,
  mapV1TimeRangeToSelectedTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils";
import {
  getExistingScheduleFormValues,
  getInitialScheduleFormValues,
} from "@rilldata/web-common/features/scheduled-reports/time-utils";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  V1ExportFormat,
  type V1Expression,
  type V1MetricsViewAggregationRequest,
  type V1Notifier,
  type V1Query,
  type V1ReportSpec,
  type V1TimeRange,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export enum ReportRunAs {
  Recipient = "recipient",
  Creator = "creator",
}
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

export function getNewReportInitialFormValues(
  userEmail: string | undefined,
  aggregationRequest: V1MetricsViewAggregationRequest,
) {
  return {
    title: "",
    webOpenMode: ReportRunAs.Creator as string,
    ...getInitialScheduleFormValues(),
    exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV as V1ExportFormat,
    exportLimit: "",
    exportIncludeHeader: false,
    pdfIncludeFilters: true,
    pdfAllTabs: true,
    ...extractNotification(undefined, userEmail, false),
    ...extractRowsAndColumns(aggregationRequest),
  };
}

export function getNewCanvasReportInitialFormValues(
  userEmail: string | undefined,
) {
  return {
    ...getNewReportInitialFormValues(userEmail, {}),
    exportFormat: V1ExportFormat.EXPORT_FORMAT_PDF as V1ExportFormat,
  };
}

export function getExistingReportInitialFormValues(
  reportSpec: V1ReportSpec,
  userEmail: string | undefined,
  aggregationRequest: V1MetricsViewAggregationRequest,
) {
  const pdfOptions = getPdfOptionsFromWebOpenState(
    reportSpec.annotations?.web_open_state,
  );
  return {
    title: reportSpec.displayName ?? "",
    webOpenMode:
      reportSpec.annotations?.web_open_mode || (ReportRunAs.Creator as string),
    ...getExistingScheduleFormValues(reportSpec.refreshSchedule),
    exportFormat:
      reportSpec?.exportFormat ?? V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED,
    exportLimit: reportSpec.exportLimit === "0" ? "" : reportSpec.exportLimit,
    exportIncludeHeader: reportSpec.exportIncludeHeader ?? false,
    ...pdfOptions,
    ...extractNotification(reportSpec.notifiers, userEmail, true),
    ...extractRowsAndColumns(aggregationRequest),
  };
}

// For canvas PDF reports, the rendering options are stored as extra params
// in the web_open_state annotation (the canvas state query string).
export function getPdfOptionsFromWebOpenState(
  webOpenState: string | undefined,
) {
  const stateParams = new URLSearchParams(webOpenState ?? "");
  return {
    pdfIncludeFilters: stateParams.get("pdf_include_filters") !== "false",
    pdfAllTabs: stateParams.get("pdf_all_tabs") !== "false",
  };
}

// Internal PDF rendering options stored alongside the canvas state in web_open_state.
// They are not part of the canvas state and must be stripped from any params applied
// to a canvas or shown in a user-facing URL.
const INTERNAL_REPORT_PARAMS = ["pdf_include_filters", "pdf_all_tabs"];

export function stripInternalReportParams(
  params: URLSearchParams,
): URLSearchParams {
  for (const param of INTERNAL_REPORT_PARAMS) params.delete(param);
  return params;
}

// Parses a report's "metrics_view_filters" annotation (a JSON object of metrics view
// name to filter expression in protojson format) back into the shape sent on ReportOptions.
export function parseMetricsViewFiltersAnnotation(
  annotation: string | undefined,
): Record<string, V1Expression> | undefined {
  if (!annotation) return undefined;
  try {
    return JSON.parse(annotation) as Record<string, V1Expression>;
  } catch {
    return undefined;
  }
}

export function isCanvasReportSpec(reportSpec: V1ReportSpec): boolean {
  return !!reportSpec.annotations?.canvas;
}

export function getDashboardNameFromReport(reportSpec: V1ReportSpec): string {
  if (reportSpec.annotations?.canvas) return reportSpec.annotations.canvas;

  if (reportSpec.annotations?.explore) return reportSpec.annotations.explore;

  if (reportSpec.annotations?.web_open_path)
    return getExploreName(reportSpec.annotations.web_open_path);

  const queryArgsJson = JSON.parse(
    (reportSpec.resolverProperties?.query_args_json as string | undefined) ||
      reportSpec.queryArgsJson ||
      "{}",
  );

  return (
    queryArgsJson?.metrics_view_name ??
    queryArgsJson?.metricsViewName ??
    queryArgsJson?.metrics_view ??
    queryArgsJson?.metricsView ??
    ""
  );
}

export function getFiltersAndTimeControlsFromAggregationRequest(
  client: RuntimeClient,
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

  const metricsViewMetadata = new ExploreMetricsViewMetadata(
    client,
    metricsViewName,
    exploreName,
  );
  const filters = new Filters(metricsViewMetadata, {
    whereFilter: dimensionFilters,
    dimensionsWithInlistFilter: [],
    dimensionThresholdFilters: dimensionThresholdFilters,
    dimensionFilterExcludeMode: includeExcludeModeFromFilters(dimensionFilters),
  });
  const timeControls = new TimeControls(metricsViewMetadata, {
    selectedTimeRange,
    selectedComparisonTimeRange,
    showTimeComparison: !!selectedComparisonTimeRange,
    selectedTimezone: timeRange?.timeZone ?? "UTC",
  });
  return { filters, timeControls };
}

export function extractRowsAndColumns(
  aggregationRequest: V1MetricsViewAggregationRequest,
) {
  const { rows, dimensionColumns, measureColumns } =
    splitDimensionsAndMeasuresAsRowsAndColumns(aggregationRequest);

  const rowFields = rows.map((dimension) =>
    getDimensionNameFromAggregationDimension(dimension),
  );

  const columnsFromDimensions = dimensionColumns.map((dimension) =>
    getDimensionNameFromAggregationDimension(dimension),
  );

  const columnsFromMeasures = measureColumns.map((measure) => measure.name!);

  return {
    rows: rowFields,
    columns: [...columnsFromDimensions, ...columnsFromMeasures],
  };
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
