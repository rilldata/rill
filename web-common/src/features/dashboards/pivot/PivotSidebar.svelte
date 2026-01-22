<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
  import { type TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import type { PivotState } from "web-common/src/features/dashboards/pivot/types.ts";
  import PivotDrag from "./PivotDrag.svelte";
  import { timePillActions, timePillSelectors } from "./time-pill-store";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";

  export let pivotState: PivotState;
  export let measures: PivotChipData[];
  export let dimensions: PivotChipData[];
  export let timeControlsForPillActions: Pick<
    TimeControlState,
    "timeStart" | "timeEnd" | "minTimeGrain"
  >;

  $: ({ rows, columns, tableMode } = pivotState);
  $: splitColumns = splitPivotChips(columns);

  let sidebarHeight = 0;
  let searchText = "";

  onMount(() => {
    timePillActions.initTimeDimension("time", "Time");
  });

  $: if (
    timeControlsForPillActions.timeStart &&
    timeControlsForPillActions.timeEnd
  ) {
    timePillActions.setTimeControls(
      timeControlsForPillActions.timeStart,
      timeControlsForPillActions.timeEnd,
      timeControlsForPillActions.minTimeGrain,
    );
  }

  $: if (rows && columns) {
    timePillActions.updateUsedGrains("time", rows, splitColumns.dimension);
  }

  // Get reactive values from the store
  $: shouldShowTimePill = timePillSelectors.getAllGrainsUsed("time");

  $: timeGrainOptions = !$shouldShowTimePill
    ? [
        {
          id: "time",
          title: "Time",
          type: PivotChipType.Time,
        },
      ]
    : [];

  $: filteredMeasures = filterBasedOnSearch(measures, searchText);
  $: filteredDimensions = filterBasedOnSearch(dimensions, searchText);

  function filterBasedOnSearch(fullList: PivotChipData[], search: string) {
    return fullList.filter((d) =>
      d.title.toLowerCase().includes(search.toLowerCase()),
    );
  }
</script>

<div
  class="sidebar"
  bind:clientHeight={sidebarHeight}
  transition:slide={{ axis: "x" }}
>
  <div class="input-wrapper sticky top-0 z-10">
    <Search theme background bind:value={searchText} />
  </div>

  <PivotDrag title="Time" items={timeGrainOptions} {tableMode} />

  <PivotDrag title="Measures" items={filteredMeasures} />

  <PivotDrag title="Dimensions" items={filteredDimensions} {tableMode} />
</div>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col relative overflow-y-scroll;
    @apply h-full border-r z-0 w-60;
    transition-property: width;
    will-change: width;
    @apply select-none bg-surface-elevated;
  }

  .input-wrapper {
    @apply flex w-full h-fit items-center;
    @apply border-b;
    @apply gap-x-2 p-2;
  }
</style>
