<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { createQueryServiceMetricsViewRows } from "@rilldata/web-common/runtime-client";
  import PivotDrag from "./PivotDrag.svelte";
  import { PivotChipType } from "./types";

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
      where: sanitiseExpression($dashboardStore.whereFilter, undefined),
      timeStart: $timeControlsStore.timeStart,
      timeEnd: $timeControlsStore.timeEnd,
    },
    {
      query: {
        enabled: $timeControlsStore.ready && !!$dashboardStore?.whereFilter,
      },
    },
  );

  $: columnsInTable = $dashboardStore?.pivot?.columns;
  $: rowsInTable = $dashboardStore?.pivot?.columns;

  // Todo: Move to external selectors
  $: measures = $visibleMeasures
    .filter((m) => !columnsInTable.includes(m.name as string))
    .map((measure) => ({
      id: measure.name || "Unknown",
      title: measure.label || measure.name || "Unknown",
      type: PivotChipType.Measure,
    }));

  $: dimensions = $visibleDimensions
    .filter((d) => !rowsInTable.includes((d.name ?? d.column) as string))
    .map((dimension) => ({
      id: dimension.name || dimension.column || "Unknown",
      title: dimension.label || dimension.name || dimension.column || "Unknown",
      type: PivotChipType.Dimension,
    }));

  $: timeDimesions = (
    $tableQuery?.data?.meta?.filter((d) => d.type === "CODE_TIMESTAMP") ?? []
  ).map((t) => ({
    id: t.name ?? "Unknown",
    title: t.name ?? "Unknown",
    type: PivotChipType.Time,
  }));
</script>

<div class="sidebar">
  <div class="container">
    <PivotDrag title="Time" items={timeDimesions} />
    <PivotDrag title="Measures" items={measures} />
    <PivotDrag title="Dimensions" items={dimensions} />
  </div>
</div>

<style lang="postcss">
  .sidebar {
    @apply h-full min-w-fit py-2 p-4;
    @apply overflow-y-auto;
    @apply bg-white border-r border-slate-200;
  }

  .container {
    @apply flex flex-col gap-y-4;
    @apply min-w-[120px];
  }
</style>
