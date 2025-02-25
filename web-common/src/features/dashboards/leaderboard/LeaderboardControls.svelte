<script lang="ts">
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/DashboardVisibilityDropdown.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import LeaderboardActiveMeasureNamesDropdown from "@rilldata/web-common/components/menu/LeaderboardActiveMeasureNamesDropdown.svelte";

  export let exploreName: string;

  const {
    selectors: {
      measures: {
        leaderboardMeasureNames,
        getMeasureByName,
        visibleMeasures,
        allMeasures,
      },
      dimensions: { visibleDimensions, allDimensions },
    },
    actions: {
      dimensions: { toggleDimensionVisibility },
      contextCol: { setContextColumn },
      setLeaderboardMeasureNames,
    },
  } = getStateManagers();

  $: measures = getSimpleMeasures($visibleMeasures);

  $: metricsExplorer = $metricsExplorerStore.entities[exploreName];

  $: activeLeaderboardMeasures = $leaderboardMeasureNames.map((name) =>
    $getMeasureByName(name),
  );
  // $: console.log("activeLeaderboardMeasures: ", activeLeaderboardMeasures);

  // First measure is used for sorting and validation
  $: firstMeasure = activeLeaderboardMeasures[0];
  $: validPercentOfTotal = firstMeasure?.validPercentOfTotal || false;

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

  function toggleLeaderboardMeasureNames() {
    const allSelected =
      allMeasureNames.length === $leaderboardMeasureNames.length;
    setLeaderboardMeasureNames(
      allSelected ? [allMeasureNames[0]] : allMeasureNames,
    );
  }
</script>

<div>
  {#if measures.length && firstMeasure}
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

      <LeaderboardActiveMeasureNamesDropdown
        {measures}
        {firstMeasure}
        tooltipText="Choose measures to filter by"
        selectedMeasureNames={$leaderboardMeasureNames}
        onSelect={(names) => {
          setLeaderboardMeasureNames(names);
        }}
        onToggleSelectAll={() => {
          toggleLeaderboardMeasureNames();
        }}
      />
    </div>
  {/if}
</div>
