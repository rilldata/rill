<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import PivotDrag from "./PivotDrag.svelte";
  import { getAllowedTimeGrains } from "@rilldata/web-common/lib/time/grains";
  import { PivotChipType } from "./types";
  // import { SearchIcon } from "lucide-svelte";
  import type { PivotChipData } from "./types";
  import Search from "@rilldata/web-common/components/icons/Search.svelte";
  const stateManagers = getStateManagers();
  const {
    selectors: {
      pivot: { measures, dimensions, columns, rows },
    },
  } = stateManagers;

  const timeControlsStore = useTimeControlStore(getStateManagers());

  let inputEl: HTMLInputElement;
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

  function filterBasedOnSearch(fullList: PivotChipData[], search: string) {
    return fullList.filter((d) =>
      d.title.toLowerCase().includes(search.toLowerCase()),
    );
  }
</script>

<div class="sidebar">
  <div
    class="flex w-full items-center p-2 h-[34px] gap-x-2 border-b border-slate-200"
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

  <PivotDrag title="Time" items={timeGrainOptions} />

  <PivotDrag title="Measures" items={filteredMeasures} />

  <PivotDrag title="Dimensions" items={filteredDimensions} />
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
