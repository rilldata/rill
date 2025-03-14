<script lang="ts">
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/DashboardVisibilityDropdown.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import LeaderboardMeasureCountSelector from "@rilldata/web-common/components/menu/LeaderboardMeasureCountSelector.svelte";
  import { featureFlags } from "../../feature-flags";
  import LeaderboardActiveMeasureDropdown from "@rilldata/web-common/components/menu/LeaderboardActiveMeasureDropdown.svelte";

  export let exploreName: string;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: {
        leaderboardMeasureCount,
        visibleMeasures,
        allMeasures,
        leaderboardMeasureName,
        getMeasureByName,
      },
      dimensions: { visibleDimensions, allDimensions },
      sorting: { sortByMeasure },
    },
    actions: {
      dimensions: { toggleDimensionVisibility },
      contextColumn: { setContextColumn },
      setLeaderboardMeasureCount,
      setLeaderboardMeasureName,
      sorting: { setDefaultSort },
    },
  } = StateManagers;

  const { leaderboardMeasureCount: leaderboardMeasureCountFeatureFlag } =
    featureFlags;

  $: measures = getSimpleMeasures($visibleMeasures);

  $: metricsExplorer = $metricsExplorerStore.entities[exploreName];

  $: activeLeaderboardMeasure = $getMeasureByName($leaderboardMeasureName);

  // If any measure has validPercentOfTotal, then the percent of total context column is valid
  $: validPercentOfTotal = $visibleMeasures.some(
    (measure) => measure.validPercentOfTotal,
  );

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
      <DashboardVisibilityDropdown
        category="Dimensions"
        tooltipText="Choose dimensions to display"
        onSelect={(name) => toggleDimensionVisibility(allDimensionNames, name)}
        selectableItems={$allDimensions.map(({ name, displayName }) => ({
          name: name || "",
          label: displayName || name || "",
        }))}
        selectedItems={visibleDimensionsNames}
        onToggleSelectAll={() => {
          toggleDimensionVisibility(allDimensionNames);
        }}
      />

      {#if !$leaderboardMeasureCountFeatureFlag}
        <LeaderboardActiveMeasureDropdown
          leaderboardMeasureName={$leaderboardMeasureName}
          {setLeaderboardMeasureName}
          {measures}
          {activeLeaderboardMeasure}
        />
      {/if}

      {#if $leaderboardMeasureCountFeatureFlag}
        <LeaderboardMeasureCountSelector
          measures={$allMeasures}
          count={$leaderboardMeasureCount}
          onMeasureCountChange={(count) => {
            setLeaderboardMeasureCount(count);
          }}
          setSort={() => setDefaultSort()}
          sortByMeasure={$sortByMeasure}
        />
      {/if}
    </div>
  {/if}
</div>
