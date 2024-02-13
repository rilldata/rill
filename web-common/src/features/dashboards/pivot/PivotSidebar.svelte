<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import PivotDrag from "./PivotDrag.svelte";
  import { getAllowedTimeGrains } from "@rilldata/web-common/lib/time/grains";
  import { PivotChipType } from "./types";

  const stateManagers = getStateManagers();
  const {
    selectors: {
      pivot: { measures, dimensions, hasTimePill },
    },
  } = stateManagers;

  const timeControlsStore = useTimeControlStore(getStateManagers());

  $: timeGrainOptions = getAllowedTimeGrains(
    new Date($timeControlsStore.timeStart!),
    new Date($timeControlsStore.timeEnd!),
  ).map((tgo) => {
    return {
      id: tgo.grain,
      title: tgo.label,
      type: PivotChipType.Time,
    };
  });
</script>

<div class="sidebar">
  <div class="container">
    <PivotDrag title="Time" items={timeGrainOptions} disabled={$hasTimePill} />

    <PivotDrag title="Measures" items={$measures} />
    <PivotDrag title="Dimensions" items={$dimensions} />
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
