<script lang="ts">
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import TableIcon from "../../components/icons/TableIcon.svelte";
  import PreviewTable from "../../components/preview-table/PreviewTable.svelte";
  import type { VirtualizedTableColumns } from "../../components/virtualized-table/types";
  import {
    createQueryServiceMetricsViewAggregation,
    V1Expression,
    V1MetricsViewSpec,
    V1TimeRange,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { useMetricsView } from "../dashboards/selectors";
  import { getLabelForFieldName } from "./utils";

  export let metricsView: string;
  export let measure: string;
  export let dimension: string;
  export let filter: V1Expression;
  export let criteria: V1Expression | undefined = undefined;
  export let timeRange: V1TimeRange | undefined = undefined;

  $: metricsViewQuery = useMetricsView($runtime.instanceId, metricsView);
  $: metricsViewSpec = $metricsViewQuery.data;

  $: aggregation = createQueryServiceMetricsViewAggregation(
    $runtime.instanceId,
    metricsView,
    {
      measures: [{ name: measure }],
      dimensions: dimension ? [{ name: dimension }] : [],
      where: sanitiseExpression(filter, undefined),
      having: sanitiseExpression(undefined, criteria),
      timeRange: timeRange,
    },
    {
      query: {
        enabled: !!measure && !!metricsViewSpec,
        select: (data) => {
          const rows = data.data;
          const schema = data.schema?.fields?.map((field) => {
            return {
              name: field.name,
              type: field.type?.code,
              label: getLabelForFieldName(
                metricsViewSpec as V1MetricsViewSpec,
                field.name as string,
              ),
            };
          }) as VirtualizedTableColumns[];
          return { rows, schema };
        },
      },
    },
  );

  // TODO: throttle fetches
</script>

{#if !$aggregation.data}
  <div class="pt-5 pb-10 flex flex-col justify-center items-center gap-1">
    <TableIcon size="32px" className="text-slate-300" />
    <div class="flex flex-col justify-center items-center text-sm">
      <div class="text-gray-600 font-semibold">No data to preview</div>
      <div class="text-gray-500 font-normal">
        To see preview, select measures above.
      </div>
    </div>
  </div>
{:else}
  <div class="max-h-64 overflow-auto">
    <PreviewTable
      rows={$aggregation.data.rows}
      columnNames={$aggregation.data.schema}
    />
  </div>
{/if}
