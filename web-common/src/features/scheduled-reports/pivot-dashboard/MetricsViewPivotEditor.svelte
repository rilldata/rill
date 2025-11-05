<script lang="ts">
  import type { PivotDataStoreConfig } from "@rilldata/web-common/features/dashboards/pivot/types.ts";
  import {
    getAvailableDimensions,
    getAvailableMeasures,
  } from "@rilldata/web-common/features/scheduled-reports/pivot-dashboard/pivot-data-config.ts";
  import { PivotStore } from "@rilldata/web-common/features/scheduled-reports/pivot-dashboard/pivot-store.ts";
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import { createPivotDataStore } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store.ts";
  import PivotEmpty from "@rilldata/web-common/features/dashboards/pivot/PivotEmpty.svelte";
  import PivotHeader from "@rilldata/web-common/features/dashboards/pivot/PivotHeader.svelte";
  import PivotSidebar from "@rilldata/web-common/features/dashboards/pivot/PivotSidebar.svelte";
  import PivotTable from "@rilldata/web-common/features/dashboards/pivot/PivotTable.svelte";
  import PivotToolbar from "@rilldata/web-common/features/dashboards/pivot/PivotToolbar.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { type Readable, readable } from "svelte/store";

  export let metricsViewName: string;
  export let pivotStore: PivotStore;
  export let pivotConfigStore: Readable<PivotDataStoreConfig>;

  $: ({
    state,
    columnMeasures,
    columnDimensions,

    setRows,
    setColumns,
    addField,
    setTableMode,
    setExpanded,
    setSort,
    setRowPage,
    setActiveCell,
    removeActiveCell,
  } = pivotStore);
  $: ({ time } = $pivotConfigStore);

  $: timeControlsForPillActions = {
    timeStart: time?.timeStart,
    timeEnd: time?.timeEnd,
    minTimeGrain: time?.minTimeGrain,
  };

  let showPanels = true;

  $: pivotDataStore = createPivotDataStore(
    {
      metricsViewName: readable(metricsViewName),
      enabled: true,
      queryClient,
    },
    pivotConfigStore,
  );

  $: availableMeasures = getAvailableMeasures(pivotStore, pivotConfigStore);
  $: availableDimensions = getAvailableDimensions(pivotStore, pivotConfigStore);

  $: ({ isFetching, assembled } = $pivotDataStore);

  $: hasColumnAndNoMeasure =
    $columnDimensions.length > 0 && $columnMeasures.length === 0;
</script>

<div class="layout" class:h-full={!$dynamicHeight}>
  {#if showPanels}
    <PivotSidebar
      pivotState={$state}
      measures={$availableMeasures}
      dimensions={$availableDimensions}
      {timeControlsForPillActions}
      {addField}
    />
  {/if}
  <div
    class="flex flex-col overflow-hidden"
    class:w-full={$dynamicHeight}
    class:size-full={!$dynamicHeight}
  >
    {#if showPanels}
      <PivotHeader
        pivotState={$state}
        measures={$availableMeasures}
        dimensions={$availableDimensions}
        {timeControlsForPillActions}
        {setRows}
        {setColumns}
        {addField}
      />
    {/if}
    <div
      class="content"
      class:size-full={!$dynamicHeight}
      role="presentation"
      on:mousedown|self={removeActiveCell}
    >
      <PivotToolbar
        pivotState={$state}
        {setTableMode}
        collapseAll={() => setExpanded({})}
        {isFetching}
        bind:showPanels
      />

      {#if $pivotDataStore?.error?.length}
        <PivotError errors={$pivotDataStore.error} />
      {:else if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
        <PivotEmpty {assembled} {isFetching} {hasColumnAndNoMeasure} />
      {:else}
        <PivotTable
          {pivotDataStore}
          overscan={60}
          config={pivotConfigStore}
          pivotState={state}
          setPivotExpanded={setExpanded}
          setPivotSort={setSort}
          setPivotRowPage={setRowPage}
          setPivotActiveCell={setActiveCell}
        />
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .layout {
    @apply flex box-border overflow-hidden;
  }

  .content {
    @apply flex w-full flex-col bg-gray-50 overflow-hidden;
    @apply p-2 gap-y-2;
  }
</style>
