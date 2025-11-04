<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import { getComparisonLabel } from "@rilldata/web-common/lib/time/comparisons";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
  import { getRangeLabel } from "../time-controls/new-time-controls";
  import RangeDisplay from "../time-controls/super-pill/components/RangeDisplay.svelte";
  import { DateTime, Interval } from "luxon";

  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let hasBoldTimeRange: boolean = true;

  $: selectedLabel = getRangeLabel(
    timeRange.isoDuration ?? timeRange.expression,
  );

  $: showRange =
    selectedLabel === "Custom" ||
    selectedLabel?.startsWith("-") ||
    !isNaN(Number(selectedLabel?.[0]));
</script>

<Chip type="time" readOnly>
  <svelte:fragment slot="body">
    <div class="text-xs text-slate-800 flex gap-x-1.5">
      <div class="font-bold">
        {#if showRange}
          Custom
        {:else}
          {selectedLabel}
        {/if}
      </div>
      {#if showRange && timeRange.start && timeRange.end}
        <RangeDisplay
          interval={Interval.fromDateTimes(
            DateTime.fromISO(timeRange.start).setZone(timeRange.timeZone),
            DateTime.fromISO(timeRange.end).setZone(timeRange.timeZone),
          )}
          timeGrain={timeRange.roundToGrain}
        />
      {/if}
    </div>
  </svelte:fragment>
</Chip>

{#if comparisonTimeRange}
  <Chip type="time" readOnly>
    <svelte:fragment slot="body">
      <div class="text-xs text-slate-800 px-2">
        vs
        <span class:font-bold={hasBoldTimeRange}>
          {getComparisonLabel(comparisonTimeRange)}
        </span>
      </div>
    </svelte:fragment>
  </Chip>
{/if}
