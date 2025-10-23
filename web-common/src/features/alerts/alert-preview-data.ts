import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import {
  type AlertFormValues,
  getAlertQueryArgsFromFormValues,
} from "@rilldata/web-common/features/alerts/form-utils";
import { getComparisonProperties } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
  ComparisonPercentOfTotal,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import type {
  Filters,
  FiltersState,
} from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import type {
  TimeControls,
  TimeControlState,
} from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import {
  createQueryServiceMetricsViewAggregation,
  queryServiceMetricsViewAggregation,
  type StructTypeField,
  TypeCode,
  type V1ExploreSpec,
  type V1MetricsViewAggregationRequest,
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
  filters: Filters,
  timeControls: TimeControls,
): CreateQueryResult<AlertPreviewResponse> {
  return derived(
    [
      useExploreValidSpec(get(runtime).instanceId, formValues.exploreName),
      filters.getStore(),
      timeControls.getStore(),
    ],
    ([validExploreSpec, filtersState, timeControlsState], set) =>
      createQueryServiceMetricsViewAggregation(
        get(runtime).instanceId,
        formValues.metricsViewName,
        getAlertPreviewQueryRequest(
          formValues,
          filtersState,
          timeControlsState,
          validExploreSpec.data?.explore ?? {},
        ),
        {
          query: getAlertPreviewQueryOptions(
            formValues,
            validExploreSpec.data?.metricsView,
          ),
        },
        queryClient,
      ).subscribe(set),
  );
}

function getAlertPreviewQueryRequest(
  formValues: AlertFormValues,
  filtersArgs: FiltersState,
  timeControlArgs: TimeControlState,
  exploreSpec: V1ExploreSpec,
): V1MetricsViewAggregationRequest {
  const req = getAlertQueryArgsFromFormValues(
    formValues,
    filtersArgs,
    timeControlArgs,
    exploreSpec,
  );

  req.limit = "50"; // arbitrary limit to make sure we do not pull too much of data
  if (!timeControlArgs.selectedTimeRange?.end) return req;

  if (req.timeRange && !req.timeRange.expression) {
    req.timeRange.end = timeControlArgs.selectedTimeRange.end.toISOString();
  }
  if (req.comparisonTimeRange && !req.comparisonTimeRange.expression) {
    req.comparisonTimeRange.end =
      timeControlArgs.selectedTimeRange.end.toISOString();
  }
  return req;
}

function getAlertPreviewQueryOptions(
  formValues: AlertFormValues,
  metricsViewSpec: V1MetricsViewSpec | undefined,
): Partial<
  CreateQueryOptions<
    Awaited<ReturnType<typeof queryServiceMetricsViewAggregation>>,
    unknown,
    AlertPreviewResponse
  >
> {
  return {
    enabled: !!formValues.measure && !!metricsViewSpec,
    select: (resp) => {
      return {
        rows: resp.data as V1MetricsViewAggregationResponseDataItem[],
        schema: (resp.schema?.fields
          ?.map((field) => getSchemaEntryForField(metricsViewSpec ?? {}, field))
          .filter(Boolean) ?? []) as VirtualizedTableColumns[],
      };
    },
  };
}

function getSchemaEntryForField(
  metricsViewSpec: V1MetricsViewSpec,
  field: StructTypeField,
): VirtualizedTableColumns | undefined {
  if (metricsViewSpec.dimensions) {
    for (const dimension of metricsViewSpec.dimensions) {
      if (dimension.name === field.name) {
        return {
          name: field.name as string,
          type: field.type?.code ?? TypeCode.CODE_STRING,
          label: dimension.displayName || field.name,
          enableResize: false,
          enableSorting: false,
        };
      }
    }
  }

  if (metricsViewSpec.measures) {
    for (const measure of metricsViewSpec.measures) {
      if (measure.name + ComparisonDeltaPreviousSuffix === field.name)
        return undefined;

      let label: VirtualizedTableColumns["label"] =
        measure.displayName || field.name;
      let format = measure.formatPreset;
      let type: string = field.type?.code ?? TypeCode.CODE_STRING;
      if (
        measure.name + ComparisonDeltaAbsoluteSuffix === field.name ||
        measure.name + ComparisonDeltaRelativeSuffix === field.name ||
        measure.name + ComparisonPercentOfTotal === field.name
      ) {
        const comparisonProps = getComparisonProperties(field.name, measure);
        label = comparisonProps.component;
        format = comparisonProps.format;
        type = comparisonProps.type;
      } else if (measure.name !== field.name) {
        continue;
      }

      return {
        name: field.name as string,
        type,
        label,
        format,
        enableResize: false,
        enableSorting: false,
      };
    }
  }

  return {
    name: field.name as string,
    type: field.type?.code ?? TypeCode.CODE_STRING,
    label: field.name,
    enableResize: false,
    enableSorting: false,
  };
}
