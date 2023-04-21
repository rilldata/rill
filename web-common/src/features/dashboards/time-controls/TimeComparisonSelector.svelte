<!-- @component 
This component needs to do the following:
1. display the set of available comparisons in the menu.
2. dispatch to TimeControl.svelte the selected comparison.
3. read the existing active comparison from somewhere.
-->
<script lang="ts">
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { getComparisonRange } from "@rilldata/web-common/lib/time/comparisons";
  import {
    NO_COMPARISON_LABEL,
    TIME_COMPARISON,
  } from "@rilldata/web-common/lib/time/config";
  import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import type { V1TimeGrain } from "../../../runtime-client";
  import CustomTimeRangeInput from "./CustomTimeRangeInput.svelte";
  import CustomTimeRangeMenuItem from "./CustomTimeRangeMenuItem.svelte";
  import SelectorButton from "./SelectorButton.svelte";

  const dispatch = createEventDispatcher();

  export let currentStart: Date;
  export let currentEnd: Date;
  export let boundaryStart: Date;
  export let boundaryEnd: Date;
  export let minTimeGrain: V1TimeGrain;

  export let showComparison = true;
  export let isComparisonRangeAvailable = true;
  export let selectedComparison;
  export let comparisonOptions: TimeComparisonOption[];

  $: comparisonOption = selectedComparison?.name;

  /** compile the comparison options */
  let options: {
    name: TimeComparisonOption;
    start: Date;
    end: Date;
  }[];
  $: if (comparisonOptions !== undefined)
    options = Object.entries(comparisonOptions)?.map(([key, value]) => {
      const comparisonTimeRange = getComparisonRange(
        currentStart,
        currentEnd,
        value
      );
      return {
        name: value,
        key,
        start: comparisonTimeRange.start,
        end: comparisonTimeRange.end,
      };
    });

  function onSelectCustomComparisonRange(
    startDate: string,
    endDate: string,
    closeMenu: () => void
  ) {
    intermediateSelection = TimeComparisonOption.CUSTOM;
    closeMenu();
    dispatch("select-comparison", {
      name: TimeComparisonOption.CUSTOM,
      start: new Date(startDate),
      end: new Date(endDate),
    });
  }

  const onCompareRangeSelect = (comparisonOption) => {
    const comparisonTimeRange = getComparisonRange(
      currentStart,
      currentEnd,
      comparisonOption
    );

    dispatch("select-comparison", {
      name: comparisonOption,
      start: comparisonTimeRange.start,
      end: comparisonTimeRange.end,
    });
  };

  let isCustomRangeOpen = false;
  let isCalendarRecentlyClosed = false;

  function onClickOutside(closeMenu: () => void) {
    if (!isCalendarRecentlyClosed) {
      closeMenu();
    }
  }

  function onCalendarClose() {
    isCalendarRecentlyClosed = true;
    setTimeout(() => {
      isCalendarRecentlyClosed = false;
    }, 300);
  }

  $: label = showComparison
    ? TIME_COMPARISON[comparisonOption]?.label
    : NO_COMPARISON_LABEL;

  $: intermediateSelection = showComparison
    ? comparisonOption
    : NO_COMPARISON_LABEL;
</script>

<Tooltip distance={8}>
  <WithTogglableFloatingElement let:toggleFloatingElement let:active>
    <SelectorButton
      {active}
      disabled={!isComparisonRangeAvailable}
      on:click={() => {
        if (isComparisonRangeAvailable) toggleFloatingElement();
      }}
    >
      <span class="font-normal">
        {#if !isComparisonRangeAvailable}
          <span class="italic text-gray-500">Time comparison not available</span
          >
        {:else}
          {showComparison ? "Comparing to" : ""}
          <span class="font-bold">{label}</span>
        {/if}
      </span>
    </SelectorButton>
    <Menu
      slot="floating-element"
      on:escape={toggleFloatingElement}
      on:click-outside={() => onClickOutside(toggleFloatingElement)}
    >
      <MenuItem
        selected={!showComparison}
        on:before-select={() => {
          intermediateSelection = NO_COMPARISON_LABEL;
        }}
        on:select={() => {
          dispatch("disable-comparison");
          toggleFloatingElement();
        }}
      >
        <span class:font-bold={intermediateSelection === NO_COMPARISON_LABEL}>
          {NO_COMPARISON_LABEL}
        </span>
      </MenuItem>
      <Divider />
      {#each options as option}
        {@const preset = TIME_COMPARISON[option.name]}
        <MenuItem
          selected={option.name === intermediateSelection}
          on:before-select={() => {
            intermediateSelection = option.name;
          }}
          on:select={() => {
            onCompareRangeSelect(option.name);
            toggleFloatingElement();
          }}
        >
          <span class:font-bold={intermediateSelection === option.name}>
            {preset?.label || option.name}
          </span>
        </MenuItem>
        {#if option.name === TimeComparisonOption.CONTIGUOUS && options.length > 2}
          <Divider />
        {/if}
      {/each}
      {#if options.length >= 1}
        <Divider />
      {/if}

      <CustomTimeRangeMenuItem
        on:select={() => {
          isCustomRangeOpen = !isCustomRangeOpen;
        }}
        open={isCustomRangeOpen}
      />
      {#if isCustomRangeOpen}
        <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
          <CustomTimeRangeInput
            {boundaryStart}
            {boundaryEnd}
            defaultDate={selectedComparison}
            {minTimeGrain}
            on:apply={(e) => {
              onSelectCustomComparisonRange(
                e.detail.startDate,
                e.detail.endDate,
                toggleFloatingElement
              );
            }}
            on:close-calendar={onCalendarClose}
          />
        </div>
      {/if}
    </Menu>
  </WithTogglableFloatingElement>
  <TooltipContent slot="tooltip-content" maxWidth="220px">
    {#if isComparisonRangeAvailable}
      Select a time range to compare to the selected time range
    {:else}
      Select a shorter or more recent time range to enable comparisons.
    {/if}
  </TooltipContent>
</Tooltip>
