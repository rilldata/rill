<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import PivotDrag from "./PivotDrag.svelte";
  import { getAllowedTimeGrains } from "@rilldata/web-common/lib/time/grains";
  import { PivotChipType } from "./types";
<<<<<<< Updated upstream
=======
  import type { PivotChipData } from "./types";
  import Search from "@rilldata/web-common/components/icons/Search.svelte";

  const CHIP_HEIGHT = 34;
>>>>>>> Stashed changes

  const stateManagers = getStateManagers();

  const {
    selectors: {
      pivot: { measures, dimensions, columns, rows },
    },
  } = stateManagers;

  const timeControlsStore = useTimeControlStore(getStateManagers());

<<<<<<< Updated upstream
=======
  let inputEl: HTMLInputElement;
  let sidebarHeight = 0;
  let searchText = "";

>>>>>>> Stashed changes
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
<<<<<<< Updated upstream
</script>

<div class="sidebar">
  <PivotDrag title="Time" items={timeGrainOptions} />

  <span class="splitter" />

  <PivotDrag title="Measures" items={$measures} />

  <span class="splitter" />

  <PivotDrag title="Dimensions" items={$dimensions} />
=======

  $: filteredMeasures = filterBasedOnSearch($measures, searchText);

  $: filteredDimensions = filterBasedOnSearch($dimensions, searchText);

  $: totalChipSpaces = (sidebarHeight - 150) / CHIP_HEIGHT;

  $: totalChips =
    filteredMeasures.length +
    filteredDimensions.length +
    timeGrainOptions.length;

  $: chipsPerSection = Math.floor(totalChipSpaces / 3);

  $: extraSpace = totalChipSpaces - totalChips > 0;

  function filterBasedOnSearch(fullList: PivotChipData[], search: string) {
    return fullList.filter((d) =>
      d.title.toLowerCase().includes(search.toLowerCase()),
    );
  }
</script>

<div class="sidebar" bind:clientHeight={sidebarHeight}>
  <div
    class="flex w-full items-center p-2 h-fit gap-x-2 border-b border-slate-200"
  >
    <button on:click={() => inputEl.focus()}>
      <Search size="16px" />
    </button>

    <input
      bind:value={searchText}
      bind:this={inputEl}
      type="text"
      placeholder="Search"
      class="w-full h-full"
    />
  </div>

  <PivotDrag
    {extraSpace}
    {chipsPerSection}
    title="Time"
    items={timeGrainOptions}
  />

  <PivotDrag
    {extraSpace}
    {chipsPerSection}
    title="Measures"
    items={filteredMeasures}
  />

  <PivotDrag
    {extraSpace}
    {chipsPerSection}
    title="Dimensions"
    items={filteredDimensions}
  />
>>>>>>> Stashed changes
</div>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col items-start;
    @apply h-full min-w-60 w-fit;
    @apply bg-white border-r border-slate-200;
    @apply overflow-hidden;
  }
</style>
