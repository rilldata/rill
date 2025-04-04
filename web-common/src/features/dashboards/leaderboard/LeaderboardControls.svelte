<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import LeaderboardActiveMeasureDropdown from "@rilldata/web-common/components/menu/LeaderboardActiveMeasureDropdown.svelte";
  import DashboardMetricsDraggableList from "@rilldata/web-common/components/menu/DashboardMetricsDraggableList.svelte";
  import LeaderboardActiveMeasureNamesDropdown from "@rilldata/web-common/components/menu/LeaderboardActiveMeasureNamesDropdown.svelte";

  export let exploreName: string;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: {
        leaderboardMeasureName,
        leaderboardMeasures,
        getMeasureByName,
        visibleMeasures,
        allMeasures,
      },
      dimensions: { visibleDimensions, allDimensions },
    },
    actions: {
      contextColumn: { setContextColumn },
      dimensions: { setDimensionVisibility },
      setLeaderboardMeasureName,
      setLeaderboardMeasures,
    },
  } = StateManagers;

  $: measures = getSimpleMeasures($visibleMeasures);

  $: metricsExplorer = $metricsExplorerStore.entities[exploreName];

  $: activeLeaderboardMeasure = $getMeasureByName($leaderboardMeasureName);

  $: validPercentOfTotal =
    activeLeaderboardMeasure?.validPercentOfTotal || false;

  $: visibleDimensionsNames = $visibleDimensions
    .map(({ name }) => name)
    .filter(isDefined);
  $: allDimensionNames = $allDimensions
    .map(({ name }) => name)
    .filter(isDefined);
  $: allMeasureNames = $allMeasures.map(({ name }) => name).filter(isDefined);

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

  $: console.log("leaderboardMeasures: ", $leaderboardMeasures);
</script>

<div>
  {#if measures.length}
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
      <LeaderboardActiveMeasureDropdown
        sortBy={$leaderboardMeasureName}
        setSortBy={setLeaderboardMeasureName}
        {measures}
        {activeLeaderboardMeasure}
      />
      <LeaderboardActiveMeasureNamesDropdown
        {measures}
        sortBy={$leaderboardMeasureName}
        tooltipText="Choose leaderboard measures to filter by"
        selectedMeasureNames={$leaderboardMeasures}
        onSelect={(names) => {
          setLeaderboardMeasures(names);
        }}
        onToggleSelectAll={() => {
          console.log("toggle all");
        }}
      />
    </div>
  {/if}
</div>
