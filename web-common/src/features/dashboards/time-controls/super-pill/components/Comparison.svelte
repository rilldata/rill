<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { getComparisonRange } from "@rilldata/web-common/lib/time/comparisons";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
  } from "@rilldata/web-common/lib/time/types";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { DateTime, Interval } from "luxon";
  import { metricsExplorerStore } from "../../../stores/dashboard-stores";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CalendarPlusDateInput from "./CalendarPlusDateInput.svelte";

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
  export let zone: string;

  $: interval = selectedComparison?.start
    ? Interval.fromDateTimes(selectedComparison.start, selectedComparison.end)
    : null;

  $: firstVisibleMonth = interval?.start ?? DateTime.now();

  let open = false;

  $: comparisonOption =
    (selectedComparison?.name as TimeComparisonOption | undefined) || null;
  $: firstOption = timeComparisonOptionsState[0];
  $: label =
    TIME_COMPARISON[comparisonOption ?? firstOption?.name]?.label ??
    "custom period";

  $: selectedLabel = showComparison ? comparisonOption : "custom period";

  function applyRange(range: Interval<true>) {
    onSelectComparisonRange(
      TimeComparisonOption.CUSTOM,
      range.start.toJSDate(),
      range.end.toJSDate(),
    );
  }

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

<DropdownMenu.Root
  bind:open
  onOpenChange={(open) => {
    if (open && interval?.isValid) {
      firstVisibleMonth = interval.start;
    }
  }}
>
  <DropdownMenu.Trigger asChild let:builder>
    {#if showComparison}
      <button
        use:builder.action
        {...builder}
        aria-label="Select time comparison option"
      >
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

  <DropdownMenu.Content align="start" class="w-72">
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

    <DropdownMenu.Separator />
    {#if interval?.isValid}
      <CalendarPlusDateInput
        {firstVisibleMonth}
        {interval}
        {zone}
        {applyRange}
        closeMenu={() => (open = false)}
      />
    {/if}
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
