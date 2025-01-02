<script lang="ts">
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import {
    getAllowedTimeGrains,
    isGrainBigger,
  } from "@rilldata/web-common/lib/time/grains";
  import type {
    AvailableTimeGrain,
    DashboardTimeControls,
    TimeGrain,
    TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

  export let selectedTimeRange: DashboardTimeControls | undefined;
  export let selectedComparisonTimeRange: DashboardTimeControls | undefined;

  const ctx = getCanvasStateManagers();
  const { canvasStore } = ctx;

  let timeGrainOptions: TimeGrain[];
  // TODO: Change this
  let minTimeGrain = V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  let open = false;

  $: timeGrainOptions =
    selectedTimeRange?.start && selectedTimeRange?.end
      ? getAllowedTimeGrains(
          new Date(selectedTimeRange.start),
          new Date(selectedTimeRange.end),
        )
      : [];

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  $: activeTimeGrain = selectedTimeRange?.interval;
  $: activeTimeGrainLabel =
    activeTimeGrain && TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label;

  $: capitalizedLabel = activeTimeGrainLabel
    ?.split(" ")
    .map((word) => {
      return word.charAt(0).toUpperCase() + word.slice(1);
    })
    .join(" ");

  $: timeGrains = minTimeGrain
    ? timeGrainOptions
        .filter((timeGrain) => !isGrainBigger(minTimeGrain, timeGrain.grain))
        .map((timeGrain) => {
          return {
            main: timeGrain.label,
            key: timeGrain.grain,
          };
        })
    : [];

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        selectedComparisonTimeRange,
      );
    }
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    $canvasStore.timeControls.selectTimeRange(
      timeRange,
      timeGrain,
      comparisonTimeRange,
    );
  }
</script>

{#if activeTimeGrain && timeGrainOptions.length && minTimeGrain}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger asChild let:builder>
      <Chip
        type="time"
        builders={[builder]}
        active={open}
        label="Select a time grain"
      >
        <div slot="body" class="flex gap-x-2 items-center">
          <span>by {capitalizedLabel}</span>
        </div>
      </Chip>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="min-w-40" align="start">
      {#each timeGrains as option (option.key)}
        <DropdownMenu.CheckboxItem
          role="menuitem"
          checked={option.key === activeTimeGrain}
          class="text-xs cursor-pointer"
          on:click={() => onTimeGrainSelect(option.key)}
        >
          {option.main}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
