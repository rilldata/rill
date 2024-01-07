<script lang="ts">
  import PivotTable from "./PivotTable.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { getAxisForDimensions, usePivotDataStore } from "./pivot-data-store";
  import PivotEmpty from "@rilldata/web-common/features/dashboards/pivot/PivotEmpty.svelte";

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      measures: { visibleMeasures },
      dimensions: { dimensionTableColumnName },
      activeMeasure: { activeMeasureName },
    },
    metricsViewName,
    runtime,
  } = stateManagers;

  $: pivotDataStore = usePivotDataStore(stateManagers);

  let pivotDataCopy: unknown[] = [];
  let columnCopy: unknown[] = [];

  $: if ($pivotDataStore?.data) {
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
