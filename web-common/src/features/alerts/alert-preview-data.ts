import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import { getLabelForFieldName } from "@rilldata/web-common/features/alerts/utils";
import {
  measureFilterResolutionsStore,
  type ResolvedMeasureFilter,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { getOffset } from "@rilldata/web-common/lib/time/transforms";
import { TimeOffsetType } from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceMetricsViewAggregation,
  queryServiceMetricsViewAggregation,
  type V1Expression,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type {
  CreateQueryOptions,
  CreateQueryResult,
} from "@tanstack/svelte-query";
import { derived } from "svelte/store";

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
      ctx.selectors.timeRangeSelectors.timeControlsState,
      measureFilterResolutionsStore(ctx),
    ],
    (
      [
        runtime,
        metricsViewName,
        dashboard,
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
          timeControls,
          resolvedMeasureFilters,
        ),
        {
          query: getAlertPreviewQueryOptions(
            ctx,
            params,
            timeControls,
            resolvedMeasureFilters,
          ),
        },
      ).subscribe(set);
    },
  );
}

function getAlertPreviewQueryRequest(
  params: AlertPreviewParams,
  dashboard: MetricsExplorerEntity,
  timeControls: TimeControlState,
  resolvedMeasureFilters: ResolvedMeasureFilter,
): V1MetricsViewAggregationRequest {
  return {
    measures: [{ name: params.measure }],
    dimensions: params.dimension ? [{ name: params.dimension }] : [],
    where: sanitiseExpression(
      dashboard.whereFilter,
      resolvedMeasureFilters.filter,
    ),
    having: sanitiseExpression(undefined, params.criteria),
    timeRange: {
      start:
        params.splitByTimeGrain && timeControls.timeEnd
          ? getOffset(
              new Date(timeControls.timeEnd),
              params.splitByTimeGrain,
              TimeOffsetType.SUBTRACT,
            )
          : timeControls.timeStart,
      end: timeControls.timeEnd,
    },
  };
}

function getAlertPreviewQueryOptions(
  ctx: StateManagers,
  params: AlertPreviewParams,
  timeControls: TimeControlState,
  resolvedMeasureFilters: ResolvedMeasureFilter,
): CreateQueryOptions<
  Awaited<ReturnType<typeof queryServiceMetricsViewAggregation>>,
  unknown,
  AlertPreviewResponse
> {
  return {
    enabled:
      !!params.measure && timeControls.ready && resolvedMeasureFilters.ready,
    select: (data) => {
      const rows = data.data as V1MetricsViewAggregationResponseDataItem[];
      const schema = data.schema?.fields?.map((field) => {
        return {
          name: field.name,
          type: field.type?.code,
          label: getLabelForFieldName(ctx, field.name as string),
        };
      }) as VirtualizedTableColumns[];
      return { rows, schema };
    },
    queryClient: ctx.queryClient,
  };
}
