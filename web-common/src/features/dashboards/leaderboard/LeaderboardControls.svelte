<script lang="ts">
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/DashboardVisibilityDropdown.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import ActiveMeasureNamesDropdown from "@rilldata/web-common/components/menu/ActiveMeasureNameDropdown.svelte";
  import ContextColumnDropdown from "@rilldata/web-common/components/menu/ContextColumnDropdown.svelte";

  export let exploreName: string;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: {
        leaderboardMeasureName,
        getMeasureByName,
        visibleMeasures,
        allMeasures,
      },
      dimensions: { visibleDimensions, allDimensions },
      contextColumn: { contextColumnFilters },
    },
    actions: {
      dimensions: { toggleDimensionVisibility },
      contextCol: { setContextColumn, setContextColumnFilters },
      setLeaderboardMeasureName,
    },
  } = StateManagers;

  let active = false;

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

      <ActiveMeasureNamesDropdown
        {measures}
        firstMeasure={activeLeaderboardMeasure}
        tooltipText="Choose measures to display"
        selectedMeasureNames={[$leaderboardMeasureName]}
        onToggle={(name) => {
          setLeaderboardMeasureName(name);
        }}
        onSelectAll={() => {
          setLeaderboardMeasureName(allMeasureNames[0]);
        }}
      />

      <ContextColumnDropdown
        tooltipText="Choose context columns to display"
        isValidPercentOfTotal={validPercentOfTotal}
        selectedFilters={$contextColumnFilters}
        {measures}
        selectedMeasureNames={[$leaderboardMeasureName]}
        onToggle={setContextColumnFilters}
        onSelectAll={() => {
          console.log("TODO: show for all measures");
        }}
      />
    </div>
  {/if}
</div>
