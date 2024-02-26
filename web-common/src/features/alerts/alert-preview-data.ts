import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import { getLabelForFieldName } from "@rilldata/web-common/features/alerts/utils";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { mapDurationToGrain } from "@rilldata/web-common/lib/time/grains";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
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
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/query-core";
import type {
  CreateQueryOptions,
  CreateQueryResult,
} from "@tanstack/svelte-query";
import { derived, get } from "svelte/store";

export type AlertPreviewParams = Pick<
  AlertFormValues,
  | "metricsViewName"
  | "whereFilter"
  | "timeRange"
  | "measure"
  | "splitByDimension"
  | "splitByTimeGrain"
> & {
  criteria: V1Expression | undefined;
};
export type AlertPreviewResponse = {
  rows: V1MetricsViewAggregationResponseDataItem[];
  schema: VirtualizedTableColumns[];
};

export function getAlertPreviewData(
  queryClient: QueryClient,
  params: AlertPreviewParams,
): CreateQueryResult<AlertPreviewResponse> {
  return derived(
    [useMetricsView(get(runtime).instanceId, params.metricsViewName)],
    ([metricsViewResp], set) =>
      createQueryServiceMetricsViewAggregation(
        get(runtime).instanceId,
        params.metricsViewName,
        // Add an arbitrary limit to not pull too much of data
        getAlertPreviewQueryRequest(params, metricsViewResp.data ?? {}, 50),
        {
          query: getAlertPreviewQueryOptions(
            queryClient,
            params,
            metricsViewResp.data,
          ),
        },
      ).subscribe(set),
  );
}

export function getAlertPreviewQueryRequest(
  params: AlertPreviewParams,
  metricsViewSpec: V1MetricsViewSpec,
  limit?: number,
): V1MetricsViewAggregationRequest {
  const dimensions: V1MetricsViewAggregationDimension[] = [];
  const sort: V1MetricsViewAggregationSort[] = [];

  const pivotByTime =
    !!params.splitByTimeGrain && !!metricsViewSpec?.timeDimension;
  const grain = mapDurationToGrain(params.splitByTimeGrain ?? "");
  if (pivotByTime && grain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED) {
    dimensions.push({
      name: metricsViewSpec.timeDimension,
      timeZone: "UTC", // TODO
      timeGrain: grain,
    });
    sort.push({ name: metricsViewSpec.timeDimension, desc: true });
  }
  if (params.splitByDimension) {
    dimensions.push({ name: params.splitByDimension });
    sort.push({ name: params.splitByDimension, desc: true });
  }

  let timeRange: V1TimeRange = {
    isoDuration: params.timeRange.isoDuration,
  };
  if (params.timeRange.isoDuration === TimeRangePreset.CUSTOM) {
    timeRange = {
      start: params.timeRange.start,
      end: params.timeRange.end,
    };
  } else if (params.timeRange.isoDuration === TimeRangePreset.DEFAULT) {
    timeRange.isoDuration =
      metricsViewSpec.defaultTimeRange ?? TimeRangePreset.ALL_TIME;
  }

  return {
    metricsView: params.metricsViewName,
    measures: [{ name: params.measure }],
    dimensions,
    where: sanitiseExpression(params.whereFilter, undefined),
    having: sanitiseExpression(undefined, params.criteria),
    timeRange,
    ...(limit ? { limit: limit.toString() } : {}),
    sort,
  };
}

function getAlertPreviewQueryOptions(
  queryClient: QueryClient,
  params: AlertPreviewParams,
  metricsViewSpec: V1MetricsViewSpec | undefined,
): CreateQueryOptions<
  Awaited<ReturnType<typeof queryServiceMetricsViewAggregation>>,
  unknown,
  AlertPreviewResponse
> {
  return {
    enabled: !!params.measure && !!metricsViewSpec,
    select: (resp) => {
      const rows = resp.data as V1MetricsViewAggregationResponseDataItem[];
      const schema = resp.schema?.fields?.map((field) => {
        return {
          name: field.name,
          type: field.type?.code,
          label: getLabelForFieldName(
            metricsViewSpec ?? {},
            field.name as string,
          ),
        };
      }) as VirtualizedTableColumns[];
      return { rows, schema };
    },
    queryClient,
  };
}
