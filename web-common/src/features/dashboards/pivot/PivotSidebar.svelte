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

  const CHIP_HEIGHT = 34;

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

<div
  class="sidebar"
  bind:clientHeight={sidebarHeight}
  transition:slide={{ axis: "x" }}
>
  <div class="input-wrapper">
    <Search theme bind:value={searchText} />
  </div>

  <PivotDrag
    title="Time"
    {extraSpace}
    {chipsPerSection}
    items={timeGrainOptions}
    {tableMode}
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
    {tableMode}
    otherChipCounts={[timeGrainOptions.length, filteredDimensions.length]}
  />
</div>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col flex-none relative overflow-hidden;
    @apply h-full border-r z-0 w-60;
    transition-property: width;
    will-change: width;
    @apply select-none bg-surface;
  }

  .input-wrapper {
    @apply flex w-full h-fit items-center;
    @apply border-b;
    @apply gap-x-2 p-2;
  }
</style>
