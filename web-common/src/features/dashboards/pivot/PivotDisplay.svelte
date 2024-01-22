<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import PivotTable from "./PivotTable.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";
  import { usePivotDataStore } from "./pivot-data-store";
  import PivotEmpty from "./PivotEmpty.svelte";

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
    <div class="p-2 px-4">
      <PivotToolbar bind:showPanels />
      <div class="table-view">
        {#if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
          <PivotEmpty />
        {:else}
          <PivotTable
            data={$pivotDataStore.data}
            columns={$pivotDataStore.columnDef}
          />
        {/if}
      </div>
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
