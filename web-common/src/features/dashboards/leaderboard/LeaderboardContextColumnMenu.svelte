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
    const key = value.key;
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

  let selection: SelectMenuItem;

  $: selection = options.find(
    (option) => option.key === metricsExplorer?.leaderboardContextColumn
  );
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
