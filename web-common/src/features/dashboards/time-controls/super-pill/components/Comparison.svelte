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

  export let currentInterval: Interval<true>;
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
    : currentInterval;

  $: firstVisibleMonth = interval?.start ?? DateTime.now();

  let open = false;
  let showSelector = false;

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
</script>

<DropdownMenu.Root
  bind:open
  closeOnItemClick={false}
  onOpenChange={(open) => {
    if (open && interval && interval?.isValid) {
      firstVisibleMonth = interval.start;
    }
    showSelector = false;
  }}
>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      use:builder.action
      {...builder}
      aria-label="Select time comparison option"
    >
      <div class="gap-x-2 flex" class:inactive={!showComparison}>
        {#if showComparison}
          <b>vs</b>

          <p class="line-clamp-1">{label.toLowerCase()}</p>
        {:else}
          no comparison period
        {/if}
      </div>
      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="p-0 overflow-hidden">
    <div class="flex">
      <div class="flex flex-col border-r w-48 p-1">
        <DropdownMenu.Item
          class="flex gap-x-2"
          on:click={() => {
            metricsExplorerStore.disableAllComparisons(metricViewName);
          }}
        >
          <span class:font-bold={!showComparison}> No comparison period </span>
        </DropdownMenu.Item>

        {#each timeComparisonOptionsState as option (option.name)}
          {@const preset = TIME_COMPARISON[option.name]}
          {@const selected = selectedLabel === option.name}
          <DropdownMenu.Item
            class="flex gap-x-2"
            on:click={() => {
              onCompareRangeSelect(option.name);
              open = false;
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

        <DropdownMenu.Item
          on:click={() => {
            showSelector = true;
          }}
          data-range="custom"
        >
          <span
            class:font-bold={comparisonOption === TimeComparisonOption.CUSTOM &&
              showComparison}
          >
            Custom
          </span>
        </DropdownMenu.Item>
      </div>
      {#if showSelector || (comparisonOption === TimeComparisonOption.CUSTOM && showComparison)}
        <div class="bg-slate-50 flex flex-col w-64 px-2 py-1">
          {#if interval?.isValid}
            <CalendarPlusDateInput
              {firstVisibleMonth}
              {interval}
              {zone}
              {applyRange}
              closeMenu={() => (open = false)}
            />
          {/if}
        </div>
      {/if}
    </div>
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
