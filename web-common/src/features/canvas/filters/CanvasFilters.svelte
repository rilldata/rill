<script lang="ts">
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import CanvasGrainSelector from "@rilldata/web-common/features/canvas/filters/CanvasGrainSelector.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import CanvasComparisonPill from "./CanvasComparisonPill.svelte";
  import CanvasSuperPill from "./CanvasSuperPill.svelte";

  const ctx = getCanvasStateManagers();
  const { timeControls } = ctx.canvasEntity;

  $: selectedTimeRange = timeControls.selectedTimeRange;
  $: selectedComparisonTimeRange = timeControls?.selectedComparisonTimeRange;
  $: activeTimeZone = timeControls.selectedTimezone;
  $: allTimeRange = timeControls.allTimeRange;
</script>

<div class="flex flex-col gap-y-2 w-full h-20 justify-center ml-2">
  <div class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center">
    <Calendar size="16px" />
    <CanvasSuperPill
      allTimeRange={$allTimeRange}
      selectedTimeRange={$selectedTimeRange}
      activeTimeZone={$activeTimeZone}
    />
    <CanvasComparisonPill
      allTimeRange={$allTimeRange}
      selectedTimeRange={$selectedTimeRange}
      selectedComparisonTimeRange={$selectedComparisonTimeRange}
    />
    <CanvasGrainSelector
      selectedTimeRange={$selectedTimeRange}
      selectedComparisonTimeRange={$selectedComparisonTimeRange}
    />
  </div>
</div>
