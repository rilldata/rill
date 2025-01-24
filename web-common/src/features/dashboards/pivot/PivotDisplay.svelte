<script lang="ts">
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import PivotEmpty from "./PivotEmpty.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotTable from "./PivotTable.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";
  import { usePivotForExplore } from "./pivot-data-store";

  const stateManagers = getStateManagers();

  let showPanels = true;

  $: pivotDataStore = usePivotForExplore(stateManagers);

  $: ({ isFetching, assembled } = $pivotDataStore);

  $: ({ exploreName, dashboardStore } = stateManagers);

  function removeActiveCell() {
    if (!$dashboardStore.pivot.activeCell) return;
    metricsExplorerStore.removePivotActiveCell($exploreName);
  }
</script>

<div class="layout">
  {#if showPanels}
    <PivotSidebar />
  {/if}
  <div class="flex flex-col size-full overflow-hidden">
    {#if showPanels}
      <PivotHeader />
    {/if}
    <div
      class="content"
      role="presentation"
      on:mousedown|self={removeActiveCell}
    >
      <PivotToolbar {isFetching} bind:showPanels />

      {#if $pivotDataStore?.error?.length}
        <PivotError errors={$pivotDataStore.error} />
      {:else if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
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
    @apply flex w-full flex-col bg-slate-100 overflow-hidden size-full;
    @apply p-2 gap-y-2;
  }
</style>
