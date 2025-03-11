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
        leaderboardMeasureCount,
        getMeasureByName,
        visibleMeasures,
        allMeasures,
      },
      dimensions: {
        visibleDimensions,
        allDimensions,
        dimensionShowAllMeasures,
      },
      contextColumn: { contextColumns },
    },
    actions: {
      dimensions: { toggleDimensionVisibility, toggleDimensionShowAllMeasures },
      contextColumn: { setContextColumn, setContextColumns },
      setLeaderboardMeasureCount,
    },
  } = StateManagers;

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
    metricsExplorer?.leaderboardContextColumns?.includes(
      LeaderboardContextColumn.PERCENT,
    )
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

      <LeaderboardMeasureCountSelector
        measures={$allMeasures}
        count={$leaderboardMeasureCount}
        onMeasureCountChange={(count) => {
          setLeaderboardMeasureCount(count);
        }}
      />

      <ContextColumnDropdown
        tooltipText="Choose context columns to display"
        isValidPercentOfTotal={validPercentOfTotal}
        isTimeComparisonActive={!!comparisonTimeRange}
        contextColumns={$contextColumns}
        dimensionShowAllMeasures={$dimensionShowAllMeasures}
        onContextColumnChange={setContextColumns}
        onShowForAllMeasures={toggleDimensionShowAllMeasures}
      />
    </div>
  {/if}
</div>
