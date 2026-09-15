<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
  import ExportMenu from "../../exports/ExportMenu.svelte";
  import { featureFlags } from "../../feature-flags";
  import { PivotChipType } from "../pivot/types";
  import TimeGrainSelector from "../time-controls/TimeGrainSelector.svelte";
  import ExcludeButton from "../toolbars/ExcludeButton.svelte";
  import SearchButton from "../toolbars/SearchButton.svelte";
  import SelectAllButton from "../toolbars/SelectAllButton.svelte";
  import StartPivotButton from "../toolbars/StartPivotButton.svelte";
  import { getTDDExportQuery } from "./tdd-export";
  import type { TDDComparison } from "./types";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";

  interface Props {
    exploreName: string;
    dimensionName: string;
    isFetching?: boolean;
    comparing: TDDComparison | undefined;
    areAllTableRowsSelected?: boolean;
    isRowsEmpty?: boolean;
    expandedMeasureName: string;
    onToggleSearchItems: () => void;
    hideStartPivotButton?: boolean;
  }

  let {
    exploreName,
    dimensionName,
    isFetching = false,
    comparing,
    areAllTableRowsSelected = false,
    isRowsEmpty = false,
    expandedMeasureName,
    onToggleSearchItems,
    hideStartPivotButton = false,
  }: Props = $props();

  const { adminServer, exports } = featureFlags;
  const stateManagers = getStateManagers();

  const {
    selectors: {
      measures: { measureLabel, allMeasures },
      dimensions: { getDimensionDisplayName },
    },
    dashboardStore,
    dashboardConfigProvider,
    expressionFilterManager,
    timeFilterManager,
  } = stateManagers;

  let { metricsViewsProvider } = $derived(dashboardConfigProvider);
  let { smallestTimeGrain } = $derived(metricsViewsProvider);

  let { timeGrain, timeStart, timeEnd } = $derived(timeFilterManager);

  const selectableMeasures = $derived(
    $allMeasures
      .filter((m) => m.name !== undefined || m.displayName !== undefined)
      .map((m) =>
        // Note: undefined values are filtered out above, so the
        // empty string fallback is unreachable.
        ({
          name: m.name || "",
          label: m.displayName || "",
        }),
      ),
  );

  const selectedMeasureLabel = $derived(
    $allMeasures.find((m) => m.name === expandedMeasureName)?.displayName ||
      expandedMeasureName,
  );

  const excludeMode = $derived(
    expressionFilterManager.sortedFilterManagers.dimensions.find(
      (dfm) => dfm.name === dimensionName,
    )?.exclude ?? false,
  );

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
    expressionFilterManager.dimensionFilterAction(
      dimensionName,
      (dimensionManager) => dimensionManager.toggleExclude(),
    );
  }

  function switchMeasure(measureName: string) {
    metricsExplorerStore.setExpandedMeasureName(exploreName, measureName);
  }

  let showReplacePivotModal = $state(false);
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
    if (!timeGrain || !expandedMeasureName) return;

    const dateUnit = V1TimeGrainToDateTimeUnit[timeGrain];
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
        id: timeGrain,
        title: dateUnit,
        type: PivotChipType.Time,
      },
      {
        id: expandedMeasureName,
        title: $measureLabel(expandedMeasureName),
        type: PivotChipType.Measure,
      },
    ]);
  }
</script>

<div class="tdd-header bg-surface-background">
  <div class="flex gap-x-6 items-center font-normal text-fg-secondary">
    <div class="flex items-center gap-x-4">
      <div class="flex items-center gap-x-1">
        <Row size="16px" />
        {m.dashboard_rows()}
      </div>

      <ComparisonSelector {exploreName} />
    </div>

    <div class="flex items-center gap-x-4 pl-2">
      <div class="flex items-center gap-x-1">
        <Column size="16px" />
        {m.dashboard_columns()}
      </div>
      <div class="flex items-center gap-x-2">
        <TimeGrainSelector
          tdd
          activeTimeGrain={timeGrain}
          onTimeGrainSelect={(grain) => timeFilterManager.onSelectGrain(grain)}
          {timeStart}
          {timeEnd}
          minTimeGrain={smallestTimeGrain}
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
          label={m.dashboard_export_table_data()}
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
