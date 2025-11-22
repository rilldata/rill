<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import LeaderboardAdvancedActions from "@rilldata/web-common/components/menu/LeaderboardAdvancedActions.svelte";
  import ReplacePivotDialog from "@rilldata/web-common/features/dashboards/pivot/ReplacePivotDialog.svelte";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { onDestroy } from "svelte";
  import ExportMenu from "../../exports/ExportMenu.svelte";
  import { SortType } from "../proto-state/derived-types";
  import { getStateManagers } from "../state-managers/state-managers";
  import ExcludeButton from "../toolbars/ExcludeButton.svelte";
  import SearchButton from "../toolbars/SearchButton.svelte";
  import SelectAllButton from "../toolbars/SelectAllButton.svelte";
  import StartPivotButton from "../toolbars/StartPivotButton.svelte";
  import { getDimensionTableExportQuery } from "./dimension-table-export";

  export let dimensionName: string;
  export let areAllTableRowsSelected = false;
  export let isRowsEmpty = true;
  export let searchText: string;
  export let hideStartPivotButton = false;
  export let onToggleSearchItems: () => void;

  const stateManagers = getStateManagers();
  const {
    selectors: {
      sorting: { sortedByDimensionValue },
      dimensions: { getDimensionDisplayName },
      dimensionFilters: { isFilterExcludeMode },
      measures: { visibleMeasures },
      leaderboard: { leaderboardMeasureNames },
    },
    actions: {
      sorting: { toggleSort },
      dimensions: { setPrimaryDimension },
      dimensionsFilter: { toggleDimensionFilterMode },
      leaderboard: {
        toggleLeaderboardShowContextForAllMeasures,
        setLeaderboardSortByMeasureName,
      },
    },
    timeRangeSummaryStore,
    dashboardStore,
    exploreName,
  } = stateManagers;

  const { adminServer, exports } = featureFlags;

  $: exploreHasTimeDimension = !!$timeRangeSummaryStore.data;

  $: excludeMode = $isFilterExcludeMode(dimensionName);

  let isLeaderboardActionsOpen = false;

  function resetSearchText() {
    searchText = "";
  }

  function onSubmit() {
    if (!areAllTableRowsSelected) {
      onToggleSearchItems();
      resetSearchText();
    }
  }

  const goBackToLeaderboard = () => {
    if ($sortedByDimensionValue) {
      toggleSort(SortType.VALUE);
    }

    // Reset expanded dimension
    setPrimaryDimension("");

    // If user previously sorted by a measure that is not in the leaderboard measure names in expanded view,
    // we need to set a new sort measure from the available leaderboard measures
    if (
      !$leaderboardMeasureNames.includes(
        $dashboardStore.leaderboardSortByMeasureName,
      )
    ) {
      setLeaderboardSortByMeasureName($leaderboardMeasureNames[0]);
    }
  };
  function toggleFilterMode() {
    toggleDimensionFilterMode(dimensionName);
  }

  let showReplacePivotModal = false;
  function startPivotForDimensionTable() {
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

    const rowDimensions = dimensionName
      ? [
          {
            id: dimensionName,
            title: $getDimensionDisplayName(dimensionName),
            type: PivotChipType.Dimension,
          },
        ]
      : [];

    const measures = $visibleMeasures
      .filter((m) => m.name !== undefined)
      .map((m) => {
        return {
          id: m.name as string,
          title: m.displayName || (m.name as string),
          type: PivotChipType.Measure,
        };
      });

    metricsExplorerStore.createPivot($exploreName, rowDimensions, measures);
  }

  onDestroy(() => {
    searchText = "";
  });
</script>

<div class="flex-none overflow-x-auto" style="height: 40px">
  <div class="flex justify-start items-center p-1 h-7 gap-x-2">
    <Button
      type="link"
      forcedStyle="padding: 0; gap: 4px;"
      onClick={() => goBackToLeaderboard()}
    >
      <Back size="16px" />
      <span>All Dimensions</span>
    </Button>

    <div class="shrink-0 flex items-center gap-x-1">
      <SelectAllButton
        {areAllTableRowsSelected}
        disabled={isRowsEmpty}
        {onToggleSearchItems}
      />

      <ExcludeButton {excludeMode} onClick={toggleFilterMode} />

      <SearchButton
        bind:value={searchText}
        {onSubmit}
        onClose={resetSearchText}
      />

      {#if $exports}
        <ExportMenu
          label="Export dimension table data"
          includeScheduledReport={$adminServer && exploreHasTimeDimension}
          getQuery={(isScheduled) =>
            getDimensionTableExportQuery(stateManagers, isScheduled)}
          exploreName={$exploreName}
        />
      {/if}

      {#if !hideStartPivotButton}
        <StartPivotButton onClick={startPivotForDimensionTable} />
      {/if}

      <LeaderboardAdvancedActions
        isOpen={isLeaderboardActionsOpen}
        toggle={toggleLeaderboardShowContextForAllMeasures}
      />
    </div>
  </div>
</div>

<ReplacePivotDialog
  open={showReplacePivotModal}
  onCancel={() => {
    showReplacePivotModal = false;
  }}
  onReplace={createPivot}
/>
