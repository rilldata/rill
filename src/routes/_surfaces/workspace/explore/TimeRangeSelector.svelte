<script lang="ts">
  import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";

  import FloatingElement from "$lib/components/floating-element/FloatingElement.svelte";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import { Menu, MenuItem } from "$lib/components/menu";
  import { setExploreSelectedTimeRangeAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import { store } from "$lib/redux-store/store-root";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { tick } from "svelte";
  import { getTimeRangeNameForButton, prettyFormatTimeRange } from "./utils";

  export let metricsDefId: string;

  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  let selectableTimeRanges: TimeSeriesTimeRange[];
  $: selectableTimeRanges = $metricsExplorer?.selectableTimeRanges;
  let selectedTimeRange: TimeSeriesTimeRange;
  $: if ($metricsExplorer?.selectedTimeRange)
    selectedTimeRange = selectableTimeRanges.find(
      (selectableTimeRange) =>
        selectableTimeRange.start ===
          $metricsExplorer?.selectedTimeRange?.start &&
        selectableTimeRange.end === $metricsExplorer?.selectedTimeRange?.end
    );

  let timeSelectorMenu;
  let timeSelectorMenuOpen = false;
  let clickOutsideListener;
  $: if (!timeSelectorMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }

  const buttonClickHandler = async () => {
    timeSelectorMenuOpen = !timeSelectorMenuOpen;
    if (!clickOutsideListener) {
      await tick();
      clickOutsideListener = onClickOutside(() => {
        timeSelectorMenuOpen = false;
      }, timeSelectorMenu);
    }
  };

  let target: HTMLElement;
</script>

<button
  bind:this={target}
  class="px-4 py-2 rounded flex flex-row gap-x-4 hover:bg-gray-200 transition-tranform duration-100"
  on:click={buttonClickHandler}
>
  <span class="font-bold">
    {getTimeRangeNameForButton(selectedTimeRange)}
  </span>
  <span>
    {prettyFormatTimeRange(selectedTimeRange)}
  </span>
  <span class="transition-transform" class:-rotate-180={timeSelectorMenuOpen}>
    <CaretDownIcon size="16px" />
  </span>
</button>

{#if timeSelectorMenuOpen}
  <div bind:this={timeSelectorMenu}>
    <FloatingElement
      relationship="direct"
      location="bottom"
      alignment="start"
      {target}
      distance={8}
    >
      <Menu on:escape={() => (timeSelectorMenuOpen = false)}>
        {#each selectableTimeRanges as timeRange}
          <MenuItem
            on:select={() =>
              setExploreSelectedTimeRangeAndUpdate(
                store.dispatch,
                metricsDefId,
                {
                  name: timeRange.name,
                  start: timeRange.start,
                  end: timeRange.end,
                }
              )}
          >
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
