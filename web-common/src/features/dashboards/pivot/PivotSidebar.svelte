<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import PivotDrag from "./PivotDrag.svelte";
  import { getAllowedTimeGrains } from "@rilldata/web-common/lib/time/grains";
  import { PivotChipType } from "./types";
  import type { PivotChipData } from "./types";
  import Search from "@rilldata/web-common/components/icons/Search.svelte";

  const CHIP_HEIGHT = 34;

  const stateManagers = getStateManagers();
  const {
    selectors: {
      pivot: { measures, dimensions, columns, rows },
    },
  } = stateManagers;

  const timeControlsStore = useTimeControlStore(getStateManagers());

  let inputEl: HTMLInputElement;
  let sidebarHeight = 0;
  let searchText = "";

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

  $: filteredMeasures = filterBasedOnSearch($measures, searchText);
  $: filteredDimensions = filterBasedOnSearch($dimensions, searchText);

  // All of the following reactive statements can be avoided
  // If and when Chrome/Firefox supports max-height with flex-basis
  $: availableChipSpaces = Math.floor((sidebarHeight - 120) / CHIP_HEIGHT);

  $: totalChips =
    filteredMeasures.length +
    filteredDimensions.length +
    timeGrainOptions.length;

  $: chipsPerSection = Math.floor(availableChipSpaces / 3);

  $: extraSpace = availableChipSpaces - totalChips > 0;

  function filterBasedOnSearch(fullList: PivotChipData[], search: string) {
    return fullList.filter((d) =>
      d.title.toLowerCase().includes(search.toLowerCase()),
    );
  }
</script>

<div class="sidebar" bind:clientHeight={sidebarHeight}>
  <div class="input-wrapper">
    <button on:click={() => inputEl.focus()}>
      <Search size="16px" />
    </button>

    <input
      type="text"
      placeholder="Search"
      class="w-full h-full"
      bind:value={searchText}
      bind:this={inputEl}
    />
  </div>

  <PivotDrag
    title="Time"
    {extraSpace}
    {chipsPerSection}
    items={timeGrainOptions}
    otherChipCounts={[filteredDimensions.length, filteredMeasures.length]}
  />

  <PivotDrag
    title="Measures"
    {extraSpace}
    {chipsPerSection}
    items={filteredMeasures}
    otherChipCounts={[timeGrainOptions.length, filteredDimensions.length]}
  />

  <PivotDrag
    title="Dimensions"
    {extraSpace}
    {chipsPerSection}
    items={filteredDimensions}
    otherChipCounts={[timeGrainOptions.length, filteredDimensions.length]}
  />
</div>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col items-start;
    @apply h-full min-w-60 w-fit;
    @apply bg-white border-r border-slate-200;
    @apply overflow-hidden;
  }

  .input-wrapper {
    @apply flex w-full h-fit items-center;
    @apply border-b border-slate-200;
    @apply gap-x-2 p-2;
  }
</style>
