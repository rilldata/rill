<!-- @component 
This component needs to do the following:
1. display the set of available comparisons in the menu.
2. dispatch to TimeControl.svelte the selected comparison.
3. read the existing active comparison from somewhere.
-->
<script lang="ts">
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
  import type { V1TimeGrain } from "../../../runtime-client";
  import CustomTimeRangeInput from "./CustomTimeRangeInput.svelte";
  import SelectorButton from "./SelectorButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

  const dispatch = createEventDispatcher();

  export let currentStart: Date;
  export let currentEnd: Date;
  export let boundaryStart: Date;
  export let boundaryEnd: Date;
  export let minTimeGrain: V1TimeGrain;
  export let zone: string;
  export let showComparison = true;
  export let selectedComparison: DashboardTimeControls | undefined;

  const {
    selectors: {
      timeRangeSelectors: { timeComparisonOptionsState },
    },
  } = getStateManagers();

  let open = false;

  $: comparisonOption = selectedComparison?.name;

  $: label =
    showComparison && comparisonOption
      ? TIME_COMPARISON[comparisonOption]?.label
      : NO_COMPARISON_LABEL;

  $: intermediateSelection = showComparison
    ? comparisonOption
    : NO_COMPARISON_LABEL;

  function onSelectCustomComparisonRange(startDate: string, endDate: string) {
    intermediateSelection = TimeComparisonOption.CUSTOM;

    dispatch("select-comparison", {
      name: TimeComparisonOption.CUSTOM,
      start: new Date(startDate),
      end: new Date(endDate),
    });
  }

  function onCompareRangeSelect(comparisonOption: TimeComparisonOption) {
    const comparisonTimeRange = getComparisonRange(
      currentStart,
      currentEnd,
      comparisonOption,
    );

    dispatch("select-comparison", {
      name: comparisonOption,
      start: comparisonTimeRange.start,
      end: comparisonTimeRange.end,
    });
  }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip distance={8} suppress={open}>
      <SelectorButton
        builders={[builder]}
        active={open}
        label="Select time comparison option"
      >
        <div class="flex items-center gap-x-3">
          <span class="ui-copy-icon"><ClockCircle size="16px" /></span>
          <span
            class="font-normal justify-center"
            style:transform="translateY(-1px)">{label}</span
          >
        </div>
      </SelectorButton>
      <TooltipContent maxWidth="220px" slot="tooltip-content">
        Select a time range to compare to the selected time range
      </TooltipContent>
    </Tooltip>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start">
    {#each $timeComparisonOptionsState as option}
      {@const preset = TIME_COMPARISON[option.name]}
      <DropdownMenu.Item
        on:click={() => {
          onCompareRangeSelect(option.name);
        }}
      >
        <span class:font-bold={intermediateSelection === option.name}>
          {preset?.label || option.name}
        </span>
      </DropdownMenu.Item>
      {#if option.name === TimeComparisonOption.CONTIGUOUS && $timeComparisonOptionsState.length > 2}
        <DropdownMenu.Separator />
      {/if}
    {/each}
    {#if $timeComparisonOptionsState.length >= 1}
      <DropdownMenu.Separator />
    {/if}

    <CustomTimeRangeInput
      {boundaryStart}
      {boundaryEnd}
      defaultDate={selectedComparison}
      {minTimeGrain}
      {zone}
      on:apply={(e) => {
        onSelectCustomComparisonRange(e.detail.startDate, e.detail.endDate);
      }}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
