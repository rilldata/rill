<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import PivotEmpty from "./PivotEmpty.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotTable from "./PivotTable.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";
  import { usePivotDataStore } from "./pivot-data-store";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";

  const stateManagers = getStateManagers();

  let showPanels = true;

  $: pivotDataStore = usePivotDataStore(stateManagers);

  $: ({ isFetching, assembled } = $pivotDataStore);

  $: ({ metricsViewName, dashboardStore } = stateManagers);

  function removeActiveCell() {
    if (!$dashboardStore.pivot.activeCell) return;
    metricsExplorerStore.removePivotActiveCell($metricsViewName);
  }
</script>

<div class="layout">
  {#if showPanels}
    <PivotSidebar />
  {/if}
  <div class="content">
    {#if showPanels}
      <PivotHeader />
    {/if}
    <PivotToolbar {isFetching} bind:showPanels {removeActiveCell} />
    <div class="table-view">
      {#if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
        <PivotEmpty {assembled} {isFetching} />
      {:else}
        <PivotTable {pivotDataStore} />
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .layout {
    @apply flex box-border h-full overflow-hidden;
  }

  .content {
    @apply flex w-full flex-col bg-slate-100 overflow-hidden;
  }

  .table-view {
    @apply p-2 pt-0 w-full h-full;
    @apply flex items-start;
    @apply overflow-hidden;
  }
</style>
