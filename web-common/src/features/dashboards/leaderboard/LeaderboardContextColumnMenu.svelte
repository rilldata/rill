<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { runtime } from "../../../runtime-client/runtime-store";

  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";
  import SelectMenu from "@rilldata/web-common/components/menu/compositions/SelectMenu.svelte";
  import type { SelectMenuItem } from "@rilldata/web-common/components/menu/types";

  export let metricViewName: string;
  export let validPercentOfTotal: boolean;

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;
  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  const handleContextValueButtonGroupClick = (evt) => {
    const value: SelectMenuItem = evt.detail;
    const key = value.key;

    if (key === LeaderboardContextColumn.HIDDEN) {
      metricsExplorerStore.hideContextColumn(metricViewName);
    } else if (key === LeaderboardContextColumn.DELTA_CHANGE) {
      metricsExplorerStore.displayDeltaChange(metricViewName);
    } else if (key === LeaderboardContextColumn.PERCENT) {
      metricsExplorerStore.displayPercentOfTotal(metricViewName);
    } else if (key === LeaderboardContextColumn.DELTA_ABSOLUTE) {
      metricsExplorerStore.displayDeltaAbsolute(metricViewName);
    }
  };

  let options: SelectMenuItem[];
  $: options = [
    {
      main: "Percent of total",
      key: LeaderboardContextColumn.PERCENT,
      disabled: !validPercentOfTotal,
    },
    {
      main: "Percent change",
      key: LeaderboardContextColumn.DELTA_CHANGE,
      disabled:
        !hasTimeSeries ||
        !metricsExplorer.showComparison ||
        metricsExplorer.selectedComparisonTimeRange === undefined,
    },
    {
      main: "Absolute change",
      key: LeaderboardContextColumn.DELTA_ABSOLUTE,
      disabled:
        !hasTimeSeries ||
        !metricsExplorer.showComparison ||
        metricsExplorer.selectedComparisonTimeRange === undefined,
    },
    {
      main: "No context column",
      key: LeaderboardContextColumn.HIDDEN,
    },
  ];

  let selection: SelectMenuItem;

  $: selection = options.find(
    (option) => option.key === metricsExplorer?.leaderboardContextColumn
  );
</script>

<SelectMenu
  {options}
  {selection}
  fixedText="with"
  ariaLabel="Select a context column"
  paddingTop={2}
  paddingBottom={2}
  alignment="end"
  on:select={handleContextValueButtonGroupClick}
/>
