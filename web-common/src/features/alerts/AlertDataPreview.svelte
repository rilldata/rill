<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
  import TableIcon from "../../components/icons/TableIcon.svelte";
  import PreviewTable from "../../components/preview-table/PreviewTable.svelte";
  import type { V1Expression, V1TimeRange } from "../../runtime-client";

  export let metricsViewName: string;
  export let measure: string;
  export let splitByDimension: string;
  export let splitByTimeGrain: string;
  export let whereFilter: V1Expression;
  export let criteria: V1Expression | undefined = undefined;
  export let timeRange: V1TimeRange;

  const queryClient = useQueryClient();

  $: alertPreviewQuery = getAlertPreviewData(queryClient, {
    metricsViewName,
    measure,
    splitByDimension,
    splitByTimeGrain,
    whereFilter,
    criteria,
    timeRange,
  });
</script>

{#if $alertPreviewQuery.isFetching}
  <div class="p-2 flex flex-col justify-center">
    <Spinner status={EntityStatus.Running} />
  </div>
{:else if !$alertPreviewQuery.data}
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
      rows={$alertPreviewQuery.data.rows}
      columnNames={$alertPreviewQuery.data.schema}
    />
  </div>
{/if}
