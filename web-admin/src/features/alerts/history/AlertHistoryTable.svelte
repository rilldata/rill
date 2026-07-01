<script lang="ts">
  import AlertHistoryTableCompositeCell from "@rilldata/web-admin/features/alerts/history/AlertHistoryTableCompositeCell.svelte";
  import NoAlertRunsYet from "@rilldata/web-admin/features/alerts/history/NoAlertRunsYet.svelte";
  import { useAlert } from "@rilldata/web-admin/features/alerts/selectors";
  import ResourceList from "@rilldata/web-common/features/resources/ResourceList.svelte";
  import type { V1AlertExecution } from "@rilldata/web-common/runtime-client/gen/index.schemas";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let alert: string;

  const runtimeClient = useRuntimeClient();

  $: alertQuery = useAlert(runtimeClient, alert);

  /**
   * Table column definitions.
   * - "composite": Renders all dashboard data in a single cell.
   * - Others: Used for sorting and filtering but not displayed.
   */
  const columns: ColumnDef<V1AlertExecution>[] = [
    {
      id: "composite",
      cell: (info) =>
        renderComponent(AlertHistoryTableCompositeCell, {
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
    <h1 class="text-fg-secondary text-lg font-bold">{m.alert_recent_history()}</h1>
    <p class="text-fg-secondary text-sm">{m.alert_showing_recent_checks()}</p>
  </div>
  {#if $alertQuery.error}
    <div class="text-red-500">
      {$alertQuery.error.message}
    </div>
  {:else if $alertQuery.isLoading}
    <div class="text-fg-secondary">{m.alert_loading()}</div>
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
