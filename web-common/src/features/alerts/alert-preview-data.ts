import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import { getLabelForFieldName } from "@rilldata/web-common/features/alerts/utils";
import {
  measureFilterResolutionsStore,
  type ResolvedMeasureFilter,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { mapDurationToGrain } from "@rilldata/web-common/lib/time/grains";
import {
  createQueryServiceMetricsViewAggregation,
  queryServiceMetricsViewAggregation,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewAggregationSort,
  type V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type {
  CreateQueryOptions,
  CreateQueryResult,
} from "@tanstack/svelte-query";
import { derived, get } from "svelte/store";

export type AlertPreviewParams = {
  measure: string;
  dimension: string;
  criteria: V1Expression | undefined;
  splitByTimeGrain: string | undefined;
};
export type AlertPreviewResponse = {
  rows: V1MetricsViewAggregationResponseDataItem[];
  schema: VirtualizedTableColumns[];
};

export function getAlertPreviewData(
  ctx: StateManagers,
  params: AlertPreviewParams,
): CreateQueryResult<AlertPreviewResponse> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      ctx.dashboardStore,
      useMetricsView(ctx),
      ctx.selectors.timeRangeSelectors.timeControlsState,
      measureFilterResolutionsStore(ctx),
    ],
    (
      [
        runtime,
        metricsViewName,
        dashboard,
        metricsView,
        timeControls,
        resolvedMeasureFilters,
      ],
      set,
    ) => {
      return createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricsViewName,
        getAlertPreviewQueryRequest(
          params,
          dashboard,
          metricsView?.data,
          timeControls,
          resolvedMeasureFilters,
        ),
        {
          query: getAlertPreviewQueryOptions(
            ctx,
            params,
            metricsView?.data ?? {},
          ),
        },
      ).subscribe(set);
    },
  );
}

function getAlertPreviewQueryRequest(
  params: AlertPreviewParams,
  dashboard: MetricsExplorerEntity,
  metricsViewSpec: V1MetricsViewSpec | undefined,
  timeControls: TimeControlState,
  resolvedMeasureFilters: ResolvedMeasureFilter,
): V1MetricsViewAggregationRequest {
  const dimensions: V1MetricsViewAggregationDimension[] = [];
  const sort: V1MetricsViewAggregationSort[] = [];

  const pivotByTime =
    !!params.splitByTimeGrain && !!metricsViewSpec?.timeDimension;
  const grain = mapDurationToGrain(params.splitByTimeGrain ?? "");
  if (pivotByTime && grain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED) {
    dimensions.push({
      name: metricsViewSpec.timeDimension,
      timeZone: dashboard.selectedTimezone,
      timeGrain: grain,
    });
    sort.push({ name: metricsViewSpec.timeDimension, desc: true });
  }
  if (params.dimension) {
    dimensions.push({ name: params.dimension });
    sort.push({ name: params.dimension, desc: true });
  }

  return {
    measures: [{ name: params.measure }],
    dimensions,
    where: sanitiseExpression(
      dashboard.whereFilter,
      resolvedMeasureFilters.filter,
    ),
    having: sanitiseExpression(undefined, params.criteria),
    timeRange: {
      start: timeControls.timeStart,
      end: timeControls.timeEnd,
    },
    limit: "50", // arbitrary limit to make sure we do not pull too much of data
    sort,
  };
}

function getAlertPreviewQueryOptions(
  ctx: StateManagers,
  params: AlertPreviewParams,
  metricsViewSpec: V1MetricsViewSpec,
): CreateQueryOptions<
  Awaited<ReturnType<typeof queryServiceMetricsViewAggregation>>,
  unknown,
  AlertPreviewResponse
> {
  const {
    selectors: {
      measureFilters: { getResolvedFilterForMeasureFilters },
      timeRangeSelectors: { timeControlsState },
    },
  } = ctx;
  const timeControls = get(timeControlsState);
  // 1st get is to fetch the selector, 2nd get is to get async data.
  const resolvedMeasureFilters = get(get(getResolvedFilterForMeasureFilters));
  return {
    enabled:
      !!params.measure &&
      !!metricsViewSpec &&
      timeControls.ready &&
      resolvedMeasureFilters.ready,
    select: (resp) => {
      const rows = resp.data as V1MetricsViewAggregationResponseDataItem[];
      const schema = resp.schema?.fields?.map((field) => {
        return {
          name: field.name,
          type: field.type?.code,
          label: getLabelForFieldName(metricsViewSpec, field.name as string),
        };
      }) as VirtualizedTableColumns[];
      return { rows, schema };
    },
    queryClient: ctx.queryClient,
  };
}
