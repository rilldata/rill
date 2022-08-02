<script lang="ts">
  import type {
    TimeRangeName,
    TimeSeriesTimeRange,
  } from "$common/database-service/DatabaseTimeSeriesActions";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { tick } from "svelte";
  import {
    getDefaultTimeRangeName,
    getSelectableTimeRangeNames,
    makeTimeRanges,
    prettyFormatTimeRange,
  } from "./timeRangeUtils";

  export let metricsDefId: string;
  export let selectedTimeRangeName: TimeRangeName;
  export let onSelectTimeRangeName: (timeRangeName: TimeRangeName) => void;

  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  let selectableTimeRanges: TimeSeriesTimeRange[];

  const getSelectableTimeRanges = (
    allTimeRangeInDataset: TimeSeriesTimeRange
  ) => {
    // TODO: replace this with a call to the `/meta` endpoint, once available.
    const selectableTimeRangeNames = getSelectableTimeRangeNames(
      allTimeRangeInDataset
    );
    const selectableTimeRanges = makeTimeRanges(
      selectableTimeRangeNames,
      allTimeRangeInDataset
    );
    return selectableTimeRanges;
  };
  $: if ($metricsExplorer?.allTimeRange) {
    selectableTimeRanges = getSelectableTimeRanges(
      $metricsExplorer.allTimeRange
    );
  }

  $: if (!selectedTimeRangeName) {
    const defaultTimeRangeName = getDefaultTimeRangeName();
    onSelectTimeRangeName(defaultTimeRangeName);
  }

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
        timeRangeNameMenuOpen = false;
      }, timeRangeNameMenu);
    }
  };

  let target: HTMLElement;
  /// End boilerplate for DIY Dropdown menu ///
</script>

<button
  bind:this={target}
  class="px-4 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 transition-tranform duration-100"
  on:click={buttonClickHandler}
>
  <div class="flex flew-row gap-x-3">
    <span class="font-bold">
      <!-- This conditional shouldn't be necessary because there should always be a selected (at least default) time range -->
      {selectedTimeRangeName ?? "Select a time range"}
    </span>
    <span>
      {prettyFormatTimeRange($metricsExplorer.selectedTimeRange)}
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
        {#each selectableTimeRanges as timeRange}
          <MenuItem on:select={() => onSelectTimeRangeName(timeRange.name)}>
            <div class="font-bold">
              {timeRange.name}
            </div>
            <div slot="right" let:hovered class:opacity-0={!hovered}>
              {prettyFormatTimeRange(timeRange)}
            </div>
          </MenuItem>
        {/each}
      </Menu>
    </FloatingElement>
  </div>
{/if}
