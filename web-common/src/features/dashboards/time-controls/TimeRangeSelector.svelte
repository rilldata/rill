<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import DefaultTimeRangeMenuItem from "@rilldata/web-common/features/dashboards/time-controls/DefaultTimeRangeMenuItem.svelte";
  import TimeRangeScrubChip from "@rilldata/web-common/features/dashboards/time-controls/TimeRangeScrubChip.svelte";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import {
    ALL_TIME,
    DEFAULT_TIME_RANGES,
  } from "@rilldata/web-common/lib/time/config";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import {
    DashboardTimeControls,
    TimeRange,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { Menu, MenuItem } from "../../../components/menu";
  import Divider from "../../../components/menu/core/Divider.svelte";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import type { V1TimeGrain } from "../../../runtime-client";
  import CustomTimeRangeInput from "./CustomTimeRangeInput.svelte";
  import CustomTimeRangeMenuItem from "./CustomTimeRangeMenuItem.svelte";

  export let metricViewName: string;
  export let boundaryStart: Date;
  export let boundaryEnd: Date;
  export let minTimeGrain: V1TimeGrain;
  export let selectedRange: DashboardTimeControls;

  const dispatch = createEventDispatcher();

  $: dashboardStore = useDashboardStore(metricViewName);

  const ctx = getStateManagers();
  const timeControlsStore = useTimeControlStore(ctx);
  const metricsView = useMetricsView(ctx);
  const {
    selectors: {
      timeRangeSelectors: { timeRangeSelectorState },
    },
  } = ctx;

  let isCustomRangeOpen = false;
  let isCalendarRecentlyClosed = false;

  $: hasSubRangeSelected = $dashboardStore?.selectedScrubRange?.end;

  function setIntermediateSelection(timeRange: TimeRange) {
    return () => {
      intermediateSelection =
        timeRange.name + (timeRange.offset ? "_" + timeRange.offset : "");
    };
  }

  function onSelectRelativeTimeRange(
    timeRange: TimeRange,
    closeMenu: () => void,
  ) {
    closeMenu();
    dispatch("select-time-range", {
      name: timeRange.name,
      start: timeRange.start,
      offset: timeRange.offset,
      end: timeRange.end,
    });
  }

  function onSelectCustomTimeRange(
    startDate: string,
    endDate: string,
    closeMenu: () => void,
  ) {
    setIntermediateSelection({ name: TimeRangePreset.CUSTOM })();
    closeMenu();
    dispatch("select-time-range", {
      name: TimeRangePreset.CUSTOM,
      start: startDate,
      end: endDate,
    });
  }

  function zoomScrub(toggleFloatingElement) {
    const { start, end } = getOrderedStartEnd(
      $dashboardStore?.selectedScrubRange?.start,
      $dashboardStore?.selectedScrubRange?.end,
    );
    onSelectRelativeTimeRange(
      {
        name: TimeRangePreset.CUSTOM,
        start,
        end,
      },
      toggleFloatingElement,
    );
    dispatch("remove-scrub");
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

  $: currentSelectedRange = $dashboardStore?.selectedTimeRange?.name;
  $: currentSelectedOffset = $dashboardStore?.selectedTimeRange?.offset;
  $: intermediateSelection =
    currentSelectedRange +
    (currentSelectedOffset ? "_" + currentSelectedOffset : "");

  const handleMenuOpen = () => {
    if (intermediateSelection !== TimeRangePreset.CUSTOM) {
      isCustomRangeOpen = false;
    }
  };

  function timeRangeEquals(timeRange: TimeRange) {
    return (
      intermediateSelection ===
      timeRange.name + (timeRange.offset ? "_" + timeRange.offset : "")
    );
  }
</script>

<WithTogglableFloatingElement
  alignment="start"
  distance={8}
  let:active
  let:toggleFloatingElement
  on:open={handleMenuOpen}
>
  {#if hasSubRangeSelected}
    <div class="flex">
      <TimeRangeScrubChip
        on:click={toggleFloatingElement}
        on:remove={() => dispatch("remove-scrub")}
        {active}
        start={$dashboardStore?.selectedScrubRange?.start}
        end={$dashboardStore?.selectedScrubRange?.end}
        zone={$dashboardStore?.selectedTimezone}
      />
    </div>
  {:else}
    <button
      class:bg-gray-200={active}
      class="px-3 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 items-baseline"
      on:click={toggleFloatingElement}
      aria-label="Select time range"
    >
      <div class="flex flew-row gap-x-3">
        <div class="font-bold flex flex-row items-center gap-x-3">
          <span class="ui-copy-icon"><Calendar size="16px" /></span>
          <span style:transform="translateY(1px)">
            <!-- This conditional shouldn't be necessary because there should always be a selected (at least default) time range -->
            {#if intermediateSelection === TimeRangePreset.CUSTOM}
              Custom range
            {:else if currentSelectedRange}
              {#if currentSelectedRange in DEFAULT_TIME_RANGES}
                {DEFAULT_TIME_RANGES[currentSelectedRange].label}
              {:else if currentSelectedOffset}
                {humaniseISODuration(currentSelectedRange)}, {humaniseISODuration(
                  currentSelectedOffset,
                )} ago
              {:else}
                Last {humaniseISODuration(currentSelectedRange)}
              {/if}
            {:else}
              Select a time range
            {/if}
          </span>
        </div>
        <span style:transform="translateY(1px)">
          {prettyFormatTimeRange(
            $timeControlsStore?.selectedTimeRange?.start,
            $timeControlsStore?.selectedTimeRange?.end,
            $timeControlsStore?.selectedTimeRange?.name,
            $dashboardStore?.selectedTimezone,
          )}
        </span>
      </div>
      <IconSpaceFixer pullRight>
        <div class="transition-transform" class:-rotate-180={active}>
          <CaretDownIcon size="14px" />
        </div>
      </IconSpaceFixer>
    </button>
  {/if}
  <Menu
    label="Time range selector"
    let:toggleFloatingElement
    maxWidth="300px"
    on:click-outside={() => onClickOutside(toggleFloatingElement)}
    on:escape={toggleFloatingElement}
    slot="floating-element"
  >
    {@const allTime = {
      name: TimeRangePreset.ALL_TIME,
      label: ALL_TIME.label,
      start: boundaryStart,
      end: new Date(boundaryEnd.getTime() + 1), // end is exclusive
    }}
    {#if hasSubRangeSelected}
      <MenuItem
        on:before-select={setIntermediateSelection({
          name: TimeRangePreset.CUSTOM,
        })}
        on:select={() => zoomScrub(toggleFloatingElement)}
      >
        <span> Zoom to subrange </span>
        <span slot="right">Z</span>
      </MenuItem>
      <Divider />
    {/if}
    <MenuItem
      on:before-select={setIntermediateSelection(allTime)}
      on:select={() =>
        onSelectRelativeTimeRange(allTime, toggleFloatingElement)}
    >
      <span class:font-bold={timeRangeEquals(allTime)}>
        {allTime.label}
      </span>
    </MenuItem>
    {#if $timeRangeSelectorState.showDefaultItem}
      <DefaultTimeRangeMenuItem
        on:before-select={setIntermediateSelection({
          name: $metricsView.data?.defaultTimeRange,
        })}
        on:select={() =>
          onSelectRelativeTimeRange(
            $timeControlsStore.defaultTimeRange,
            toggleFloatingElement,
          )}
        selected={intermediateSelection === $metricsView.data?.defaultTimeRange}
        isoDuration={$metricsView.data?.defaultTimeRange}
      />
    {/if}
    {#if $timeRangeSelectorState.latestWindowTimeRanges?.length}
      <Divider />
      {#each $timeRangeSelectorState.latestWindowTimeRanges as timeRange}
        <MenuItem
          on:before-select={setIntermediateSelection(timeRange)}
          on:select={() =>
            onSelectRelativeTimeRange(timeRange, toggleFloatingElement)}
        >
          <span class:font-bold={timeRangeEquals(timeRange)}>
            {timeRange.label}
          </span>
        </MenuItem>
      {/each}
    {/if}
    {#if $timeRangeSelectorState.periodToDateRanges?.length}
      <Divider />
      {#each $timeRangeSelectorState.periodToDateRanges as timeRange}
        <MenuItem
          on:before-select={setIntermediateSelection(timeRange)}
          on:select={() =>
            onSelectRelativeTimeRange(timeRange, toggleFloatingElement)}
        >
          <span class:font-bold={timeRangeEquals(timeRange)}>
            {timeRange.label}
          </span>
        </MenuItem>
      {/each}
    {/if}
    <Divider />
    <CustomTimeRangeMenuItem
      on:select={() => {
        isCustomRangeOpen = !isCustomRangeOpen;
      }}
      open={isCustomRangeOpen}
      selected={intermediateSelection === TimeRangePreset.CUSTOM}
    />
    {#if isCustomRangeOpen}
      <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
        <CustomTimeRangeInput
          {boundaryStart}
          {boundaryEnd}
          {minTimeGrain}
          zone={$dashboardStore?.selectedTimezone}
          defaultDate={selectedRange}
          on:apply={(e) =>
            onSelectCustomTimeRange(
              e.detail.startDate,
              e.detail.endDate,
              toggleFloatingElement,
            )}
          on:close-calendar={onCalendarClose}
        />
      </div>
    {/if}
  </Menu>
</WithTogglableFloatingElement>
