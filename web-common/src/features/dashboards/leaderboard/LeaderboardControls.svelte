<script lang="ts">
  import DashboardMetricsDraggableList from "@rilldata/web-common/components/menu/DashboardMetricsDraggableList.svelte";
  import LeaderboardActiveMeasureDropdown from "@rilldata/web-common/components/menu/LeaderboardActiveMeasureDropdown.svelte";
  import LeaderboardMeasureCountSelector from "@rilldata/web-common/components/menu/LeaderboardMeasureCountSelector.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import LeaderboardActiveMeasureNamesDropdown from "@rilldata/web-common/components/menu/LeaderboardActiveMeasureNamesDropdown.svelte";

  export let exploreName: string;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: {
        leaderboardMeasureCount,
        leaderboardSortByMeasureName,
        getMeasureByName,
        visibleMeasures,
      },
      dimensions: { visibleDimensions, allDimensions },
    },
    actions: {
      contextColumn: { setContextColumn },
      dimensions: { setDimensionVisibility },
      setLeaderboardMeasureCount,
      setLeaderboardSortByMeasureName,
      setLeaderboardMeasureNames,
    },
  } = StateManagers;

  const { leaderboardMeasureCount: leaderboardMeasureCountFeatureFlag } =
    featureFlags;

  $: measures = getSimpleMeasures($visibleMeasures);

  $: metricsExplorer = $metricsExplorerStore.entities[exploreName];

  $: activeLeaderboardMeasure = $getMeasureByName(
    $leaderboardSortByMeasureName,
  );

  $: validPercentOfTotal = leaderboardMeasureCountFeatureFlag
    ? $visibleMeasures.some((measure) => measure.validPercentOfTotal)
    : activeLeaderboardMeasure?.validPercentOfTotal || false;

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
      <DashboardMetricsDraggableList
        type="dimension"
        onSelectedChange={(items) =>
          setDimensionVisibility(items, allDimensionNames)}
        allItems={$allDimensions}
        selectedItems={visibleDimensionsNames}
      />
      <LeaderboardMeasureCountSelector
        measures={$visibleMeasures}
        count={$leaderboardMeasureCount}
        onMeasureCountChange={(count) => {
          setLeaderboardMeasureCount(count);
        }}
      />
      <LeaderboardActiveMeasureDropdown
        leaderboardSortByMeasureName={$leaderboardSortByMeasureName}
        {setLeaderboardSortByMeasureName}
        {measures}
        {activeLeaderboardMeasure}
      />
      <LeaderboardActiveMeasureNamesDropdown
        {measures}
        sortBy={$leaderboardSortByMeasureName}
        tooltipText="Choose measures to filter by"
        selectedMeasureNames={[$leaderboardSortByMeasureName]}
        onSelect={(names) => {
          setLeaderboardMeasureNames(names);
        }}
        onToggleSelectAll={() => {
          console.log("toggle all");
        }}
      />
    </div>
  {/if}
</div>
