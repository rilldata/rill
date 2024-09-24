<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
  import type {
    AvailableTimeGrain,
    DashboardTimeControls,
    TimeGrain,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "../../../runtime-client";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import { getAllowedTimeGrains } from "@rilldata/web-common/lib/time/grains";
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import type { TimeRange } from "@rilldata/web-common/lib/time/types";

  export let exploreName: string;
  export let tdd = false;

  const ctx = getStateManagers();
  const { dashboardStore, validSpecStore } = ctx;
  const timeControlsStore = useTimeControlStore(ctx);

  let timeGrainOptions: TimeGrain[];
  let open = false;

  $: ({ minTimeGrain, timeStart, timeEnd, selectedTimeRange } =
    $timeControlsStore);

  $: timeGrainOptions =
    timeStart && timeEnd
      ? getAllowedTimeGrains(new Date(timeStart), new Date(timeEnd))
      : [];

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  $: activeTimeGrain = $timeControlsStore.selectedTimeRange?.interval;
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
        $dashboardStore?.selectedComparisonTimeRange,
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
    metricsExplorerStore.selectTimeRange(
      exploreName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
      $validSpecStore.data?.metricsView ?? {},
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
          <svelte:element this={tdd ? "b" : "span"}>
            {tdd ? "Time" : "by"}
          </svelte:element>

          <svelte:element this={tdd ? "span" : "b"}>
            {capitalizedLabel}
          </svelte:element>
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
