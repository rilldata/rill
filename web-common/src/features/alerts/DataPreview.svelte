<script lang="ts">
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import TableIcon from "../../components/icons/TableIcon.svelte";
  import PreviewTable from "../../components/preview-table/PreviewTable.svelte";
  import type { VirtualizedTableColumns } from "../../components/virtualized-table/types";
  import {
    createQueryServiceMetricsViewAggregation,
    V1Expression,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";

  export let metricsView: string;
  export let measure: string;
  export let dimension: string;
  export let criteria: V1Expression | undefined = undefined;

  $: aggregation = createQueryServiceMetricsViewAggregation(
    $runtime.instanceId,
    metricsView,
    {
      measures: [{ name: measure }],
      dimensions: dimension ? [{ name: dimension }] : [],
      having: sanitiseExpression(criteria),
    },
    {
      query: {
        enabled: !!measure,
      },
    },
  );

  let rows: any;
  let tableColumns: VirtualizedTableColumns[];
  $: {
    if ($aggregation.isSuccess) {
      rows = $aggregation.data.data;
      tableColumns = $aggregation.data.schema?.fields?.map((field) => {
        return {
          name: field.name,
          type: field.type?.code,
        };
      }) as VirtualizedTableColumns[];
    }
  }

  $: console.log(criteria);
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
  <PreviewTable {rows} columnNames={tableColumns} />
{/if}
