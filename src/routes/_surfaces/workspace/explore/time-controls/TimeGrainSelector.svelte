<script lang="ts">
  import type { TimeGrain } from "$common/database-service/DatabaseTimeSeriesActions";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { createEventDispatcher, tick } from "svelte";
  import { prettyTimeGrain, TimeGrainOption } from "./time-range-utils";
  import type { Readable } from "svelte/store";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { selectTimeGrainApi } from "$lib/redux-store/explore/explore-apis.js";
  import { store } from "$lib/redux-store/store-root.js";

  export let metricsDefId: string;

  const dispatch = createEventDispatcher();

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  let selectableTimeGrains: TimeGrainOption[];
  $: selectableTimeGrains = $metricsExplorer?.selectableTimeGrains ?? [];

  let selectedTimeGrain: TimeGrain;
  $: selectedTimeGrain = $metricsExplorer?.selectedTimeGrain;

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

  const onTimeGrainSelect = (timeGrain: TimeGrain) => {
    timeGrainMenuOpen = !timeGrainMenuOpen;
    dispatch("select-time-grain", { timeGrain });
    store.dispatch(selectTimeGrainApi({ metricsDefId, timeGrain }));
  };
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
            on:select={() => onTimeGrainSelect(timeGrain)}
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
