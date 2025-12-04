<script lang="ts">
  import { getPivotExportQuery } from "@rilldata/web-common/features/dashboards/pivot/pivot-export.ts";
  import ExportMenu from "@rilldata/web-common/features/exports/ExportMenu.svelte";
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { derived } from "svelte/store";
  import { useTimeControlStore } from "web-common/src/features/dashboards/time-controls/time-control-store.ts";
  import { getPivotConfig } from "./pivot-data-config";
  import { usePivotForExplore } from "./pivot-data-store";
  import PivotEmpty from "./PivotEmpty.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotTable from "./PivotTable.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";

  export let isEmbedded: boolean = false;

  const stateManagers = getStateManagers();
  const {
    exploreName,
    dashboardStore,
    selectors: {
      pivot: { columns, measures, dimensions },
    },
    timeRangeSummaryStore,
  } = stateManagers;

  const { adminServer, exports } = featureFlags;

  const timeControlsStore = useTimeControlStore(stateManagers);
  $: timeControlsForPillActions = {
    timeStart: $timeControlsStore.timeStart,
    timeEnd: $timeControlsStore.timeEnd,
    minTimeGrain: $timeControlsStore.minTimeGrain,
  };

  $: exploreHasTimeDimension = !!$timeRangeSummaryStore.data;

  const { cloudDataViewer, readOnly } = featureFlags;

  $: isRillDeveloper = $readOnly === false;
  $: canShowDataViewer = Boolean($cloudDataViewer || isRillDeveloper);

  const pivotExploreState = derived(dashboardStore, (dashboard) => {
    return dashboard?.pivot;
  });

  let showPanels = true;

  $: pivotDataStore = usePivotForExplore(stateManagers);
  $: pivotConfig = getPivotConfig(stateManagers);

  $: ({ isFetching, assembled } = $pivotDataStore);

  $: hasColumnAndNoMeasure =
    $columns.dimension.length > 0 && $columns.measure.length === 0;

  function removeActiveCell() {
    if (!$dashboardStore.pivot.activeCell) return;
    metricsExplorerStore.removePivotActiveCell($exploreName);
  }
</script>

<div class="layout" class:h-full={!$dynamicHeight}>
  {#if showPanels}
    <PivotSidebar
      pivotState={$dashboardStore.pivot}
      measures={$measures}
      dimensions={$dimensions}
      {timeControlsForPillActions}
    />
  {/if}
  <div
    class="flex flex-col overflow-hidden"
    class:w-full={$dynamicHeight}
    class:size-full={!$dynamicHeight}
  >
    {#if showPanels}
      <PivotHeader
        pivotState={$dashboardStore.pivot}
        setRows={(rows) =>
          metricsExplorerStore.setPivotRows($exploreName, rows)}
        setColumns={(columns) =>
          metricsExplorerStore.setPivotColumns($exploreName, columns)}
      />
    {/if}
    <div
      class="content"
      class:size-full={!$dynamicHeight}
      role="presentation"
      on:mousedown|self={removeActiveCell}
    >
      <PivotToolbar
        pivotState={$dashboardStore.pivot}
        setTableMode={(tableMode, rows, columns) =>
          metricsExplorerStore.setPivotTableMode(
            $exploreName,
            tableMode,
            rows,
            columns,
          )}
        collapseAll={() =>
          metricsExplorerStore.setPivotExpanded($exploreName, {})}
        {isFetching}
        bind:showPanels
      >
        <svelte:fragment slot="export-menu">
          {#if $exports}
            <ExportMenu
              label="Export pivot data"
              includeScheduledReport={$adminServer && exploreHasTimeDimension}
              getQuery={(isScheduled) =>
                getPivotExportQuery(stateManagers, isScheduled)}
              exploreName={$exploreName}
            />
          {/if}
        </svelte:fragment>
      </PivotToolbar>

      {#if $pivotDataStore?.error?.length}
        <PivotError errors={$pivotDataStore.error} />
      {:else if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
        <PivotEmpty
          {assembled}
          {isFetching}
          {hasColumnAndNoMeasure}
          {isEmbedded}
        />
      {:else}
        <PivotTable
          {pivotDataStore}
          overscan={60}
          config={pivotConfig}
          pivotState={pivotExploreState}
          setPivotExpanded={(expanded) =>
            metricsExplorerStore.setPivotExpanded($exploreName, expanded)}
          setPivotSort={(sorting) =>
            metricsExplorerStore.setPivotSort($exploreName, sorting)}
          setPivotRowPage={(page) =>
            metricsExplorerStore.setPivotRowPage($exploreName, page)}
          {canShowDataViewer}
          setPivotActiveCell={(rowId, columnId) =>
            metricsExplorerStore.setPivotActiveCell(
              $exploreName,
              rowId,
              columnId,
            )}
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
