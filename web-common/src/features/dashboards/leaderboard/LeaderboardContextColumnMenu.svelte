<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import SelectMenu from "@rilldata/web-common/components/menu/compositions/SelectMenu.svelte";
  import type { SelectMenuItem } from "@rilldata/web-common/components/menu/types";

  export let metricViewName: string;
  export let validPercentOfTotal: boolean;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];
  const timeControlsStore = useTimeControlStore(getStateManagers());

  const handleContextValueButtonGroupClick = (evt) => {
    const value: SelectMenuItem = evt.detail;
    // CAST SAFETY: the value.key passed up from the evt must
    // be a LeaderboardContextColumn
    const key = value.key as LeaderboardContextColumn;
    metricsExplorerStore.setContextColumn(metricViewName, key);
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
      disabled: !$timeControlsStore.showComparison,
    },
    {
      main: "Absolute change",
      key: LeaderboardContextColumn.DELTA_ABSOLUTE,
      disabled: !$timeControlsStore.showComparison,
    },
    {
      main: "No context column",
      key: LeaderboardContextColumn.HIDDEN,
    },
  ];

  // CAST SAFETY: the selection will always be one of the options
  $: selection = options.find(
    (option) => option.key === metricsExplorer.leaderboardContextColumn,
  ) as SelectMenuItem;
</script>

<SelectMenu
  alignment="end"
  ariaLabel="Select a context column"
  fixedText="with"
  on:select={handleContextValueButtonGroupClick}
  {options}
  paddingBottom={2}
  paddingTop={2}
  {selection}
/>
