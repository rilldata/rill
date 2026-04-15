<script lang="ts">
  import { getPivotExportQuery } from "@rilldata/web-common/features/dashboards/pivot/pivot-export.ts";
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import ExportMenu from "@rilldata/web-common/features/exports/ExportMenu.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import { derived } from "svelte/store";
  import { useTimeControlStore } from "web-common/src/features/dashboards/time-controls/time-control-store.ts";
  import { getPivotConfig } from "./pivot-data-config";
  import { usePivotForExplore } from "./pivot-data-store";
  import PivotEmpty from "./PivotEmpty.svelte";
  import PivotHeader from "./PivotHeader.svelte";
  import type { PivotChipData } from "./types";
  import PivotSidebar from "./PivotSidebar.svelte";
  import PivotTable from "./PivotTable.svelte";
  import PivotToolbar from "./PivotToolbar.svelte";

  export let isEmbedded: boolean = false;

  const stateManagers = getStateManagers();
  const {
    exploreName,
    dashboardStore,
    validSpecStore,
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

  // Build a description lookup from the metricsView so pivot chips get
  // descriptions even when the metricsView loads after the pivot state
  // is deserialized (e.g. public URL first load).
  $: descriptionMap = new Map<string, string | undefined>([
    ...($validSpecStore.data?.metricsView?.dimensions ?? []).map(
      (d) => [d.name, d.description] as [string, string | undefined],
    ),
    ...($validSpecStore.data?.metricsView?.measures ?? []).map(
      (m) => [m.name, m.description] as [string, string | undefined],
    ),
  ]);

  function enrichDescriptions(chips: PivotChipData[]): PivotChipData[] {
    return chips.map((chip) =>
      chip.description ? chip : { ...chip, description: descriptionMap.get(chip.id) },
    );
  }

  $: enrichedPivotState = {
    ...$dashboardStore.pivot,
    rows: enrichDescriptions($dashboardStore.pivot.rows),
    columns: enrichDescriptions($dashboardStore.pivot.columns),
  };

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
      pivotState={enrichedPivotState}
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
        pivotState={enrichedPivotState}
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
      onmousedown={(e) => {
        if (e.target === e.currentTarget) removeActiveCell();
      }}
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
        setRowLimit={(limit) =>
          metricsExplorerStore.setPivotRowLimit($exploreName, limit)}
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
          setPivotOutermostRowLimit={(limit) =>
            metricsExplorerStore.setPivotOutermostRowLimit($exploreName, limit)}
          setPivotRowLimitForExpanded={(expandIndex, limit) =>
            metricsExplorerStore.setPivotRowLimitForExpandedRow(
              $exploreName,
              expandIndex,
              limit,
            )}
        />
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .layout {
    @apply flex box-border overflow-hidden size-full;
  }

  .content {
    @apply flex w-full flex-col bg-surface-subtle overflow-hidden;
    @apply p-2 gap-y-2;
  }
</style>
