<script lang="ts">
  import type {
    TimeRangeName,
    TimeSeriesTimeRange,
  } from "$common/database-service/DatabaseTimeSeriesActions";
  import { FloatingElement } from "$lib/components/floating-element";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import { Menu, MenuItem } from "$lib/components/menu";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { createEventDispatcher, tick } from "svelte";
  import { prettyFormatTimeRange } from "./time-range-utils";
  import type { Readable } from "svelte/store";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { selectTimeRangeNameApi } from "$lib/redux-store/explore/explore-apis";
  import { store } from "$lib/redux-store/store-root";

  export let metricsDefId: string;

  const dispatch = createEventDispatcher();

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  let selectableTimeRanges: TimeSeriesTimeRange[];
  $: selectableTimeRanges = $metricsExplorer?.selectableTimeRanges ?? [];

  let selectedTimeRangeName: TimeRangeName;
  $: selectedTimeRangeName = $metricsExplorer?.selectedTimeRange?.name;

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

  const onTimeRangeSelect = (timeRangeName: TimeRangeName) => {
    timeRangeNameMenuOpen = !timeRangeNameMenuOpen;
    store.dispatch(
      selectTimeRangeNameApi({
        metricsDefId,
        timeRangeName,
      })
    );
    dispatch("select-time-range-name", {
      timeRangeName,
    });
  };
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
      {prettyFormatTimeRange($metricsExplorer?.selectedTimeRange)}
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
          <MenuItem on:select={() => onTimeRangeSelect(timeRange.name)}>
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
