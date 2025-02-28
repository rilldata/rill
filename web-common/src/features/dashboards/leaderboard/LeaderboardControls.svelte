<script lang="ts">
  import DashboardVisibilityDropdown from "@rilldata/web-common/components/menu/DashboardVisibilityDropdown.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "../state-managers/state-managers";
  import LeaderboardActiveMeasureNamesDropdown from "@rilldata/web-common/components/menu/LeaderboardActiveMeasureNamesDropdown.svelte";
  import ContextColumnDropdown from "@rilldata/web-common/components/menu/ContextColumnDropdown.svelte";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";

  export let exploreName: string;
  export let comparisonTimeRange: V1TimeRange | undefined;

  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: {
        leaderboardMeasureNames,
        getMeasureByName,
        visibleMeasures,
        allMeasures,
      },
      dimensions: { visibleDimensions, allDimensions },
      contextColumn: { contextColumnFilters },
    },
    actions: {
      dimensions: { toggleDimensionVisibility },
      contextColumn: { setContextColumn, setContextColumnFilters },
      toggleLeaderboardMeasureNames,
    },
  } = StateManagers;

  $: measures = getSimpleMeasures($visibleMeasures);

  $: metricsExplorer = $metricsExplorerStore.entities[exploreName];

  $: activeLeaderboardMeasures = $leaderboardMeasureNames.map((name) =>
    $getMeasureByName(name),
  );

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
</script>

<div>
  {#if measures.length && firstMeasure}
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

      <LeaderboardActiveMeasureNamesDropdown
        {measures}
        {firstMeasure}
        tooltipText="Choose measures to display"
        selectedMeasureNames={$leaderboardMeasureNames}
        onToggle={(name) => {
          toggleLeaderboardMeasureNames(allMeasureNames, name);
        }}
        onSelectAll={() => {
          toggleLeaderboardMeasureNames(allMeasureNames);
        }}
      />

      <!-- TODO: should we have another dropdown for sorting? -->
      <div class="ui-copy-muted truncate">
        Sort by <strong>{firstMeasure.displayName || firstMeasure.name}</strong>
      </div>

      {#if Boolean(comparisonTimeRange)}
        <ContextColumnDropdown
          tooltipText="Choose context columns to display"
          isValidPercentOfTotal={validPercentOfTotal}
          isTimeComparisonActive={Boolean(comparisonTimeRange)}
          selectedFilters={$contextColumnFilters}
          onContextColumnChange={setContextColumnFilters}
          {measures}
          selectedMeasureNames={$leaderboardMeasureNames}
          onSelectAll={() => {
            toggleLeaderboardMeasureNames(allMeasureNames);
          }}
        />
      {/if}
    </div>
  {/if}
</div>
