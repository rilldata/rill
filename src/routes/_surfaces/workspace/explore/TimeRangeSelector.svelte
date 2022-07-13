<script lang="ts">
  import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";

  import Button from "$lib/components/Button.svelte";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
  import { setExploreSelectedTimeRangeAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";
  import { store } from "$lib/redux-store/store-root";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { tick } from "svelte";
  import {
    getTimeRangeNameForButton,
    makeSelectableTimeRanges,
    prettyFormatTimeRange,
  } from "./utils";

  export let metricsDefId: string;

  $: metricsLeaderboard = getMetricsExploreById(metricsDefId);

  let selectableTimeRanges: TimeSeriesTimeRange[];
  $: if ($metricsLeaderboard?.timeRange) {
    selectableTimeRanges = makeSelectableTimeRanges(
      $metricsLeaderboard.timeRange
    );
  }
  let selectedTimeRange: TimeSeriesTimeRange;
  $: if ($metricsLeaderboard?.selectedTimeRange) {
    selectedTimeRange = $metricsLeaderboard?.selectedTimeRange;
  } else if (selectableTimeRanges) {
    selectedTimeRange = selectableTimeRanges[0];
  }

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

<Button
  bind:element={target}
  override="border-none gap-x-4"
  on:click={buttonClickHandler}
>
  <span class="font-bold">
    {getTimeRangeNameForButton(selectedTimeRange)}
  </span>
  <span>
    {prettyFormatTimeRange(selectedTimeRange)}
  </span>
  <CaretDownIcon size="16px" />
</Button>

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
