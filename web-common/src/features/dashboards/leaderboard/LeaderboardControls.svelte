<script lang="ts">
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/DashboardVisibilityDropdown.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import LeaderboardMeasureCountSelector from "@rilldata/web-common/components/menu/LeaderboardMeasureCountSelector.svelte";
  import ContextColumnDropdown from "@rilldata/web-common/components/menu/ContextColumnDropdown.svelte";
  import { get } from "svelte/store";

  export let exploreName: string;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: { leaderboardMeasureName, getMeasureByName, visibleMeasures },
      dimensions: { visibleDimensions, allDimensions },
      contextColumn: { contextColumnFilters },
    },
    actions: {
      dimensions: { toggleDimensionVisibility },
      contextCol: { setContextColumn, setContextColumnFilters },
      setLeaderboardMeasureCount,
    },
  } = StateManagers;

  $: measures = getSimpleMeasures($visibleMeasures);

  $: metricsExplorer = get(metricsExplorerStore).entities[exploreName];
  $: leaderboardMeasureCount = metricsExplorer?.leaderboardMeasureCount || 1;

  $: activeLeaderboardMeasure = $getMeasureByName($leaderboardMeasureName);

  $: validPercentOfTotal =
    activeLeaderboardMeasure?.validPercentOfTotal || false;

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
  {#if measures.length && activeLeaderboardMeasure}
    <div
      class="flex flex-row items-center ui-copy-muted gap-x-1"
      style:max-width="450px"
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
        firstMeasure={activeLeaderboardMeasure}
        tooltipText="Choose number of measures to display"
        selectedMeasureCount={leaderboardMeasureCount}
        onToggle={(count) => {
          setLeaderboardMeasureCount(count);
        }}
      />

      <!-- <ContextColumnDropdown
        tooltipText="Choose context columns to display"
        isValidPercentOfTotal={validPercentOfTotal}
        selectedFilters={$contextColumnFilters}
        {measures}
        selectedMeasureNames={[$leaderboardMeasureName]}
        onToggle={setContextColumnFilters}
        onSelectAll={() => {
          console.log("TODO: show for all measures");
        }}
      /> -->
    </div>
  {/if}
</div>
