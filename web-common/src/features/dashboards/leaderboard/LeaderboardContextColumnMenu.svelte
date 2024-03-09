<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import SelectMenu from "@rilldata/web-common/components/menu/shad-cn/SelectMenu.svelte";
  import type { SelectMenuItem } from "@rilldata/web-common/components/menu/types";

  export let validPercentOfTotal: boolean;

  const {
    selectors: {
      contextColumn: { contextColumn },
    },
    actions: {
      contextCol: { setContextColumn },
    },
  } = getStateManagers();

  const timeControlsStore = useTimeControlStore(getStateManagers());

  function handleContextValueButtonGroupClick(
    e: CustomEvent<SelectMenuItem & { key: LeaderboardContextColumn }>,
  ) {
    setContextColumn(e.detail.key);
  }

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
</script>

<SelectMenu
  fixedText="with"
  ariaLabel="Select a context column"
  {options}
  selections={[$contextColumn]}
  on:select={handleContextValueButtonGroupClick}
/>
