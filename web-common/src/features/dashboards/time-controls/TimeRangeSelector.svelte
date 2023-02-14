<script lang="ts">
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { Menu, MenuItem } from "../../../components/menu";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";
  import CustomTimeRangeInput from "./CustomTimeRangeInput.svelte";
  import CustomTimeRangeMenuItem from "./CustomTimeRangeMenuItem.svelte";
  import { TimeRange, TimeRangeName } from "./time-control-types";
  import {
    getSelectableRelativeTimeRanges,
    prettyFormatTimeRange,
  } from "./time-range-utils";

  export let metricViewName: string;
  export let selectedTimeRange: TimeRange;
  export let allTimeRange: TimeRange;

  const dispatch = createEventDispatcher();

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  let selectableRelativeTimeRanges: TimeRange[];
  let isCustomRangeOpen = false;
  let isCalendarRecentlyClosed = false;

  $: if (allTimeRange) {
    selectableRelativeTimeRanges =
      getSelectableRelativeTimeRanges(allTimeRange);
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
    closeMenu();
    dispatch("select-time-range", {
      name: TimeRangeName.Custom,
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
    }, 500);
  }
</script>

<WithTogglableFloatingElement
  let:toggleFloatingElement
  let:active
  distance={8}
  alignment="start"
>
  <button
    class="px-3 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 transition-tranform duration-100"
    on:click={toggleFloatingElement}
  >
    <div class="flex flew-row gap-x-3">
      <div class="font-bold flex flex-row items-center gap-x-3">
        <!-- This conditional shouldn't be necessary because there should always be a selected (at least default) time range -->
        <span class="ui-copy-icon"><Calendar size="16px" /></span>
        <span style:transform="translateY(1px)">
          {selectedTimeRange.name ?? "Select a time range"}
        </span>
      </div>
      <span style:transform="translateY(1px)">
        {prettyFormatTimeRange(metricsExplorer?.selectedTimeRange)}
      </span>
    </div>
    <span class="transition-transform" class:-rotate-180={active}>
      <CaretDownIcon size="16px" />
    </span>
  </button>
  <Menu
    slot="floating-element"
    on:escape={toggleFloatingElement}
    on:click-outside={() => onClickOutside(toggleFloatingElement)}
  >
    {#each selectableRelativeTimeRanges as relativeTimeRange}
      <MenuItem
        on:select={() =>
          onSelectRelativeTimeRange(relativeTimeRange, toggleFloatingElement)}
      >
        {relativeTimeRange.name}
      </MenuItem>
    {/each}
    <MenuItem
      on:select={() =>
        onSelectRelativeTimeRange(allTimeRange, toggleFloatingElement)}
    >
      {TimeRangeName.AllTime}
    </MenuItem>
    <hr class="my-2" />
    <CustomTimeRangeMenuItem
      open={isCustomRangeOpen}
      on:select={() => (isCustomRangeOpen = !isCustomRangeOpen)}
    />
    {#if isCustomRangeOpen}
      <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
        <CustomTimeRangeInput
          {metricViewName}
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
