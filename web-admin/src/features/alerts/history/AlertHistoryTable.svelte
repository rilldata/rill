<script lang="ts">
  import AlertHistoryTableCompositeCell from "@rilldata/web-admin/features/alerts/history/AlertHistoryTableCompositeCell.svelte";
  import NoAlertRunsYet from "@rilldata/web-admin/features/alerts/history/NoAlertRunsYet.svelte";
  import { useAlert } from "@rilldata/web-admin/features/alerts/selectors";
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import type { V1AlertExecution } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";

  export let alert: string;

  $: ({ instanceId } = $runtime);

  $: alertQuery = useAlert(instanceId, alert);

  /**
   * Table column definitions.
   * - "composite": Renders all dashboard data in a single cell.
   * - Others: Used for sorting and filtering but not displayed.
   */
  const columns: ColumnDef<V1AlertExecution>[] = [
    {
      id: "composite",
      cell: (info) =>
        flexRender(AlertHistoryTableCompositeCell, {
          alertTime: info.row.original.executionTime,
          timeZone:
            $alertQuery.data.resource.alert.spec.refreshSchedule.timeZone,
          currentExecution:
            $alertQuery.data.resource.alert.state.currentExecution,
          result: info.row.original.result,
        }),
    },
  ];
</script>

<div class="flex flex-col gap-y-4 w-full">
  <div class="flex flex-col gap-y-1">
    <h1 class="text-gray-600 text-lg font-bold">Recent history</h1>
    <p class="text-gray-500 text-sm">Showing up to 25 most recent checks</p>
  </div>
  {#if $alertQuery.error}
    <div class="text-red-500">
      {$alertQuery.error.message}
    </div>
  {:else if $alertQuery.isLoading}
    <div class="text-gray-500">Loading...</div>
  {:else if !$alertQuery.data?.resource.alert.state.executionHistory.length}
    <NoAlertRunsYet />
  {:else}
    <ResourceList
      kind="alert"
      {columns}
      data={$alertQuery.data?.resource.alert.state.executionHistory}
      toolbar={false}
      fixedRowHeight={false}
    />
  {/if}
</div>
