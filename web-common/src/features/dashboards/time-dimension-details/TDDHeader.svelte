<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import SearchableFilterChip from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterChip.svelte";
  import { ephemeralMeasureDialog } from "@rilldata/web-common/features/dashboards/ephemeral-measures/dialog-store";
  import { precisionStateFor } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
  import { exploreExactSurface } from "@rilldata/web-common/features/dashboards/exact-measures/explore-surfaces";
  import PrecisionChip from "@rilldata/web-common/features/dashboards/exact-measures/PrecisionChip.svelte";
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
  import ExcludeButton from "../toolbars/ExcludeButton.svelte";
  import SearchButton from "../toolbars/SearchButton.svelte";
  import SelectAllButton from "../toolbars/SelectAllButton.svelte";
  import StartPivotButton from "../toolbars/StartPivotButton.svelte";
  import { getTDDExportQuery } from "./tdd-export";
  import type { TDDComparison } from "./types";
  import { translateV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
  import AggregationGrainSelector from "@rilldata/web-common/features/dashboards/time-controls/AggregationGrainSelector.svelte";

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

  const { exports } = featureFlags;
  const stateManagers = getStateManagers();

  const {
    selectors: {
      measures: { measureLabel, allMeasures },
      dimensions: { getDimensionDisplayName },
      exactMeasures: { exactMeasureContext },
    },
    actions: {
      exactMeasures: { toggleExactMeasure },
    },
    dashboardStore,
    expressionFilterManager,
    timeFilterManager,
  } = stateManagers;

  const ephemeralDefsByName = $derived(
    new Map(
      ($dashboardStore?.ephemeralMeasures ?? []).map((def) => [def.name, def]),
    ),
  );

  // Precision of the expanded measure; null when it has no exact variant.
  const expandedMeasurePrecision = $derived(
    precisionStateFor(
      expandedMeasureName,
      $exactMeasureContext(exploreExactSurface.tdd),
    ),
  );

  const selectableMeasures = $derived(
    $allMeasures
      .filter((m) => m.name !== undefined || m.displayName !== undefined)
      .map((m) => {
        const def = ephemeralDefsByName.get(m.name || "");
        // Note: undefined values are filtered out above, so the
        // empty string fallback is unreachable.
        return {
          name: m.name || "",
          label: m.displayName || "",
          ...(def ? { description: def.expression, ephemeral: true } : {}),
        };
      }),
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
  // The precision chip has its own tooltip; the measure chip's tooltip hides while it is hovered.
  let precisionChipHovered = $state(false);
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
        title: translateV1TimeGrain(dashboardGrain),
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
        {#if timeFilterManager.timeGrain}
          <AggregationGrainSelector {timeFilterManager} type="tdd" />
        {/if}
        <SearchableFilterChip
          label={selectedMeasureLabel}
          onSelect={switchMeasure}
          onEditItem={(name) => {
            const def = ephemeralDefsByName.get(name);
            if (def) ephemeralMeasureDialog.set({ def });
          }}
          selectableItems={selectableMeasures}
          selectedItems={[expandedMeasureName]}
          tooltipText="Choose a measure to display"
          suppressTooltip={precisionChipHovered}
        >
          <svelte:fragment slot="additional-label">
            {#if ephemeralDefsByName.has(expandedMeasureName)}
              <span class="flex-none text-[10px] font-semibold italic">ƒx</span>
            {/if}
            {#if expandedMeasurePrecision}
              <span class="flex-none inline-flex">
                <PrecisionChip
                  state={expandedMeasurePrecision}
                  measureLabel={selectedMeasureLabel}
                  loading={expandedMeasurePrecision !== "approx" && isFetching}
                  onToggle={() =>
                    toggleExactMeasure(
                      exploreExactSurface.tdd,
                      expandedMeasureName,
                    )}
                  onHoverChange={(hovered) => (precisionChipHovered = hovered)}
                />
              </span>
            {/if}
          </svelte:fragment>
        </SearchableFilterChip>
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
          includeScheduledReport
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
