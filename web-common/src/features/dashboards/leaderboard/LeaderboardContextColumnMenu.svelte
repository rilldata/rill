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

  $: console.log(
    "leaderboardContextColumn",
    metricsExplorer?.leaderboardContextColumn
  );

  const handleContextValueButtonGroupClick = (evt) => {
    const value: SelectMenuItem = evt.detail;
    const key = value.key;

    if (key === LeaderboardContextColumn.HIDDEN) {
      metricsExplorerStore.hideContextColumn(metricViewName);
    } else if (key === LeaderboardContextColumn.DELTA_PERCENT) {
      metricsExplorerStore.displayDeltaChange(metricViewName);
    } else if (key === LeaderboardContextColumn.PERCENT) {
      metricsExplorerStore.displayPercentOfTotal(metricViewName);
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
      key: LeaderboardContextColumn.DELTA_PERCENT,
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
  $: console.log("options", options);

  $: console.log("selection", selection);
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
