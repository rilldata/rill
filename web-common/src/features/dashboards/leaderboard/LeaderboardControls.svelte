<script lang="ts">
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/DashboardVisibilityDropdown.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import LeaderboardMeasureCountSelector from "@rilldata/web-common/components/menu/LeaderboardMeasureCountSelector.svelte";
  import LeaderboardActiveMeasureDropdown from "@rilldata/web-common/components/menu/LeaderboardActiveMeasureDropdown.svelte";
  import { SortType } from "../proto-state/derived-types";
  import DashboardDraggableSelector from "@rilldata/web-common/components/menu/DashboardDraggableSelector.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";

  export let exploreName: string;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: {
        leaderboardMeasureCount,
        leaderboardMeasureName,
        getMeasureByName,
        visibleMeasures,
        allMeasures,
      },
      dimensions: { visibleDimensions, allDimensions },
      sorting: { sortByMeasure },
    },
    actions: {
      contextColumn: { setContextColumn },
      sorting: { toggleSort, setSortDescending },
      setLeaderboardMeasureCount,
      dimensions: { setDimensionVisibility, toggleDimensionVisibility },
      setLeaderboardMeasureName,
    },
  } = StateManagers;

  const {
    leaderboardMeasureCount: leaderboardMeasureCountFeatureFlag,
    reorderMeasuresDimensions,
  } = featureFlags;

  let active = false;

  $: measures = getSimpleMeasures($visibleMeasures);

  $: metricsExplorer = $metricsExplorerStore.entities[exploreName];

  $: activeLeaderboardMeasure = $getMeasureByName($leaderboardMeasureName);

  $: validPercentOfTotal = leaderboardMeasureCountFeatureFlag
    ? $visibleMeasures.some((measure) => measure.validPercentOfTotal)
    : activeLeaderboardMeasure?.validPercentOfTotal || false;

  $: allMeasureNames = $allMeasures.map(({ name }) => name).filter(isDefined);

  $: visibleMeasureNames = $visibleMeasures
    .map(({ name }) => name)
    .filter(isDefined);

  $: visibleDimensionsNames = $visibleDimensions
    .map(({ name }) => name)
    .filter(isDefined);
  $: allDimensionNames = $allDimensions
    .map(({ name }) => name)
    .filter(isDefined);

  // if the percent of total is currently being shown,
  // but it is not valid for this measure, then turn it off
  $: if (
    !validPercentOfTotal &&
    metricsExplorer?.leaderboardContextColumn ===
      LeaderboardContextColumn.PERCENT
  ) {
    setContextColumn(LeaderboardContextColumn.HIDDEN);
  }

  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }
</script>

<div>
  {#if measures.length}
    <div
      class="flex flex-row items-center ui-copy-muted gap-x-1"
      style:max-width="768px"
    >
      {#if $reorderMeasuresDimensions}
        <DashboardDraggableSelector
          type="dimension"
          onSelectedChange={(items) =>
            setDimensionVisibility(items, allDimensionNames)}
          allItems={$allDimensions}
          selectedItems={visibleDimensionsNames}
        />
      {:else}
        <DashboardVisibilityDropdown
          category="Dimensions"
          tooltipText="Choose dimensions to display"
          onSelect={(name) =>
            toggleDimensionVisibility(allDimensionNames, name)}
          selectableItems={$allDimensions.map(({ name, displayName }) => ({
            name: name || "",
            label: displayName || name || "",
          }))}
          selectedItems={visibleDimensionsNames}
          onToggleSelectAll={() => {
            toggleDimensionVisibility(allDimensionNames);
          }}
        />
      {/if}

      {#if $leaderboardMeasureCountFeatureFlag}
        <LeaderboardMeasureCountSelector
          measures={$visibleMeasures}
          count={$leaderboardMeasureCount}
          onMeasureCountChange={(count) => {
            setLeaderboardMeasureCount(count);
          }}
          resetSort={() => {
            // Fallback to the first visible measure if sort_by measure is not in the context
            toggleSort(SortType.VALUE, $visibleMeasures[0].name);
            setSortDescending();
          }}
          sortByMeasure={$sortByMeasure}
        />
      {:else}
        <LeaderboardActiveMeasureDropdown
          leaderboardMeasureName={$leaderboardMeasureName}
          {setLeaderboardMeasureName}
          {measures}
          {activeLeaderboardMeasure}
        />
      {/if}
    </div>
  {/if}
</div>
