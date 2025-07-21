import {
  getDimensionNameFromAggregationDimension,
  getAggregationDimensionFromTimeDimension,
} from "@rilldata/web-common/features/dashboards/aggregation-request/dimension-utils.ts";
import { MeasureModifierSuffixRegex } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
import {
  mergeDimensionAndMeasureFilters,
  splitWhereFilter,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  ComparisonModifierSuffixRegex,
} from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import { includeExcludeModeFromFilters } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  mapSelectedComparisonTimeRangeToV1TimeRange,
  mapSelectedTimeRangeToV1TimeRange,
  mapV1TimeRangeToSelectedComparisonTimeRange,
  mapV1TimeRangeToSelectedTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils";
import {
  Filters,
  type FiltersState,
} from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import {
  TimeControls,
  type TimeControlState,
} from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import { ExploreMetricsViewMetadata } from "@rilldata/web-common/features/dashboards/stores/ExploreMetricsViewMetadata.ts";
import {
  getExistingScheduleFormValues,
  getInitialScheduleFormValues,
} from "@rilldata/web-common/features/scheduled-reports/time-utils";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  type V1ExploreSpec,
  V1ExportFormat,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewAggregationSort,
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

export function getNewReportInitialFormValues(
  userEmail: string | undefined,
  aggregationRequest: V1MetricsViewAggregationRequest,
) {
  return {
    title: "",
    ...getInitialScheduleFormValues(),
    exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV as V1ExportFormat,
    exportLimit: "",
    exportIncludeHeader: false,
    ...extractNotification(undefined, userEmail, false),
    ...extractRowsAndColumns(aggregationRequest),
  };
}

export function getExistingReportInitialFormValues(
  reportSpec: V1ReportSpec,
  userEmail: string | undefined,
  aggregationRequest: V1MetricsViewAggregationRequest,
) {
  return {
    title: reportSpec.displayName ?? "",
    ...getExistingScheduleFormValues(reportSpec.refreshSchedule),
    exportFormat:
      reportSpec?.exportFormat ?? V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED,
    exportLimit: reportSpec.exportLimit === "0" ? "" : reportSpec.exportLimit,
    exportIncludeHeader: reportSpec.exportIncludeHeader ?? false,
    ...extractNotification(reportSpec.notifiers, userEmail, true),
    ...extractRowsAndColumns(aggregationRequest),
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

export function getFiltersAndTimeControlsFromAggregationRequest(
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

  const metricsViewMetadata = new ExploreMetricsViewMetadata(
    instanceId,
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

export function getUpdatedAggregationRequest(
  aggregationRequest: V1MetricsViewAggregationRequest,
  filtersArgs: FiltersState,
  timeControlArgs: TimeControlState,
  rows: string[],
  columns: string[],
  exploreSpec: V1ExploreSpec,
): V1MetricsViewAggregationRequest {
  const timeRange = mapSelectedTimeRangeToV1TimeRange(
    timeControlArgs.selectedTimeRange,
    timeControlArgs.selectedTimezone,
    exploreSpec,
  );
  const comparisonTimeRange = mapSelectedComparisonTimeRangeToV1TimeRange(
    timeControlArgs.selectedComparisonTimeRange,
    timeControlArgs.showTimeComparison,
    timeRange,
  );

  const allFields = new Set<string>([...rows, ...columns]);
  const isFlat = rows.length === 0;
  const pivotOn: string[] = [];

  const measures = columns
    .filter((col) => exploreSpec.measures?.includes(col))
    .flatMap((measureName) => {
      const group = [{ name: measureName }];

      if (timeControlArgs.showTimeComparison) {
        group.push(
          { name: `${measureName}${COMPARISON_DELTA}` },
          { name: `${measureName}${COMPARISON_PERCENT}` },
        );
      }

      return group;
    });
  const dimensions: V1MetricsViewAggregationDimension[] = rows.map((d) =>
    getAggregationDimensionFromTimeDimension(
      d,
      timeControlArgs.selectedTimezone,
    ),
  );
  columns
    .filter((col) => !exploreSpec.measures?.includes(col))
    .forEach((col) => {
      if (exploreSpec.dimensions?.includes(col)) {
        dimensions.push({ name: col });
        if (!isFlat) pivotOn.push(col);
        return;
      }

      const dimension = getAggregationDimensionFromTimeDimension(
        col,
        timeControlArgs.selectedTimezone,
      );
      dimensions.push(dimension);
      if (!isFlat) pivotOn.push(dimension.alias ?? dimension.name!);
    });

  const sort = getUpdatedAggregationSort(
    aggregationRequest,
    measures,
    dimensions,
    pivotOn,
    allFields,
  );

  return {
    ...aggregationRequest,
    measures,
    dimensions,
    pivotOn: !pivotOn.length ? undefined : pivotOn,
    sort,
    where: sanitiseExpression(
      mergeDimensionAndMeasureFilters(
        filtersArgs.whereFilter,
        filtersArgs.dimensionThresholdFilters,
      ),
      undefined,
    ),
    timeRange,
    comparisonTimeRange,
  };
}

export function extractRowsAndColumns(
  aggregationRequest: V1MetricsViewAggregationRequest,
) {
  const pivotedOn = new Set<string>(aggregationRequest.pivotOn ?? []);
  const rows: string[] = [];
  const columns: string[] = [];
  const isFlat = aggregationRequest.pivotOn === undefined;

  aggregationRequest.dimensions?.forEach((dimension) => {
    const dimensionName = getDimensionNameFromAggregationDimension(dimension);
    if (
      isFlat ||
      pivotedOn.has(dimension.alias!) ||
      pivotedOn.has(dimension.name!)
    ) {
      columns.push(dimensionName);
    } else {
      rows.push(dimensionName);
    }
  });
  aggregationRequest.measures?.forEach((measure) => {
    if (
      MeasureModifierSuffixRegex.test(measure.name!) ||
      ComparisonModifierSuffixRegex.test(measure.name!)
    )
      return;
    columns.push(measure.name!);
  });

  return {
    rows,
    columns,
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

function getUpdatedAggregationSort(
  aggregationRequest: V1MetricsViewAggregationRequest,
  measures: V1MetricsViewAggregationMeasure[],
  dimensions: V1MetricsViewAggregationDimension[],
  pivotOn: string[],
  allFields: Set<string>,
) {
  const hasPivot = pivotOn.length > 0;
  const sort: V1MetricsViewAggregationSort[] =
    aggregationRequest.sort?.filter((s) => {
      if (!allFields.has(s.name!)) return false;
      if (!hasPivot) return true;
      // When there is a pivot we cannot sort by measure or the pivoted dimension
      return (
        !measures.find((m) => m.name === s.name) && !pivotOn.includes(s.name!)
      );
    }) ?? [];
  if (sort.length === 0) {
    let sortField: string | undefined = measures?.[0]?.name;
    let sortFieldIsMeasure = !!sortField;
    if (!sortField || hasPivot) {
      const sortDimension = dimensions.find((d) => !pivotOn.includes(d.alias!));
      sortField = sortDimension?.alias || sortDimension?.name;
      sortFieldIsMeasure = false;
    }
    if (sortField) {
      sort.push({
        desc: sortFieldIsMeasure,
        name: sortField,
      });
    }
  }

  return sort;
}

function mapAndAddEmptyEntry(entries: string[] | undefined) {
  const finalEntries = entries ? [...entries] : [];
  finalEntries.push("");
  return finalEntries;
}
