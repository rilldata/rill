<!-- @component 
This component needs to do the following:
1. display the set of available comparisons in the menu.
2. dispatch to TimeControl.svelte the selected comparison.
3. read the existing active comparison from somewhere.
-->
<script lang="ts">
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import ClockCircle from "@rilldata/web-common/components/icons/ClockCircle.svelte";
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
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
  export let zone: string;

  export let showComparison = true;
  export let selectedComparison: DashboardTimeControls;

  $: comparisonOption = selectedComparison?.name;

  const {
    selectors: {
      timeRangeSelectors: { timeComparisonOptionsState },
    },
  } = getStateManagers();

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

  $: label =
    showComparison && comparisonOption
      ? TIME_COMPARISON[comparisonOption]?.label
      : NO_COMPARISON_LABEL;

  $: intermediateSelection = showComparison
    ? comparisonOption
    : NO_COMPARISON_LABEL;
</script>

<WithTogglableFloatingElement
  alignment="start"
  distance={8}
  let:active
  let:toggleFloatingElement
>
  <Tooltip distance={8} suppress={active}>
    <SelectorButton
      {active}
      label="Select time comparison option"
      on:click={() => {
        toggleFloatingElement();
      }}
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
  <Menu
    label="Time comparison selector"
    on:click-outside={() => onClickOutside(toggleFloatingElement)}
    on:escape={toggleFloatingElement}
    slot="floating-element"
    let:toggleFloatingElement
  >
    {#each $timeComparisonOptionsState as option}
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
      {#if option.name === TimeComparisonOption.CONTIGUOUS && $timeComparisonOptionsState.length > 2}
        <Divider />
      {/if}
    {/each}
    {#if $timeComparisonOptionsState.length >= 1}
      <Divider />
    {/if}

    <CustomTimeRangeMenuItem
      on:select={() => {
        isCustomRangeOpen = !isCustomRangeOpen;
      }}
      open={isCustomRangeOpen}
    />
    {#if isCustomRangeOpen}
      <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
        <CustomTimeRangeInput
          {boundaryStart}
          {boundaryEnd}
          defaultDate={selectedComparison}
          {minTimeGrain}
          {zone}
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
