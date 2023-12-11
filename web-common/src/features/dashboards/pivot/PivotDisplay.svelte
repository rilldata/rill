<script lang="ts">
  import PivotTable from "./PivotTable.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

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
</script>

<div class="layout">
  <PivotSidebar />
  <div class="content">
    <header>
      <PivotHeader />
    </header>
    <main>
      <PivotTable />
    </main>
  </div>
</div>

<style>
  .layout {
    display: flex;
    height: 100vh;
    box-sizing: border-box;
  }

  .content {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  main {
    flex: 1;
    padding: 1rem;
    overflow-y: auto;
  }
</style>
