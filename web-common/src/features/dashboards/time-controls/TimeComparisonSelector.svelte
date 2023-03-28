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
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { V1TimeGrain } from "../../../runtime-client";
  import { useDashboardStore } from "../dashboard-stores";
  import CustomTimeRangeInput from "./CustomTimeRangeInput.svelte";
  import CustomTimeRangeMenuItem from "./CustomTimeRangeMenuItem.svelte";
  import SelectorButton from "./SelectorButton.svelte";

  const dispatch = createEventDispatcher();

  export let metricViewName;

  export let currentStart: Date;
  export let currentEnd: Date;
  export let boundaryStart: Date;
  export let boundaryEnd: Date;

  export let showComparison = true;
  export let comparisonOption;
  export let comparisonOptions: TimeComparisonOption[];

  $: dashboardStore = useDashboardStore(metricViewName);

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

  const onCompareRangeSelect = (comparison) => {
    intermediateSelection = comparison;
    dispatch("select-comparison", comparison);
  };
  // Define a better validation criteria
  // FIXME: this is not consumed yet.
  function validateCustomTimeRange(start, end) {
    const customStartDate = new Date(start);
    const customEndDate = new Date(end);
    const selectedEndDate = new Date(currentEnd);
    if (customStartDate > customEndDate)
      return "Start date must be before end date";
    else if (customEndDate > selectedEndDate)
      return "End date must be before selected date";
    else return undefined;
  }

  let isCustomRangeOpen = false;
  let isCalendarRecentlyClosed = false;
  let intermediateSelection = undefined;

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

  $: label = TIME_COMPARISON[comparisonOption]?.label;
  $: intermediateSelection = comparisonOption;
</script>

<Tooltip distance={8}>
  <WithTogglableFloatingElement let:toggleFloatingElement let:active>
    <SelectorButton
      {active}
      disabled={!showComparison}
      on:click={() => {
        if (showComparison) toggleFloatingElement();
      }}
      ><span class="font-normal">
        {#if !showComparison}
          <span class="italic text-gray-500">Time comparison not available</span
          >
        {:else}
          Comparing to <span class="font-bold">{label}</span>
        {/if}
      </span>
    </SelectorButton>
    <Menu
      slot="floating-element"
      on:escape={toggleFloatingElement}
      on:click-outside={toggleFloatingElement}
    >
      {#if showComparison}
        {#each options as option}
          {@const preset = TIME_COMPARISON[option.name]}
          <MenuItem
            selected={option.name === intermediateSelection}
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
      {:else}
        <MenuItem selected={comparisonOption !== TimeComparisonOption.CUSTOM}
          >No comparison</MenuItem
        >
      {/if}
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
            {metricViewName}
            minTimeGrain={/** FIXME-comparisons: */ V1TimeGrain.TIME_GRAIN_MINUTE}
            on:apply={(e) => {
              /** FIXME */
            }}
            on:close-calendar={onCalendarClose}
          />
        </div>
      {/if}
    </Menu>
  </WithTogglableFloatingElement>
  <TooltipContent slot="tooltip-content" maxWidth="220px">
    {#if showComparison}
      Select a time range to compare to the selected time range
    {:else}
      The specified time range does not have any viable comparisons
    {/if}
  </TooltipContent>
</Tooltip>
