<script lang="ts">
  import DragList from "@rilldata/web-common/features/dashboards/pivot/DragList.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  const stateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      measures: { visibleMeasures },
      dimensions: { visibleDimensions },
    },
    metricsViewName,
    runtime,
  } = stateManagers;

  $: measures = $visibleMeasures.map((measure) => ({
    id: measure.name,
    title: measure.label || measure.name,
  }));

  $: dimensions = $visibleDimensions.map((dimension) => ({
    id: dimension.column || dimension.name,
    title: dimension.label || dimension.name || dimension.column,
  }));
</script>

<div class="sidebar">
  <div class="head-text">Drag these over to build your table</div>
  <h2>MEASURES</h2>
  <DragList items={measures} style="vertical" />
  <h2>DIMENSIONS</h2>
  <DragList items={dimensions} style="vertical" />
</div>

<style lang="postcss">
  .head-text {
    @apply text-gray-500 text-xs;
  }
  h2 {
    @apply font-semibold text-gray-700 pt-2;
  }
  .sidebar {
    @apply bg-slate-50;
    width: 200px;
    padding: 10px;
    height: calc(100% - 50px);
    flex: 0 0 200px;
    padding: 1rem;
    overflow-y: auto;
  }
</style>
