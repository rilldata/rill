import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import {
  AlertFormValues,
  getAlertQueryArgsFromFormValues,
} from "@rilldata/web-common/features/alerts/form-utils";
import { getLabelForFieldName } from "@rilldata/web-common/features/alerts/utils";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
import {
  createQueryServiceMetricsViewAggregation,
  queryServiceMetricsViewAggregation,
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/query-core";
import type {
  CreateQueryOptions,
  CreateQueryResult,
} from "@tanstack/svelte-query";
import { derived, get } from "svelte/store";

export type AlertPreviewResponse = {
  rows: V1MetricsViewAggregationResponseDataItem[];
  schema: VirtualizedTableColumns[];
};

export function getAlertPreviewData(
  queryClient: QueryClient,
  formValues: AlertFormValues,
): CreateQueryResult<AlertPreviewResponse> {
  return derived(
    [useMetricsView(get(runtime).instanceId, formValues.metricsViewName)],
    ([metricsViewResp], set) =>
      createQueryServiceMetricsViewAggregation(
        get(runtime).instanceId,
        formValues.metricsViewName,
        getAlertQueryArgsFromFormValues(formValues),
        {
          query: getAlertPreviewQueryOptions(
            queryClient,
            formValues,
            metricsViewResp.data,
          ),
        },
      ).subscribe(set),
  );
}

function getAlertPreviewQueryOptions(
  queryClient: QueryClient,
  formValues: AlertFormValues,
  metricsViewSpec: V1MetricsViewSpec | undefined,
): CreateQueryOptions<
  Awaited<ReturnType<typeof queryServiceMetricsViewAggregation>>,
  unknown,
  AlertPreviewResponse
> {
  return {
    enabled: !!formValues.measure && !!metricsViewSpec,
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
