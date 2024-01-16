<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import PivotTable from "./PivotTable.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";
  import { usePivotDataStore } from "./pivot-data-store";
  import PivotEmpty from "./PivotEmpty.svelte";

  const stateManagers = getStateManagers();

  $: pivotDataStore = usePivotDataStore(stateManagers);
</script>

<div class="layout">
  <PivotSidebar />
  <div class="content">
    <PivotHeader />
    <PivotToolbar />
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

<style>
  .layout {
    display: flex;
    height: 100%;
    box-sizing: border-box;
  }

  .content {
    width: 100%;
    display: flex;
    flex-direction: column;
  }

  .table-view {
    overflow-y: auto;
  }
</style>
