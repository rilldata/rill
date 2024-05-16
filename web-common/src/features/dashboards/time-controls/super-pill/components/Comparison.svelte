<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { getComparisonRange } from "@rilldata/web-common/lib/time/comparisons";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
  } from "@rilldata/web-common/lib/time/types";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { Interval } from "luxon";
  import { metricsExplorerStore } from "../../../stores/dashboard-stores";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";

  type Option = {
    name: TimeComparisonOption;
    key: number;
    start: Date;
    end: Date;
  };

  export let currentInterval: Interval;
  export let timeComparisonOptionsState: Option[];
  export let onSelectComparisonRange: (
    name: string,
    start: Date,
    end: Date,
  ) => void;
  export let showComparison: boolean | undefined;
  export let selectedComparison: DashboardTimeControls | undefined;
  export let metricViewName: string;

  let open = false;

  $: comparisonOption =
    (selectedComparison?.name as TimeComparisonOption | undefined) || null;
  $: firstOption = timeComparisonOptionsState[0];
  $: label =
    TIME_COMPARISON[comparisonOption ?? firstOption?.name]?.label ??
    "custom period";

  $: selectedLabel = showComparison ? comparisonOption : "custom period";

  // function onSelectCustomComparisonRange(startDate: string, endDate: string) {

  //   onSelectComparisonRange(
  //     TimeComparisonOption.CUSTOM,
  //     new Date(startDate),
  //     new Date(endDate),
  //   );
  // }

  function onCompareRangeSelect(comparisonOption: TimeComparisonOption) {
    if (
      currentInterval.isValid &&
      currentInterval.start &&
      currentInterval.end
    ) {
      const comparisonTimeRange = getComparisonRange(
        currentInterval.start.toJSDate(),
        currentInterval.end.toJSDate(),
        comparisonOption,
      );

      onSelectComparisonRange(
        comparisonOption,
        comparisonTimeRange.start,
        comparisonTimeRange.end,
      );
    }
  }

  function enableTimeComparison() {
    metricsExplorerStore.displayTimeComparison(metricViewName, true);
  }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    {#if showComparison}
      <button use:builder.action {...builder}>
        <div class="gap-x-2 flex" class:inactive={!showComparison}>
          <b>vs</b>

          <p class="line-clamp-1">{label.toLowerCase()}</p>
        </div>
        <span class="flex-none">
          <CaretDownIcon />
        </span>
      </button>
    {:else}
      <button on:click={enableTimeComparison}>
        <div class="gap-x-2 flex" class:inactive={!showComparison}>
          <b>vs</b>

          <p class="line-clamp-1">{label.toLowerCase()}</p>
        </div>
        <span class="flex-none">
          <CaretDownIcon />
        </span>
      </button>
    {/if}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start">
    {#each timeComparisonOptionsState as option (option.name)}
      {@const preset = TIME_COMPARISON[option.name]}
      {@const selected = selectedLabel === option.name}
      <DropdownMenu.Item
        class="flex gap-x-2"
        on:click={() => {
          onCompareRangeSelect(option.name);
        }}
      >
        <span class="w-3 aspect-square">
          {#if selected}
            <Check size="14px" />
          {/if}
        </span>
        <span class:font-bold={selected}>
          {preset?.label || option.name}
        </span>
      </DropdownMenu.Item>
      {#if option.name === TimeComparisonOption.CONTIGUOUS && timeComparisonOptionsState.length > 2}
        <DropdownMenu.Separator />
      {/if}
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply gap-x-1;
  }

  .inactive {
    @apply opacity-50;
  }
</style>
