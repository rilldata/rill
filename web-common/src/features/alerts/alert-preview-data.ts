import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import {
  type AlertFormValues,
  getAlertQueryArgsFromFormValues,
} from "@rilldata/web-common/features/alerts/form-utils";
import { getLabelForFieldName } from "@rilldata/web-common/features/alerts/utils";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  createQueryServiceMetricsViewAggregation,
  createQueryServiceMetricsViewComparison,
  createQueryServiceMetricsViewSchema,
  queryServiceMetricsViewAggregation,
  type V1Expression,
  type V1MetricsViewAggregationDimension,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewAggregationSort,
  type V1MetricsViewComparisonResponse,
  type V1MetricsViewSpec,
  type V1StructType,
  V1TypeCode,
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
        getAlertPreviewQueryRequest(params),
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

function getAlertPreviewQueryRequest(
  params: AlertPreviewParams,
): V1MetricsViewAggregationRequest {
  const dimensions: V1MetricsViewAggregationDimension[] = [];
  const sort: V1MetricsViewAggregationSort[] = [];

  if (params.splitByDimension) {
    dimensions.push({ name: params.splitByDimension });
    sort.push({ name: params.splitByDimension, desc: true });
  }

  return {
    measures: [{ name: params.measure }],
    dimensions,
    where: sanitiseExpression(params.whereFilter, undefined),
    having: sanitiseExpression(undefined, params.criteria),
    timeRange: {
      isoDuration: params.timeRange.isoDuration,
      end: params.timeRange.end,
    },
    limit: "50", // arbitrary limit to make sure we do not pull too much of data
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

export function getAlertCriteriaData(
  formValues: AlertFormValues,
): CreateQueryResult<{
  rows: Record<string, any>[];
  schema: VirtualizedTableColumns[];
}> {
  const alertPreviewQueryParams = getAlertQueryArgsFromFormValues(formValues);
  return derived(
    [
      useMetricsView(get(runtime).instanceId, formValues.metricsViewName),
      createQueryServiceMetricsViewSchema(
        get(runtime).instanceId,
        formValues.metricsViewName,
      ),
    ],
    ([metricsViewResp, metricsViewSchemaResp], set) =>
      createQueryServiceMetricsViewComparison(
        get(runtime).instanceId,
        formValues.metricsViewName,
        {
          ...alertPreviewQueryParams,
          timeRange: alertPreviewQueryParams.timeRange
            ? {
                ...alertPreviewQueryParams.timeRange,
                end: formValues.timeRange.end,
              }
            : undefined,
          comparisonTimeRange: alertPreviewQueryParams.comparisonTimeRange
            ? {
                ...alertPreviewQueryParams.comparisonTimeRange,
                end: formValues.comparisonTimeRange?.end,
              }
            : undefined,
        },
        {
          query: {
            enabled:
              !!alertPreviewQueryParams.having?.cond?.exprs?.length &&
              !!metricsViewResp.data &&
              !!metricsViewSchemaResp.data,
            select: (data) =>
              alertCriteriaDataMapper(
                formValues,
                metricsViewResp.data ?? {},
                metricsViewSchemaResp.data as V1StructType,
                data,
              ),
          },
        },
      ).subscribe(set),
  );
}

function alertCriteriaDataMapper(
  formValues: AlertFormValues,
  metricsView: V1MetricsViewSpec,
  metricsViewSchema: V1StructType,
  data: V1MetricsViewComparisonResponse,
): {
  rows: Record<string, any>[];
  schema: VirtualizedTableColumns[];
} {
  let hasDelta = false;
  let hasDeltaPerc = false;
  const rows =
    data.rows?.map((row) => {
      const retRow: Record<string, any> = {};
      if (formValues.splitByDimension) {
        retRow[formValues.splitByDimension] = row.dimensionValue;
      }
      if (formValues.measure && row.measureValues?.length) {
        const measureValue = row.measureValues[0];
        // TODO: formatting
        retRow[formValues.measure] = measureValue.baseValue;
        if ("deltaAbs" in measureValue) {
          retRow[formValues.measure + "__delta"] = measureValue.deltaAbs;
          hasDelta = true;
        }
        if ("deltaRel" in measureValue) {
          retRow[formValues.measure + "__delta_perc"] = measureValue.deltaRel;
          hasDeltaPerc = true;
        }
      }
      return retRow;
    }) ?? [];

  const schema: VirtualizedTableColumns[] = [];
  if (formValues.splitByDimension) {
    const dim = metricsView.dimensions?.find(
      (d) => d.name === formValues.splitByDimension,
    );
    const col = metricsViewSchema.fields?.find(
      (f) => f.name === formValues.splitByDimension,
    );
    schema.push({
      name: dim?.name ?? formValues.splitByDimension,
      type: (col?.type?.code as string) ?? "VARCHAR",
      label: dim?.label ?? formValues.splitByDimension,
    });
  }

  if (formValues.measure) {
    const mes = metricsView.measures?.find(
      (d) => d.name === formValues.measure,
    );
    const col = metricsViewSchema.fields?.find(
      (f) => f.name === formValues.measure,
    );
    schema.push({
      name: formValues.measure,
      type: col?.type?.code ?? V1TypeCode.CODE_STRING,
      label: mes?.label ?? formValues.measure,
    });
    if (hasDelta) {
      schema.push({
        name: formValues.measure + "__delta",
        type: col?.type?.code ?? V1TypeCode.CODE_FLOAT64,
        label: `Δ`,
      });
    }
    if (hasDeltaPerc) {
      schema.push({
        name: formValues.measure + "__delta_perc",
        type: col?.type?.code ?? V1TypeCode.CODE_FLOAT64,
        label: `Δ%`,
      });
    }
  }

  return { rows, schema };
}
