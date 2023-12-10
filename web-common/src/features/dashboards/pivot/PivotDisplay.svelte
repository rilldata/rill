<script lang="ts">
  import TDDHeader from "@rilldata/web-common/features/dashboards/time-dimension-details/TDDHeader.svelte";
  import PivotTable from "./PivotTable.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotHeader from "./PivotHeader.svelte";

  export let metricViewName;
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
