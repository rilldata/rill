import {
  aggregationRequestWithFilters,
  aggregationRequestWithPivotState,
  aggregationRequestWithTimeRange,
  buildAggregationRequest,
} from "@rilldata/web-common/features/dashboards/aggregation-request-utils.ts";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import type { PivotState } from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import { includeExcludeModeFromFilters } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { ExploreMetricsViewMetadata } from "@rilldata/web-common/features/dashboards/stores/ExploreMetricsViewMetadata.ts";
import {
  Filters,
  type FiltersState,
  getEmptyFilterState,
} from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import {
  getEmptyTimeControlState,
  TimeControls,
  type TimeControlState,
} from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import {
  mapV1TimeRangeToSelectedComparisonTimeRange,
  mapV1TimeRangeToSelectedTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import { PivotStore } from "@rilldata/web-common/features/scheduled-reports/pivot-dashboard/pivot-store.ts";
import type { ReportSource } from "@rilldata/web-common/features/scheduled-reports/report-source/utils.ts";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import type {
  V1MetricsViewAggregationRequest,
  V1TimeRange,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

type ReportPivotRenderData = {
  metadata: ExploreMetricsViewMetadata;
  filters: Filters;
  timeControls: TimeControls;
  pivotStore: PivotStore;
};
export class ReportPivotRenderStore {
  private dataStore = new Map<string, ReportPivotRenderData>();

  public constructor(
    private initFilters: FiltersState = getEmptyFilterState(),
    private initTimeControls: TimeControlState = getEmptyTimeControlState(),
    private initPivotState: PivotState = getEmptyPivotState(),
  ) {}

  public get(
    instanceId: string,
    { metricsViewName, exploreName, canvasName }: ReportSource,
    useInit: boolean,
  ): ReportPivotRenderData {
    const key = `${instanceId}__${metricsViewName}__${exploreName || canvasName}`;
    console.log(key, this.dataStore.has(key));
    if (this.dataStore.has(key)) {
      return this.dataStore.get(key)!;
    }

    const metadata = new ExploreMetricsViewMetadata(
      instanceId,
      metricsViewName,
      exploreName,
    );

    const initFilters = useInit ? this.initFilters : getEmptyFilterState();
    const filters = new Filters(metadata, initFilters);

    const initTimeControls = useInit
      ? this.initTimeControls
      : getEmptyTimeControlState();
    const timeControls = new TimeControls(metadata, initTimeControls);

    const initPivotState = useInit ? this.initPivotState : getEmptyPivotState();
    const pivotStore = new PivotStore(initPivotState);

    const renderData: ReportPivotRenderData = {
      metadata,
      filters,
      timeControls,
      pivotStore,
    };
    this.dataStore.set(key, renderData);
    return renderData;
  }
}

function getEmptyPivotState(): PivotState {
  return {
    rows: [],
    columns: [],
    sorting: [],
    expanded: {},
    columnPage: 1,
    rowPage: 1,
    enableComparison: true,
    activeCell: null,
    tableMode: "flat",
  };
}

export function getAggregationRequestForPivot(
  aggregationRequest: V1MetricsViewAggregationRequest,
  { metadata, filters, timeControls, pivotStore }: ReportPivotRenderData,
) {
  const filtersState = filters.toState();
  const timeControlsState = timeControls.toState();
  const pivotState = get(pivotStore.state);
  const metricsViewSpec = get(metadata.metricsViewSpecQuery).data ?? {};

  const updatedAggregationRequest = buildAggregationRequest(
    aggregationRequest,
    [
      aggregationRequestWithTimeRange({}, timeControlsState),
      aggregationRequestWithFilters(filtersState),
      aggregationRequestWithPivotState(pivotState, metricsViewSpec),
    ],
  );

  return updatedAggregationRequest;
}

export function getInitFiltersFromAggregationRequest(
  aggregationRequest: V1MetricsViewAggregationRequest,
): FiltersState {
  const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
    aggregationRequest.where,
  );

  return {
    whereFilter: dimensionFilters,
    dimensionsWithInlistFilter: [],
    dimensionThresholdFilters: dimensionThresholdFilters,
    dimensionFilterExcludeMode: includeExcludeModeFromFilters(dimensionFilters),
  };
}

export function getInitTimeControlsFromAggregationRequest(
  aggregationRequest: V1MetricsViewAggregationRequest,
  timeRangeSummary: V1TimeRangeSummary | undefined,
): TimeControlState {
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

  return {
    selectedTimeRange,
    selectedComparisonTimeRange,
    showTimeComparison: !!selectedComparisonTimeRange,
    selectedTimezone: timeRange?.timeZone ?? "UTC",
  };
}
