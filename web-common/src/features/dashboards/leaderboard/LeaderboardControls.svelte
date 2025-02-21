<script lang="ts">
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/DashboardVisibilityDropdown.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import MetricsDropdown from "@rilldata/web-common/components/menu/MetricsDropdown.svelte";

  export let exploreName: string;

  const {
    selectors: {
      measures: { leaderboardMeasureNames, getMeasureByName, visibleMeasures },
      dimensions: { visibleDimensions, allDimensions },
    },
    actions: {
      dimensions: { toggleDimensionVisibility },
      contextCol: { setContextColumn },
      setLeaderboardMeasureNames,
    },
  } = getStateManagers();

  $: console.log("leaderboardMeasureNames: ", $leaderboardMeasureNames);

  $: measures = getSimpleMeasures($visibleMeasures);

  $: metricsExplorer = $metricsExplorerStore.entities[exploreName];

  // TODO: support multiple active measures
  // First measure is used as the active measure
  $: activeLeaderboardMeasure = $getMeasureByName($leaderboardMeasureNames[0]);
  $: console.log("activeLeaderboardMeasure: ", activeLeaderboardMeasure);

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

      <MetricsDropdown
        {measures}
        activeMeasure={activeLeaderboardMeasure}
        tooltipText="Choose measures to filter by"
        selectedMeasureNames={$leaderboardMeasureNames}
        onSelect={(names) => {
          setLeaderboardMeasureNames(names);
        }}
      />
    </div>
  {/if}
</div>
