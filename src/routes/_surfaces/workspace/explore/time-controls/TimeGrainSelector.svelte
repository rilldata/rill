<script lang="ts">
  import type {
    TimeGrain,
    TimeRangeName,
  } from "$common/database-service/DatabaseTimeSeriesActions";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "$lib/components/menu/wrappers/WithSelectMenu.svelte";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import { createEventDispatcher } from "svelte";
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
  const SELECT_TIME_GRAIN = "select-time-grain";

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
  $: isSelectedTimeGrainInvalid =
    selectableTimeGrains &&
    selectableTimeGrains.find(
      (timeGrainOption) => timeGrainOption.timeGrain === selectedTimeGrain
    ).enabled === false;
  $: if (isSelectedTimeGrainInvalid && $metricsExplorer?.allTimeRange) {
    const defaultTimeGrain = getDefaultTimeGrain(
      selectedTimeRangeName,
      $metricsExplorer.allTimeRange
    );
    dispatch(SELECT_TIME_GRAIN, { timeGrain: defaultTimeGrain });
  }

  $: options = selectableTimeGrains
    ? selectableTimeGrains.map(({ timeGrain, enabled }) => ({
        main: timeGrain,
        disabled: !enabled,
        key: timeGrain,
        description: !enabled ? "not valid for this time range" : undefined,
      }))
    : undefined;
</script>

{#if selectedTimeGrain && selectableTimeGrains}
  <WithSelectMenu
    {options}
    selection={{ main: selectedTimeGrain, key: selectedTimeGrain }}
    on:select={(event) => {
      dispatch(SELECT_TIME_GRAIN, { timeGrain: event.detail.key });
    }}
    let:toggleMenu
    let:active
  >
    <button
      class="px-4 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 transition-tranform duration-100"
      on:click={toggleMenu}
    >
      <span class="font-bold"
        >by {prettyTimeGrain(selectedTimeGrain)} increments</span
      >
      <span class="transition-transform" class:-rotate-180={active}>
        <CaretDownIcon size="16px" />
      </span>
    </button>
  </WithSelectMenu>
{/if}
