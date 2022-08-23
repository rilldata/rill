<script lang="ts">
  import type {
    TimeRangeName,
    TimeSeriesTimeRange,
  } from "$common/database-service/DatabaseTimeSeriesActions";
  import type { MetricViewMetaResponse } from "$common/rill-developer-service/MetricViewActions";
  import { FloatingElement } from "$lib/components/floating-element";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import { Menu, MenuItem } from "$lib/components/menu";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import {
    getMetricViewMetadata,
    getMetricViewMetaQueryKey,
  } from "$lib/svelte-query/queries/metric-view";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { useQuery } from "@sveltestack/svelte-query";
  import { createEventDispatcher, tick } from "svelte";
  import {
    getSelectableTimeRangeNames,
    makeTimeRanges,
    prettyFormatTimeRange,
  } from "./time-range-utils";
  import { metricsExplorerStore } from "$lib/application-state-stores/explorer-stores";

  export let metricsDefId: string;
  export let selectedTimeRangeName: TimeRangeName;

  const dispatch = createEventDispatcher();
  const EVENT_NAME = "select-time-range-name";

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  let selectableTimeRanges: TimeSeriesTimeRange[];

  // query the `/meta` endpoint to get the all time range of the dataset
  let queryKey = getMetricViewMetaQueryKey(metricsDefId);
  const queryResult = useQuery<MetricViewMetaResponse, Error>(queryKey, () =>
    getMetricViewMetadata(metricsDefId)
  );
  $: {
    queryKey = getMetricViewMetaQueryKey(metricsDefId);
    queryResult.setOptions(queryKey, () => getMetricViewMetadata(metricsDefId));
  }
  let allTimeRange: TimeSeriesTimeRange;
  $: allTimeRange = $queryResult.data?.timeDimension?.timeRange;

  // TODO: move this logic to server-side and fetch the results from the `/meta` endpoint directly
  const getSelectableTimeRanges = (
    allTimeRangeInDataset: TimeSeriesTimeRange
  ) => {
    const selectableTimeRangeNames = getSelectableTimeRangeNames(
      allTimeRangeInDataset
    );
    const selectableTimeRanges = makeTimeRanges(
      selectableTimeRangeNames,
      allTimeRangeInDataset
    );
    return selectableTimeRanges;
  };
  $: if (allTimeRange) {
    selectableTimeRanges = getSelectableTimeRanges(allTimeRange);
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

  const onTimeRangeSelect = (timeRangeName: TimeRangeName) => {
    timeRangeNameMenuOpen = !timeRangeNameMenuOpen;
    dispatch(EVENT_NAME, { timeRangeName });
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
