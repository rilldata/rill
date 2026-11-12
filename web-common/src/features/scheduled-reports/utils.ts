import { splitDimensionsAndMeasuresAsRowsAndColumns } from "@rilldata/web-common/features/dashboards/aggregation-request-utils.ts";
import { getDimensionNameFromAggregationDimension } from "@rilldata/web-common/features/dashboards/aggregation-request/dimension-utils.ts";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils";
import {
  getExistingScheduleFormValues,
  getInitialScheduleFormValues,
} from "@rilldata/web-common/features/scheduled-reports/time-utils";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types.ts";
import {
  V1ExportFormat,
  type V1MetricsViewAggregationRequest,
  type V1Notifier,
  type V1Query,
  type V1ReportSpec,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import type { DashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";
import {
  mapV1TimeRangeToComparisonTimeOption,
  mapV1TimeRangeToRillTime,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import { expandCompressedParams } from "@rilldata/web-common/features/dashboards/url-state/compression.ts";
import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";

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

// showEmbedOwnerEmail is true if the form shows the field for the email of the embed user who creates the report.
// See showEmbedOwnerEmail in ScheduledReportDialog.
export function getNewReportInitialFormValues(
  userEmail: string | undefined,
  aggregationRequest: V1MetricsViewAggregationRequest,
  showEmbedOwnerEmail = false,
) {
  return {
    title: "",
    webOpenMode: ReportRunAs.Creator as string,
    ...getInitialScheduleFormValues(),
    exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV as V1ExportFormat,
    exportLimit: "",
    exportIncludeHeader: false,
    // Reports created from an embed are only delivered as an email attachment.
    exportAttach: EmbedStore.isEmbedded(),
    pdfIncludeFilters: true,
    pdfAllTabs: true,
    // The creator's email is in its own field if it's shown, so it isn't also added to the recipients.
    ...extractNotification(
      undefined,
      showEmbedOwnerEmail ? undefined : userEmail,
      false,
    ),
    // In an embed, only the creator's email gets an open link to the embedder's app.
    // The embed's `user_email` is used if set; otherwise the creator can enter it.
    embedOwnerEmail: showEmbedOwnerEmail ? (userEmail ?? "") : "",
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
  showEmbedOwnerEmail = false,
) {
  const pdfOptions = getPdfOptionsFromWebOpenState(
    reportSpec.annotations?.web_open_state,
  );
  // The creator's saved email is in its own field rather than in the recipients,
  // since the admin service always keeps it as a recipient (the creator can't unsubscribe).
  const savedOwnerEmail = reportSpec.annotations?.admin_embed_owner_email ?? "";
  const notification = extractNotification(
    reportSpec.notifiers,
    userEmail,
    true,
  );
  if (savedOwnerEmail) {
    notification.emailRecipients = notification.emailRecipients.filter(
      (r) => r.toLowerCase() !== savedOwnerEmail.toLowerCase(),
    );
  }
  return {
    title: reportSpec.displayName ?? "",
    webOpenMode:
      reportSpec.annotations?.web_open_mode || (ReportRunAs.Creator as string),
    ...getExistingScheduleFormValues(reportSpec.refreshSchedule),
    exportFormat:
      reportSpec?.exportFormat ?? V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED,
    exportLimit: reportSpec.exportLimit === "0" ? "" : reportSpec.exportLimit,
    exportIncludeHeader: reportSpec.exportIncludeHeader ?? false,
    exportAttach: EmbedStore.isEmbedded() || (reportSpec.exportAttach ?? false),
    ...pdfOptions,
    ...notification,
    // In an embed, the embed's `user_email` takes precedence over the saved email, the same way the admin service applies it.
    // Outside of an embed, userEmail is the editor's email, so only the saved email applies.
    embedOwnerEmail: showEmbedOwnerEmail
      ? (EmbedStore.isEmbedded() ? userEmail : "") || savedOwnerEmail
      : "",
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

export function isCanvasReportSpec(reportSpec: V1ReportSpec): boolean {
  return !!reportSpec.annotations?.canvas;
}

// AI reports run the AI resolver and deliver a link to an AI conversation instead of a dashboard export.
export function isAIReportSpec(reportSpec: V1ReportSpec): boolean {
  return reportSpec.resolver === "ai";
}

export function getDashboardNameFromReport(reportSpec: V1ReportSpec): string {
  if (reportSpec.annotations?.canvas) return reportSpec.annotations.canvas;

  if (reportSpec.annotations?.explore) return reportSpec.annotations.explore;

  // AI reports optionally scope the analysis to an explore via the resolver's `explore` property.
  if (isAIReportSpec(reportSpec)) {
    return (reportSpec.resolverProperties?.explore as string | undefined) ?? "";
  }

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

/**
 * Creates the filter managers for the report form, seeded with the report's filter state.
 * They are never synced to the url, so editing filters in the form leaves the dashboard untouched.
 * `singleParamFormMv` matches the dashboard: explore urls carry the singular `f` param,
 * canvas urls carry one param per metrics view.
 */
export function createReportFilterManagers(
  client: RuntimeClient,
  dashboardConfigProvider: Pick<
    DashboardConfigProvider,
    "metricsViewsProvider" | "yamlConfigProvider"
  >,
  singleParamFormMv: boolean,
  initialParams: URLSearchParams,
) {
  const expressionFilterManager = new ExpressionFilterManager(
    dashboardConfigProvider.metricsViewsProvider,
    dashboardConfigProvider.yamlConfigProvider,
    singleParamFormMv,
  );
  // Queued until the metrics view specs load.
  expressionFilterManager.storeSync.setUrlParams(initialParams);

  const timeFilterManager = new TimeFilterManager(
    client,
    dashboardConfigProvider.metricsViewsProvider,
    dashboardConfigProvider.yamlConfigProvider,
  );
  timeFilterManager.storeSync.setUrlParams(initialParams);

  return {
    expressionFilterManager,
    timeFilterManager,
    cleanup: () => {
      expressionFilterManager.cleanup();
      timeFilterManager.cleanup();
    },
  };
}

// Filter managers for an alert's saved aggregation request, scoped to its metrics view.
export function getFiltersAndTimeControlsFromAggregationRequest(
  client: RuntimeClient,
  metricsViewName: string,
  aggregationRequest: V1MetricsViewAggregationRequest,
) {
  const metricsViewsProvider = new MetricsViewsProvider(client, [
    metricsViewName,
  ]);
  const yamlConfigProvider = new YAMLConfigProvider();
  const { expressionFilterManager, timeFilterManager, cleanup } =
    createReportFilterManagers(
      client,
      { metricsViewsProvider, yamlConfigProvider },
      true,
      getFilterParamsFromAggregationRequest(aggregationRequest),
    );

  return {
    expressionFilterManager,
    timeFilterManager,
    cleanup: () => {
      cleanup();
      metricsViewsProvider.cleanup();
      yamlConfigProvider.cleanup?.();
    },
  };
}

// Maps an explore report's aggregation request to the url params the filter managers are seeded with.
export function getFilterParamsFromAggregationRequest(
  aggregationRequest: V1MetricsViewAggregationRequest,
) {
  const params = new URLSearchParams();

  if (aggregationRequest.where) {
    params.set(
      ExploreStateURLParams.Filters,
      convertExpressionToFilterParam(aggregationRequest.where),
    );
  }

  const timeRange = (aggregationRequest.timeRange as V1TimeRange) ?? {
    isoDuration: TimeRangePreset.ALL_TIME,
  };
  const timeRangeParam = mapV1TimeRangeToRillTime(timeRange);
  if (timeRangeParam)
    params.set(ExploreStateURLParams.TimeRange, timeRangeParam);
  if (timeRange.timeZone) {
    params.set(ExploreStateURLParams.TimeZone, timeRange.timeZone);
  }

  if (aggregationRequest.comparisonTimeRange) {
    const comparisonParam = mapV1TimeRangeToComparisonTimeOption(
      timeRange,
      aggregationRequest.comparisonTimeRange,
    );
    if (comparisonParam) {
      params.set(ExploreStateURLParams.ComparisonTimeRange, comparisonParam);
    }
  }

  return params;
}

// Writes the filter managers' state over `baseParams`, keeping params the managers do not own, like `tabs`.
export function applyReportFiltersToParams(
  baseParams: URLSearchParams,
  expressionFilterManager: ExpressionFilterManager,
  timeFilterManager: TimeFilterManager,
) {
  const params = expandCompressedParams(baseParams);
  for (const manager of [expressionFilterManager, timeFilterManager]) {
    const managerParams = new URLSearchParams();
    manager.applyFilterToParams(managerParams);
    manager.paramKeys.forEach((key) => {
      const value = managerParams.get(key);
      if (value) params.set(key, value);
      else params.delete(key);
    });
  }
  return params;
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
    // Reports created from an embed can't be sent to Slack.
    enableSlackNotification:
      isEdit && !EmbedStore.isEmbedded() ? !!slackNotifier : false,
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
