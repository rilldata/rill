<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    ALL_TIME,
    DEFAULT_TIME_RANGES,
    LATEST_WINDOW_TIME_RANGES,
    PERIOD_TO_DATE_RANGES,
  } from "@rilldata/web-common/lib/time/config";
  import {
    getChildTimeRanges,
    prettyFormatTimeRange,
  } from "@rilldata/web-common/lib/time/time-range";
  import {
    TimeRange,
    TimeRangeOption,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { Menu, MenuItem } from "../../../components/menu";
  import Divider from "../../../components/menu/core/Divider.svelte";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import type { V1TimeGrain } from "../../../runtime-client";
  import { useDashboardStore } from "../dashboard-stores";
  import CustomTimeRangeInput from "./CustomTimeRangeInput.svelte";
  import CustomTimeRangeMenuItem from "./CustomTimeRangeMenuItem.svelte";

  export let metricViewName: string;
  export let allTimeRange: TimeRange;
  export let minTimeGrain: V1TimeGrain;

  const dispatch = createEventDispatcher();

  $: dashboardStore = useDashboardStore(metricViewName);

  let isCustomRangeOpen = false;
  let isCalendarRecentlyClosed = false;

  let latestWindowTimeRanges: TimeRangeOption[];
  let periodToDateTimeRanges: TimeRangeOption[];

  // get the available latest-window time ranges
  $: if (allTimeRange) {
    latestWindowTimeRanges = getChildTimeRanges(
      allTimeRange.start,
      allTimeRange.end,
      LATEST_WINDOW_TIME_RANGES,
      minTimeGrain
    );
  }

  // get the the available period-to-date time ranges
  $: if (allTimeRange) {
    periodToDateTimeRanges = getChildTimeRanges(
      allTimeRange.start,
      allTimeRange.end,
      PERIOD_TO_DATE_RANGES,
      minTimeGrain
    );
  }

  function setIntermediateSelection(timeRangeName: string) {
    return () => {
      intermediateSelection = timeRangeName;
    };
  }

  function onSelectRelativeTimeRange(
    timeRange: TimeRange,
    closeMenu: () => void
  ) {
    closeMenu();
    dispatch("select-time-range", {
      name: timeRange.name,
      start: timeRange.start,
      end: timeRange.end,
    });
  }

  function onSelectCustomTimeRange(
    startDate: string,
    endDate: string,
    closeMenu: () => void
  ) {
    setIntermediateSelection(TimeRangePreset.CUSTOM)();
    closeMenu();
    dispatch("select-time-range", {
      name: TimeRangePreset.CUSTOM,
      start: startDate,
      end: endDate,
    });
  }

  function onClickOutside(closeMenu: () => void) {
    if (!isCalendarRecentlyClosed) {
      closeMenu();
    }
  }

  // If the user clicks outside to close the calendar, we don't want the `click-outside` event to close the whole menu.
  // A nice solution would be to check for `isCalendarOpen` in the `onClickOutside` function. However, the calendar
  // closes *before* the `click-outside` event is fired. A workaround is to check for `isCalendarRecentlyClosed`.
  function onCalendarClose() {
    isCalendarRecentlyClosed = true;
    setTimeout(() => {
      isCalendarRecentlyClosed = false;
    }, 300);
  }

  $: currentSelection = $dashboardStore?.selectedTimeRange?.name;
  $: intermediateSelection = currentSelection;
</script>

<WithTogglableFloatingElement
  alignment="start"
  distance={8}
  let:active
  let:toggleFloatingElement
>
  <button
    class:bg-gray-200={active}
    class="px-3 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 items-baseline"
    on:click={toggleFloatingElement}
  >
    <div class="flex flew-row gap-x-3">
      <div class="font-bold flex flex-row items-center gap-x-3">
        <span class="ui-copy-icon"><Calendar size="16px" /></span>
        <span style:transform="translateY(1px)">
          <!-- This conditional shouldn't be necessary because there should always be a selected (at least default) time range -->
          {#if intermediateSelection === TimeRangePreset.CUSTOM}
            Custom range
          {:else if currentSelection in DEFAULT_TIME_RANGES}
            {DEFAULT_TIME_RANGES[$dashboardStore?.selectedTimeRange?.name]
              .label}
          {:else}
            Select a time range
          {/if}
        </span>
      </div>
      <span style:transform="translateY(1px)">
        {prettyFormatTimeRange(
          $dashboardStore?.selectedTimeRange?.start,
          $dashboardStore?.selectedTimeRange?.end
        )}
      </span>
    </div>
    <IconSpaceFixer pullRight>
      <div class="transition-transform" class:-rotate-180={active}>
        <CaretDownIcon size="14px" />
      </div>
    </IconSpaceFixer>
  </button>
  <Menu
    on:click-outside={() => onClickOutside(toggleFloatingElement)}
    on:escape={toggleFloatingElement}
    slot="floating-element"
  >
    {@const allTime = {
      name: TimeRangePreset.ALL_TIME,
      label: ALL_TIME.label,
      start: allTimeRange.start,
      end: allTimeRange.end,
    }}
    <MenuItem
      on:before-select={setIntermediateSelection(allTime.name)}
      on:select={() =>
        onSelectRelativeTimeRange(allTime, toggleFloatingElement)}
    >
      <span class:font-bold={intermediateSelection === allTime.name}>
        {allTime.label}
      </span>
    </MenuItem>
    {#if latestWindowTimeRanges}
      <Divider />

      {#each latestWindowTimeRanges as timeRange}
        <MenuItem
          on:before-select={setIntermediateSelection(timeRange.name)}
          on:select={() =>
            onSelectRelativeTimeRange(timeRange, toggleFloatingElement)}
        >
          <span class:font-bold={intermediateSelection === timeRange.name}>
            {timeRange.label}
          </span>
        </MenuItem>
      {/each}
    {/if}
    {#if periodToDateTimeRanges}
      <Divider />
      {#each periodToDateTimeRanges as timeRange}
        <MenuItem
          on:before-select={setIntermediateSelection(timeRange.name)}
          on:select={() =>
            onSelectRelativeTimeRange(timeRange, toggleFloatingElement)}
        >
          <span class:font-bold={intermediateSelection === timeRange.name}>
            {timeRange.label}
          </span>
        </MenuItem>
      {/each}
      <Divider />
    {/if}
    <CustomTimeRangeMenuItem
      on:select={() => {
        isCustomRangeOpen = !isCustomRangeOpen;
      }}
      open={isCustomRangeOpen}
      selected={intermediateSelection === TimeRangePreset.CUSTOM}
    />
    {#if isCustomRangeOpen}
      <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
        <CustomTimeRangeInput
          {allTimeRange}
          {metricViewName}
          {minTimeGrain}
          on:apply={(e) =>
            onSelectCustomTimeRange(
              e.detail.startDate,
              e.detail.endDate,
              toggleFloatingElement
            )}
          on:close-calendar={onCalendarClose}
        />
      </div>
    {/if}
  </Menu>
</WithTogglableFloatingElement>
