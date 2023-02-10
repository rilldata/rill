<script lang="ts">
  import { FloatingElement } from "@rilldata/web-common/components/floating-element";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { onClickOutside } from "@rilldata/web-local/lib/util/on-click-outside";
  import { createEventDispatcher, tick } from "svelte";
  import { slide } from "svelte/transition";
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

  $: if (allTimeRange) {
    selectableRelativeTimeRanges =
      getSelectableRelativeTimeRanges(allTimeRange);
  }

  function onSelectRelativeTimeRange(timeRange: TimeRange) {
    timeRangeNameMenuOpen = !timeRangeNameMenuOpen;
    dispatch("select-time-range", {
      name: timeRange.name,
      start: timeRange.start,
      end: timeRange.end,
    });
  }

  function onSelectCustomTimeRange(startDate: string, endDate: string) {
    timeRangeNameMenuOpen = !timeRangeNameMenuOpen;
    dispatch("select-time-range", {
      name: TimeRangeName.Custom,
      start: startDate,
      end: endDate,
    });
  }

  let customTimeRangeInputOpen = false;

  /// Start boilerplate for DIY Dropdown menu ///
  let timeRangeNameMenu;
  let timeRangeNameMenuOpen = false;
  let clickOutsideListener;
  $: if (!timeRangeNameMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }

  const buttonClickHandler = async () => {
    timeRangeNameMenuOpen = !timeRangeNameMenuOpen;
    if (!clickOutsideListener) {
      await tick();
      clickOutsideListener = onClickOutside(() => {
        if (!customTimeRangeInputOpen) {
          timeRangeNameMenuOpen = false;
        }
      }, timeRangeNameMenu);
    }
  };

  let target: HTMLElement;
  /// End boilerplate for DIY Dropdown menu ///
</script>

<button
  bind:this={target}
  class="px-3 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 transition-tranform duration-100"
  on:click={buttonClickHandler}
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
  <span class="transition-transform" class:-rotate-180={timeRangeNameMenuOpen}>
    <CaretDownIcon size="16px" />
  </span>
</button>

{#if timeRangeNameMenuOpen}
  <div bind:this={timeRangeNameMenu}>
    <FloatingElement
      relationship="direct"
      location="bottom"
      alignment="start"
      {target}
      distance={8}
    >
      <Menu on:escape={() => (timeRangeNameMenuOpen = false)}>
        {#each selectableRelativeTimeRanges as relativeTimeRange}
          <MenuItem
            on:select={() => onSelectRelativeTimeRange(relativeTimeRange)}
          >
            {relativeTimeRange.name}
          </MenuItem>
        {/each}
        <MenuItem on:select={() => onSelectRelativeTimeRange(allTimeRange)}>
          {TimeRangeName.AllTime}
        </MenuItem>
        <hr class="my-2" />
        <CustomTimeRangeMenuItem
          open={customTimeRangeInputOpen}
          on:select={() =>
            (customTimeRangeInputOpen = !customTimeRangeInputOpen)}
        />
        {#if customTimeRangeInputOpen}
          <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
            <CustomTimeRangeInput
              {metricViewName}
              on:apply={(e) =>
                onSelectCustomTimeRange(e.detail.startDate, e.detail.endDate)}
            />
          </div>
        {/if}
      </Menu>
    </FloatingElement>
  </div>
{/if}
