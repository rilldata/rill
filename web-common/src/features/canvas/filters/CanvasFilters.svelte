<script lang="ts">
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import CanvasGrainSelector from "@rilldata/web-common/features/canvas/filters/CanvasGrainSelector.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import CanvasComparisonPill from "./CanvasComparisonPill.svelte";
  import CanvasSuperPill from "./CanvasSuperPill.svelte";

  const { canvasStore } = getCanvasStateManagers();

  $: selectedTimeRange = $canvasStore.timeControls.selectedTimeRange;
  $: selectedComparisonTimeRange =
    $canvasStore.timeControls?.selectedComparisonTimeRange;
  $: activeTimeZone = $canvasStore.timeControls.selectedTimezone;

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(0),
    end: new Date(),
  };
</script>

<div class="flex flex-col gap-y-2 w-full h-20 justify-center ml-2">
  <div class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center">
    <Calendar size="16px" />
    <CanvasSuperPill
      {allTimeRange}
      selectedTimeRange={$selectedTimeRange}
      activeTimeZone={$activeTimeZone}
    />
    <CanvasComparisonPill
      {allTimeRange}
      selectedTimeRange={$selectedTimeRange}
      selectedComparisonTimeRange={$selectedComparisonTimeRange}
    />
    <CanvasGrainSelector
      selectedTimeRange={$selectedTimeRange}
      selectedComparisonTimeRange={$selectedComparisonTimeRange}
    />
  </div>
</div>
