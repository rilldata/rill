<script lang="ts">
  import { getAlertPreviewData } from "@rilldata/web-common/features/alerts/alert-preview-data";
  import TableIcon from "../../components/icons/TableIcon.svelte";
  import PreviewTable from "../../components/preview-table/PreviewTable.svelte";
  import type { V1Expression } from "../../runtime-client";
  import { getStateManagers } from "../dashboards/state-managers/state-managers";

  export let measure: string;
  export let dimension: string;
  export let criteria: V1Expression | undefined = undefined;
  export let splitByTimeGrain: string | undefined = undefined;

  const ctx = getStateManagers();

  $: aggregation = getAlertPreviewData(ctx, {
    measure,
    dimension,
    criteria,
    splitByTimeGrain,
  });

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
