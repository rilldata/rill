<script lang="ts">
  import DashboardMetricsDraggableList from "@rilldata/web-common/components/menu/DashboardMetricsDraggableList.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { filterOutSomeAdvancedAggregationMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures.ts";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import LeaderboardMeasureNamesDropdown from "@rilldata/web-common/components/menu/LeaderboardMeasureNamesDropdown.svelte";
  import LeaderboardAdvancedActions from "@rilldata/web-common/components/menu/LeaderboardAdvancedActions.svelte";

  export let exploreName: string;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: { getMeasureByName, visibleMeasures },
      leaderboard: { leaderboardSortByMeasureName, leaderboardMeasureNames },
      dimensions: { visibleDimensions, allDimensions },
    },
    actions: {
      contextColumn: { setContextColumn },
      dimensions: { setDimensionVisibility },
      leaderboard: {
        setLeaderboardSortByMeasureName,
        setLeaderboardMeasureNames,
        toggleLeaderboardShowContextForAllMeasures,
      },
    },
    validSpecStore,
  } = StateManagers;

  let isLeaderboardActionsOpen = false;

  $: exploreState = $metricsExplorerStore.entities[exploreName];
  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};

  $: filteredMeasures = filterOutSomeAdvancedAggregationMeasures(
    exploreState,
    metricsViewSpec,
    $visibleMeasures,
    false,
  );

  $: activeLeaderboardMeasure = $getMeasureByName(
    $leaderboardSortByMeasureName,
  );
  $: validPercentOfTotal =
    activeLeaderboardMeasure?.validPercentOfTotal || false;

  $: visibleDimensionsNames = $visibleDimensions
    .map(({ name }) => name)
    .filter(isDefined);
  $: allDimensionNames = $allDimensions
    .map(({ name }) => name)
    .filter(isDefined);

  $: isMultiSelectEnabled = $leaderboardMeasureNames.length > 1;

  // if the percent of total is currently being shown,
  // but it is not valid for this measure, then turn it off
  $: if (
    !validPercentOfTotal &&
    exploreState?.leaderboardContextColumn === LeaderboardContextColumn.PERCENT
  ) {
    setContextColumn(LeaderboardContextColumn.HIDDEN);
  }

  function isDefined(value: string | undefined): value is string {
    return value !== undefined;
  }
</script>

<div
  class="flex flex-row items-center ui-copy-muted gap-x-1"
  style:max-width="768px"
>
  <DashboardMetricsDraggableList
    type="dimension"
    onSelectedChange={(items) =>
      setDimensionVisibility(items, allDimensionNames)}
    allItems={$allDimensions}
    selectedItems={visibleDimensionsNames}
  />
  <LeaderboardMeasureNamesDropdown
    visibleMeasures={filteredMeasures}
    leaderboardSortByMeasureName={$leaderboardSortByMeasureName}
    selectedMeasureNames={$leaderboardMeasureNames}
    {setLeaderboardMeasureNames}
    {setLeaderboardSortByMeasureName}
  />
  {#if isMultiSelectEnabled}
    <LeaderboardAdvancedActions
      isOpen={isLeaderboardActionsOpen}
      toggle={toggleLeaderboardShowContextForAllMeasures}
    />
  {/if}
</div>
