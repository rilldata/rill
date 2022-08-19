<script lang="ts">
  import type {
    TimeGrain,
    TimeRangeName,
  } from "$common/database-service/DatabaseTimeSeriesActions";
  import type { MetricViewMetaResponse } from "$common/rill-developer-service/MetricViewActions";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "$lib/components/menu/wrappers/WithSelectMenu.svelte";
  import { updateSelectedTimeGrainApi } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { store } from "$lib/redux-store/store-root";
  import {
    getMetricViewMetadata,
    getMetricViewMetaQueryKey,
  } from "$lib/svelte-query/queries/metric-view";
  import { useQuery } from "@sveltestack/svelte-query";
  import type { Readable } from "svelte/store";
  import {
    getSelectableTimeGrains,
    prettyTimeGrain,
    TimeGrainOption,
  } from "./time-range-utils";

  export let metricsDefId: string;

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  let selectedTimeGrain: TimeGrain;
  $: selectedTimeGrain = $metricsExplorer?.selectedTimeGrain;

  let selectedTimeRangeName: TimeRangeName;
  $: selectedTimeRangeName = $metricsExplorer?.selectedTimeRange?.name;

  let selectableTimeGrains: TimeGrainOption[];
  // query the `/meta` endpoint to get the full time range of the dataset
  let queryKey = getMetricViewMetaQueryKey(metricsDefId);
  const queryResult = useQuery<MetricViewMetaResponse, Error>(queryKey, () =>
    getMetricViewMetadata(metricsDefId)
  );
  $: {
    queryKey = getMetricViewMetaQueryKey(metricsDefId);
    queryResult.setOptions(queryKey, () => getMetricViewMetadata(metricsDefId));
  }
  $: if (selectedTimeRangeName && $queryResult.data?.timeDimension?.timeRange) {
    selectableTimeGrains = getSelectableTimeGrains(
      selectedTimeRangeName,
      $queryResult.data.timeDimension.timeRange
    );
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
    store.dispatch(updateSelectedTimeGrainApi({ metricsDefId, timeGrain }));
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
