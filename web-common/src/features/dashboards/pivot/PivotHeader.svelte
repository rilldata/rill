<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import DragList from "./DragList.svelte";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      measures: { measureLabel },
      dimensions: { getDimensionDisplayName },
    },
    metricsViewName,
    runtime,
  } = stateManagers;

  $: colMeasures = $dashboardStore?.pivot?.columns?.map((col) => ({
    id: col,
    title: $measureLabel(col),
  }));

  $: rowDimensions = $dashboardStore?.pivot?.rows?.map((row) => ({
    id: row,
    title: $getDimensionDisplayName(row),
  }));
</script>

<div class="header">
  <div class="header-row">
    <span class="row-label"> <Column size="16px" /> Columns</span>
    <DragList
      on:update={(e) =>
        metricsExplorerStore.setPivotColumns(
          $metricsViewName,
          e.detail?.map((item) => item.id),
        )}
      items={colMeasures}
      style="horizontal"
    />
  </div>
  <div class="header-row">
    <span class="row-label"> <Row size="16px" /> Rows</span>

    <DragList
      on:update={(e) =>
        metricsExplorerStore.setPivotRows(
          $metricsViewName,
          e.detail?.map((item) => item.id),
        )}
      items={rowDimensions}
      style="horizontal"
    />
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex flex-col;
    border-bottom: 1px solid #ddd;
  }
  .header-row {
    @apply flex items-center gap-x-2 px-2 py-1;
  }
  .row-label {
    @apply flex items-center gap-x-1 w-20;
  }
</style>
