<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import ClockCircle from "@rilldata/web-common/components/icons/ClockCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { getComparisonRange } from "@rilldata/web-common/lib/time/comparisons";
  import {
    NO_COMPARISON_LABEL,
    TIME_COMPARISON,
  } from "@rilldata/web-common/lib/time/config";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
  } from "@rilldata/web-common/lib/time/types";
  import { createEventDispatcher } from "svelte";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { Interval } from "luxon";

  export let currentInterval: Interval;
  export let timeComparisonOptionsState: {
    name: TimeComparisonOption;
    key: number;
    start: Date;
    end: Date;
  }[];
  export let onSelectComparisonRange;
  export let disableAllComparisons: () => void;
  export let showComparison: boolean | undefined;
  export let selectedComparison: DashboardTimeControls | undefined;

  // const {
  //   selectors: {
  //     timeRangeSelectors: { timeComparisonOptionsState },
  //   },
  // } = getStateManagers();

  let open = false;

  $: comparisonOption = selectedComparison?.name;

  $: label =
    showComparison && comparisonOption
      ? TIME_COMPARISON[comparisonOption]?.label
      : NO_COMPARISON_LABEL;

  $: selectedLabel = showComparison ? comparisonOption : NO_COMPARISON_LABEL;

  function onSelectCustomComparisonRange(startDate: string, endDate: string) {
    // intermediateSelection = TimeComparisonOption.CUSTOM;

    // dispatch("select-comparison", {
    //   name: TimeComparisonOption.CUSTOM,
    //   start: new Date(startDate),
    //   end: new Date(endDate),
    // });

    onSelectComparisonRange(
      TimeComparisonOption.CUSTOM,
      new Date(startDate),
      new Date(endDate),
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

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <button use:builder.action {...builder}>
      <div class="gap-x-2 flex" class:inactive={!showComparison}>
        <b>vs</b>

        <p class="line-clamp-1">{label.toLowerCase()}</p>
      </div>
      <span class="flex-none">
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start">
    {#each timeComparisonOptionsState as option}
      {@const preset = TIME_COMPARISON[option.name]}
      {@const selected = selectedLabel === option.name}
      <DropdownMenu.Item
        on:click={() => {
          if (selected) disableAllComparisons();
          else onCompareRangeSelect(option.name);
        }}
      >
        <span class:font-bold={selected}>
          {preset?.label || option.name}
        </span>
      </DropdownMenu.Item>
      {#if option.name === TimeComparisonOption.CONTIGUOUS && timeComparisonOptionsState.length > 2}
        <DropdownMenu.Separator />
      {/if}
    {/each}
    <!-- {#if $timeComparisonOptionsState.length >= 1}
      <DropdownMenu.Separator />
    {/if} -->
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
