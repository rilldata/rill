<script lang="ts">
  import type {
    TimeGrain,
    TimeRangeName,
    TimeSeriesTimeRange,
  } from "$common/database-service/DatabaseTimeSeriesActions";
  import type { MetricViewMetaResponse } from "$common/rill-developer-service/MetricViewActions";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "$lib/components/menu/wrappers/WithSelectMenu.svelte";
  import {
    getMetricViewMetadata,
    getMetricViewMetaQueryKey,
  } from "$lib/svelte-query/queries/metric-view";
  import { useQuery } from "@sveltestack/svelte-query";
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
  const EVENT_NAME = "select-time-grain";

  let selectableTimeGrains: TimeGrainOption[];

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

  $: if (selectedTimeRangeName && allTimeRange) {
    selectableTimeGrains = getSelectableTimeGrains(
      selectedTimeRangeName,
      allTimeRange
    );
  }

  // When the selected time grain is not in the list of selectable time grains (which can
  // happen when the time range name is changed), set the default time grain
  $: isSelectedTimeGrainInvalid =
    selectableTimeGrains &&
    selectableTimeGrains.find(
      (timeGrainOption) => timeGrainOption.timeGrain === selectedTimeGrain
    ).enabled === false;
  $: if (isSelectedTimeGrainInvalid && allTimeRange) {
    const defaultTimeGrain = getDefaultTimeGrain(
      selectedTimeRangeName,
      allTimeRange
    );
    dispatch(EVENT_NAME, { timeGrain: defaultTimeGrain });
  }

  $: options = selectableTimeGrains
    ? selectableTimeGrains.map(({ timeGrain, enabled }) => ({
        main: prettyTimeGrain(timeGrain),
        disabled: !enabled,
        key: timeGrain,
        description: !enabled ? "not valid for this time range" : undefined,
      }))
    : undefined;

  const onTimeGrainSelect = (timeGrain: TimeGrain) => {
    dispatch(EVENT_NAME, { timeGrain });
  };
</script>

{#if selectedTimeGrain && selectableTimeGrains}
  <WithSelectMenu
    {options}
    selection={{
      main: prettyTimeGrain(selectedTimeGrain),
      key: selectedTimeGrain,
    }}
    on:select={(event) => onTimeGrainSelect(event.detail.key)}
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
