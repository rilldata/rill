<script lang="ts">
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/DashboardVisibilityDropdown.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import ContextColumnDropdown from "@rilldata/web-common/components/menu/ContextColumnDropdown.svelte";
  import LeaderboardMeasureCountSelector from "@rilldata/web-common/components/menu/LeaderboardMeasureCountSelector.svelte";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";

  export let exploreName: string;
  export let comparisonTimeRange: V1TimeRange | undefined;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: {
        leaderboardMeasureName,
        leaderboardMeasureNames,
        getMeasureByName,
        visibleMeasures,
        allMeasures,
      },
      dimensions: { visibleDimensions, allDimensions },
      contextColumn: { contextColumnFilters },
      activeMeasure: { leaderboardMeasureCount },
    },
    actions: {
      dimensions: { toggleDimensionVisibility },
      contextColumn: { setContextColumn, setContextColumnFilters },
      setLeaderboardMeasureCount,
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

  $: console.log("activeLeaderboardMeasure", activeLeaderboardMeasure);
  $: console.log("measures.length: ", measures.length);
</script>

<div>
  {#if measures.length}
    <!-- {#if measures.length && activeLeaderboardMeasure} -->
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

      <LeaderboardMeasureCountSelector
        {measures}
        tooltipText="Choose number of measures to display"
        count={$leaderboardMeasureCount}
        onToggle={(count) => {
          setLeaderboardMeasureCount(count);
        }}
      />

      <ContextColumnDropdown
        tooltipText="Choose context columns to display"
        isValidPercentOfTotal={validPercentOfTotal}
        hasComparisonTimeRange={!!comparisonTimeRange}
        selectedFilters={$contextColumnFilters}
        {measures}
        selectedMeasureNames={$leaderboardMeasureNames}
        onToggle={setContextColumnFilters}
        onSelectAll={() => {
          console.log("Fired onSelectAll in LeaderboardControls");
        }}
      />
    </div>
  {/if}
</div>
