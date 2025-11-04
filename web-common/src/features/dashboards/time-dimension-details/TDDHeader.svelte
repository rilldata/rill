<script lang="ts">
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import SearchableFilterChip from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterChip.svelte";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import ReplacePivotDialog from "@rilldata/web-common/features/dashboards/pivot/ReplacePivotDialog.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    dimensionSearchText,
    metricsExplorerStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import ComparisonSelector from "@rilldata/web-common/features/dashboards/time-controls/ComparisonSelector.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import type {
    DashboardTimeControls,
    TimeGrain,
    TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import ExportMenu from "../../exports/ExportMenu.svelte";
  import { featureFlags } from "../../feature-flags";
  import { PivotChipType } from "../pivot/types";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import TimeGrainSelector from "../time-controls/TimeGrainSelector.svelte";
  import ExcludeButton from "../toolbars/ExcludeButton.svelte";
  import SearchButton from "../toolbars/SearchButton.svelte";
  import SelectAllButton from "../toolbars/SelectAllButton.svelte";
  import StartPivotButton from "../toolbars/StartPivotButton.svelte";
  import { getTDDExportQuery } from "./tdd-export";
  import type { TDDComparison } from "./types";

  export let exploreName: string;
  export let dimensionName: string;
  export let isFetching = false;
  export let comparing: TDDComparison | undefined;
  export let areAllTableRowsSelected = false;
  export let isRowsEmpty = false;
  export let expandedMeasureName: string;
  export let onToggleSearchItems: () => void;
  export let hideStartPivotButton = false;

  const { adminServer, exports } = featureFlags;
  const stateManagers = getStateManagers();

  const {
    selectors: {
      measures: { measureLabel, allMeasures },
      dimensions: { getDimensionDisplayName },
    },
    actions: {
      dimensionsFilter: { toggleDimensionFilterMode },
    },
    dashboardStore,
    validSpecStore,
  } = stateManagers;

  $: selectableMeasures = $allMeasures
    .filter((m) => m.name !== undefined || m.displayName !== undefined)
    .map((m) =>
      // Note: undefined values are filtered out above, so the
      // empty string fallback is unreachable.
      ({
        name: m.name || "",
        label: m.displayName || "",
      }),
    );

  $: selectedMeasureLabel =
    $allMeasures.find((m) => m.name === expandedMeasureName)?.displayName ||
    expandedMeasureName;

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  function closeSearchBar() {
    dimensionSearchText.set("");
  }

  function onSubmit() {
    if (!areAllTableRowsSelected) {
      onToggleSearchItems();
      closeSearchBar();
    }
  }

  function toggleFilterMode() {
    toggleDimensionFilterMode(dimensionName);
  }

  function switchMeasure(measureName: string) {
    metricsExplorerStore.setExpandedMeasureName(exploreName, measureName);
  }

  let showReplacePivotModal = false;
  function startPivotForTDD() {
    const pivot = $dashboardStore?.pivot;

    const pivotColumns = splitPivotChips(pivot.columns);
    if (
      pivot.rows.length ||
      pivotColumns.measure.length ||
      pivotColumns.dimension.length
    ) {
      showReplacePivotModal = true;
    } else {
      createPivot();
    }
  }

  function createPivot() {
    showReplacePivotModal = false;
    const dashboardGrain = $dashboardStore?.selectedTimeRange?.interval;
    if (!dashboardGrain || !expandedMeasureName) return;

    const timeGrain: TimeGrain = TIME_GRAIN[dashboardGrain];
    const rowDimensions = dimensionName
      ? [
          {
            id: dimensionName,
            title: $getDimensionDisplayName(dimensionName),
            type: PivotChipType.Dimension,
          },
        ]
      : [];
    metricsExplorerStore.createPivot(exploreName, rowDimensions, [
      {
        id: dashboardGrain,
        title: timeGrain.label,
        type: PivotChipType.Time,
      },
      {
        id: expandedMeasureName,
        title: $measureLabel(expandedMeasureName),
        type: PivotChipType.Measure,
      },
    ]);
  }

  const timeControlsStore = useTimeControlStore(stateManagers);

  $: ({ minTimeGrain, timeStart, timeEnd, selectedTimeRange } =
    $timeControlsStore);

  $: activeTimeGrain = selectedTimeRange?.interval;

  $: baseTimeRange = selectedTimeRange?.start &&
    selectedTimeRange?.end && {
      name: selectedTimeRange?.name,
      start: selectedTimeRange.start,
      end: selectedTimeRange.end,
    };

  function onTimeGrainSelect(timeGrain: V1TimeGrain) {
    if (baseTimeRange) {
      makeTimeSeriesTimeRangeAndUpdateAppState(
        baseTimeRange,
        timeGrain,
        $dashboardStore?.selectedComparisonTimeRange,
      );
    }
  }

  function makeTimeSeriesTimeRangeAndUpdateAppState(
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    /** we should only reset the comparison range when the user has explicitly chosen a new
     * time range. Otherwise, the current comparison state should continue to be the
     * source of truth.
     */
    comparisonTimeRange: DashboardTimeControls | undefined,
  ) {
    metricsExplorerStore.selectTimeRange(
      exploreName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
      $validSpecStore.data?.metricsView ?? {},
    );
  }
</script>

<div class="tdd-header">
  <div class="flex gap-x-6 items-center font-normal text-gray-500">
    <div class="flex items-center gap-x-4">
      <div class="flex items-center gap-x-1">
        <Row size="16px" /> Rows
      </div>

      <ComparisonSelector {exploreName} />
    </div>

    <div class="flex items-center gap-x-4 pl-2">
      <div class="flex items-center gap-x-1">
        <Column size="16px" /> Columns
      </div>
      <div class="flex items-center gap-x-2">
        <TimeGrainSelector
          tdd
          {activeTimeGrain}
          {onTimeGrainSelect}
          {timeStart}
          {timeEnd}
          {minTimeGrain}
        />
        <SearchableFilterChip
          label={selectedMeasureLabel}
          onSelect={switchMeasure}
          selectableItems={selectableMeasures}
          selectedItems={[expandedMeasureName]}
          tooltipText="Choose a measure to display"
        />
      </div>
    </div>

    {#if isFetching}
      <DelayedSpinner isLoading={isFetching} size="18px" />
    {/if}
  </div>

  {#if comparing === "dimension"}
    <div class="flex items-center gap-x-1" style:cursor="pointer">
      <SelectAllButton
        {areAllTableRowsSelected}
        disabled={isRowsEmpty}
        {onToggleSearchItems}
      />

      <ExcludeButton {excludeMode} onClick={toggleFilterMode} />

      <SearchButton
        bind:value={$dimensionSearchText}
        {onSubmit}
        onClose={closeSearchBar}
      />

      {#if $exports}
        <ExportMenu
          label="Export table data"
          includeScheduledReport={$adminServer}
          getQuery={(isScheduled) =>
            getTDDExportQuery(stateManagers, isScheduled)}
          {exploreName}
        />
      {/if}

      {#if !hideStartPivotButton}
        <StartPivotButton onClick={startPivotForTDD} />
      {/if}
    </div>
  {/if}
</div>

<ReplacePivotDialog
  open={showReplacePivotModal}
  onCancel={() => {
    showReplacePivotModal = false;
  }}
  onReplace={createPivot}
/>

<style lang="postcss">
  .tdd-header {
    @apply grid justify-between grid-flow-col items-center py-2 px-4 h-11;
  }
</style>
