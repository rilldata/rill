<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import PivotPanel from "@rilldata/web-common/components/icons/PivotPanel.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import ExportMenu from "../../exports/ExportMenu.svelte";
  import { featureFlags } from "../../feature-flags";
  import { getStateManagers } from "../state-managers/state-managers";
  import { getPivotExportArgs } from "./pivot-export";

  export let showPanels = true;
  export let isFetching = false;

  const { adminServer, exports } = featureFlags;

  const stateManagers = getStateManagers();
  const { exploreName, dashboardStore } = stateManagers;

  $: expanded = $dashboardStore?.pivot?.expanded ?? {};
  $: metricsViewProto = $dashboardStore.proto;

  // function expandVisible() {
  //   // const lowestVisibleRow = 0;
  //   const nestedLevels = 4;
  //   const maxNestedLevelsToExpand = Math.max(3, nestedLevels);
  //   const maxExpandPerLevel = 3;

  //   // Helper function to recursively expand rows
  //   function expandRow(rowId: string, level: number) {
  //     if (level > maxNestedLevelsToExpand) {
  //       return;
  //     }

  //     expanded[rowId] = true; // Expand the current row

  //     // Generate and expand child rows
  //     for (let i = 0; i < maxExpandPerLevel; i++) {
  //       let childRowId = `${rowId}.${i}`;
  //       expandRow(childRowId, level + 1);
  //     }
  //   }

  //   // Expand rows starting from the lowest visible row
  //   for (let i = 0; i < maxExpandPerLevel; i++) {
  //     expandRow(i.toString(), 1); // Start from level 1
  //   }

  //   metricsExplorerStore.setPivotExpanded($exploreName, expanded);
  // }

  const scheduledReportsQueryArgs = getPivotExportArgs(stateManagers);
</script>

<div class="flex items-center gap-x-4 select-none pointer-events-none">
  <Button
    square
    type="secondary"
    selected={showPanels}
    on:click={(e) => {
      showPanels = !showPanels;
      e.detail.currentTarget.blur();
    }}
  >
    <PivotPanel size="18px" open={showPanels} />
  </Button>

  <!-- <Button
    compact
    type="text"
    on:click={() => {
      expandVisible();
    }}
  >
    Expand Visible
  </Button> -->
  {#if Object.keys(expanded).length > 0}
    <Button
      compact
      type="text"
      on:click={() => {
        metricsExplorerStore.setPivotExpanded($exploreName, {});
      }}
    >
      Collapse All
    </Button>
  {/if}

  {#if isFetching}
    <Spinner size="18px" status={EntityStatus.Running} />
  {/if}
  <div class="grow" />
  {#if $exports}
    <ExportMenu
      label="Export pivot data"
      includeScheduledReport={$adminServer}
      exploreName={$exploreName}
      query={{
        metricsViewAggregationRequest: $scheduledReportsQueryArgs,
      }}
      {metricsViewProto}
    />
  {/if}
</div>
