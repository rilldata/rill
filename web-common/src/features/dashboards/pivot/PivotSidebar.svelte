<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import PivotDrag from "./PivotDrag.svelte";
  import { getAllowedTimeGrains } from "@rilldata/web-common/lib/time/grains";
  import { PivotChipType } from "./types";

  const stateManagers = getStateManagers();
  const {
    selectors: {
      pivot: { measures, dimensions, columns, rows },
    },
  } = stateManagers;

  const timeControlsStore = useTimeControlStore(getStateManagers());

  $: allTimeGrains = getAllowedTimeGrains(
    new Date($timeControlsStore.timeStart!),
    new Date($timeControlsStore.timeEnd!),
  ).map((tgo) => {
    return {
      id: tgo.grain,
      title: tgo.label,
      type: PivotChipType.Time,
    };
  });

  $: usedTimeGrains = $columns.dimension
    .filter((m) => m.type === PivotChipType.Time)
    .concat($rows.dimension.filter((d) => d.type === PivotChipType.Time));

  $: timeGrainOptions = allTimeGrains.filter(
    (tgo) => !usedTimeGrains.some((utg) => utg.id === tgo.id),
  );
</script>

<div class="sidebar">
  <PivotDrag title="Time" items={timeGrainOptions} />

  <span class="splitter" />

  <PivotDrag title="Measures" items={$measures} />

  <span class="splitter" />

  <PivotDrag title="Dimensions" items={$dimensions} />
</div>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col items-start;
    @apply h-full min-w-60 w-fit;
    @apply bg-white border-r border-slate-200;
    @apply overflow-hidden;
  }

  .splitter {
    @apply w-full h-[1.5px] bg-gray-200;
  }
</style>
