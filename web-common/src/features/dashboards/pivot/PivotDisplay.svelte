<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import PivotTable from "./PivotTable.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";
  import { usePivotDataStore } from "./pivot-data-store";
  import type { PivotDataRow } from "./types";
  import PivotEmpty from "./PivotEmpty.svelte";

  const stateManagers = getStateManagers();

  $: pivotDataStore = usePivotDataStore(stateManagers);

  let pivotDataCopy: PivotDataRow[] = [];
  let columnCopy: ColumnDef<PivotDataRow>[] = [];

  $: if ($pivotDataStore?.data && $pivotDataStore.columnDef) {
    pivotDataCopy = $pivotDataStore.data;
    columnCopy = $pivotDataStore.columnDef;
  }
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
        <PivotTable data={$pivotDataStore.data} columns={columnCopy} />
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
