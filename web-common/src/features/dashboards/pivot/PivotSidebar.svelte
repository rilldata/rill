<script lang="ts">
  import DragList from "@rilldata/web-common/features/dashboards/pivot/DragList.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { createQueryServiceMetricsViewRows } from "@rilldata/web-common/runtime-client";

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

  const timeControlsStore = useTimeControlStore(getStateManagers());
  $: tableQuery = createQueryServiceMetricsViewRows(
    $runtime?.instanceId,
    $metricsViewName,
    {
      limit: 1,
      filter: $dashboardStore.filters,
      timeStart: $timeControlsStore.timeStart,
      timeEnd: $timeControlsStore.timeEnd,
    },
    {
      query: {
        enabled: $timeControlsStore.ready && !!$dashboardStore?.filters,
      },
    },
  );

  $: columnsInTable = $dashboardStore?.pivot?.columns;
  $: rowsInTable = $dashboardStore?.pivot?.columns;

  // Todo: Move to external selectors
  $: measures = $visibleMeasures
    .filter((m) => !columnsInTable.includes(m.name as string))
    .map((measure) => ({
      id: measure.name,
      title: measure.label || measure.name,
    }));

  $: dimensions = $visibleDimensions
    .filter((d) => !rowsInTable.includes((d.name ?? d.column) as string))
    .map((dimension) => ({
      id: dimension.name || dimension.column,
      title: dimension.label || dimension.name || dimension.column,
    }));

  $: timeDimesions = (
    $tableQuery?.data?.meta?.filter((d) => d.type === "CODE_TIMESTAMP") ?? []
  ).map((t) => ({
    id: t.name,
    title: t.name,
  }));
</script>

<div class="sidebar">
  <div class="head-text">Drag these over to build your table</div>
  <h2>MEASURES</h2>
  <DragList items={measures} style="vertical" />
  <h2>DIMENSIONS</h2>
  <DragList items={dimensions} style="vertical" />
  <h2>TIME</h2>
  <DragList items={timeDimesions} style="vertical" />
</div>

<style lang="postcss">
  .head-text {
    @apply text-gray-500 text-xs;
  }
  h2 {
    @apply font-semibold text-gray-700 pt-2;
  }
  .sidebar {
    @apply bg-slate-50 p-2;
    min-width: 250px;
    overflow-y: auto;
  }
</style>
