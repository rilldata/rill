<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import PivotEmpty from "./PivotEmpty.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotTable from "./PivotTable.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";
  import { usePivotDataStore } from "./pivot-data-store";

  const stateManagers = getStateManagers();

  let showPanels = true;

  $: pivotDataStore = usePivotDataStore(stateManagers);
</script>

<div class="layout">
  {#if showPanels}
    <PivotSidebar />
  {/if}
  <div class="content">
    {#if showPanels}
      <PivotHeader />
    {/if}
    <PivotToolbar isFetching={$pivotDataStore.isFetching} bind:showPanels />
    <div class="table-view">
      {#if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
        <PivotEmpty />
      {:else}
        <PivotTable
          pivotStore={pivotDataStore}
          data={$pivotDataStore.data}
          columns={$pivotDataStore.columnDef}
        />
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .layout {
    @apply flex box-border h-full;
  }

  .content {
    @apply flex w-full flex-col;
  }

  .table-view {
    @apply overflow-y-auto;
  }
</style>
