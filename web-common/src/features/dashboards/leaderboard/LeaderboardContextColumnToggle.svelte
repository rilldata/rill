<script lang="ts">
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import PieChart from "@rilldata/web-common/components/icons/PieChart.svelte";
  import {
    ButtonGroup,
    SubButton,
  } from "@rilldata/web-common/components/button-group";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { runtime } from "../../../runtime-client/runtime-store";

  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    LeaderboardContextColumn,
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";

  export let metricViewName: string;
  export let validPercentOfTotal: boolean;

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;
  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  let disabledButtons: (
    | LeaderboardContextColumn.DELTA_CHANGE
    | LeaderboardContextColumn.PERCENT
  )[] = [];
  $: {
    disabledButtons = [];
    if (
      !hasTimeSeries ||
      !metricsExplorer.showComparison ||
      metricsExplorer.selectedComparisonTimeRange === undefined
    )
      disabledButtons.push(LeaderboardContextColumn.DELTA_CHANGE);
    if (validPercentOfTotal !== true)
      disabledButtons.push(LeaderboardContextColumn.PERCENT);
  }

  let selectedButton: LeaderboardContextColumn;
  // NOTE: time comparison takes precedence over percent of total
  $: selectedButton = metricsExplorer?.leaderboardContextColumn;

  const handleContextValueButtonGroupClick = (evt) => {
    const value = evt.detail;

    // hide context column if the button that is
    // clicked is already selected
    if (value === selectedButton) {
      metricsExplorerStore.hideContextColumn(metricViewName);
      return;
    }

    // If a non-selected button is clicked, show the corresponding
    // context column
    if (value === LeaderboardContextColumn.DELTA_CHANGE) {
      metricsExplorerStore.displayDeltaChange(metricViewName);
    } else if (value === LeaderboardContextColumn.PERCENT) {
      metricsExplorerStore.displayPercentOfTotal(metricViewName);
    }
  };

  $: selectedButtons = selectedButton === null ? [] : [selectedButton];

  const pieTooltips = {
    selected: "Hide percent of total",
    unselected: "Show percent of total",
    disabled:
      "To show percent of total, select a metric that is defined as summable",
  };

  const deltaTooltips = {
    selected: "Hide percent change",
    unselected: "Show percent change",
    disabled: "To show percent change, select a comparison time range",
  };
</script>

<ButtonGroup
  disabled={disabledButtons}
  on:subbutton-click={handleContextValueButtonGroupClick}
  selected={selectedButtons}
>
  <SubButton
    value={LeaderboardContextColumn.DELTA_CHANGE}
    tooltips={deltaTooltips}
    ariaLabel="Toggle percent change"
  >
    <Delta />%
  </SubButton>
  <SubButton
    value={LeaderboardContextColumn.PERCENT}
    tooltips={pieTooltips}
    ariaLabel="Toggle percent of total"
  >
    <PieChart />%
  </SubButton>
</ButtonGroup>
