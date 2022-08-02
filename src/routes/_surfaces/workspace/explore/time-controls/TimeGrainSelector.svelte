<script lang="ts">
  import type {
    TimeGrain,
    TimeRangeName,
  } from "$common/database-service/DatabaseTimeSeriesActions";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { createEventDispatcher, tick } from "svelte";
  import {
    getDefaultTimeGrain,
    getSelectableTimeGrains,
    prettyTimeGrain,
    TimeGrainOption,
  } from "./time-range-utils";

  export let metricsDefId: string;
  export let selectedTimeRangeName: TimeRangeName;
  export let selectedTimeGrain: TimeGrain;

  const dispatch = createEventDispatcher();

  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  let selectableTimeGrains: TimeGrainOption[];

  // TODO: replace this with a call to the `/meta` endpoint, once available.
  $: if (selectedTimeRangeName && $metricsExplorer?.allTimeRange) {
    selectableTimeGrains = getSelectableTimeGrains(
      selectedTimeRangeName,
      $metricsExplorer.allTimeRange
    );
  }

  // When the selected time grain is not in the list of selectable time grains (which can
  // happen when the time range name is changed), set the default time grain
  $: if (
    selectableTimeGrains &&
    selectableTimeGrains.find(
      (timeGrainOption) => timeGrainOption.timeGrain === selectedTimeGrain
    ).enabled === false
  ) {
    const defaultTimeGrain = getDefaultTimeGrain(selectedTimeRangeName);
    dispatch("select-time-grain", { timeGrain: defaultTimeGrain });
  }

  /// Start boilerplate for DIY Dropdown menu ///
  let timeSelectorMenu;
  let timeGrainMenuOpen = false;
  let clickOutsideListener;
  $: if (!timeGrainMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }

  const buttonClickHandler = async () => {
    timeGrainMenuOpen = !timeGrainMenuOpen;
    if (!clickOutsideListener) {
      await tick();
      clickOutsideListener = onClickOutside(() => {
        timeGrainMenuOpen = false;
      }, timeSelectorMenu);
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
  <span class="font-bold"
    >by {prettyTimeGrain(selectedTimeGrain)} increments</span
  >
  <span class="transition-transform" class:-rotate-180={timeGrainMenuOpen}>
    <CaretDownIcon size="16px" />
  </span>
</button>

{#if timeGrainMenuOpen}
  <div bind:this={timeSelectorMenu}>
    <FloatingElement
      relationship="direct"
      location="bottom"
      alignment="start"
      {target}
      distance={8}
    >
      <Menu on:escape={() => (timeGrainMenuOpen = false)}>
        {#each selectableTimeGrains as { timeGrain, enabled }}
          <MenuItem
            disabled={!enabled}
            on:select={() => {
              timeGrainMenuOpen = !timeGrainMenuOpen;
              dispatch("select-time-grain", { timeGrain });
            }}
          >
            <div class={!enabled ? "text-gray-500" : "font-bold "}>
              {prettyTimeGrain(timeGrain)}
            </div>
            <svelte:fragment slot="description">
              <div class="italic">
                {#if !enabled}
                  not valid for this time range
                {/if}
              </div>
            </svelte:fragment>
          </MenuItem>
        {/each}
      </Menu>
    </FloatingElement>
  </div>
{/if}
