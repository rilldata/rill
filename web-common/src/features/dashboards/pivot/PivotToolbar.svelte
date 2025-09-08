<script context="module" lang="ts">
  export const lastNestState = writable<PivotChipData[] | null>(null);
</script>

<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import PivotPanel from "@rilldata/web-common/components/icons/PivotPanel.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { writable } from "svelte/store";
  import Collapse from "../../../components/icons/Collapse.svelte";
  import Pivot from "../../../components/icons/Pivot.svelte";
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import TableIcon from "../../canvas/icons/TableIcon.svelte";
  import { ButtonGroup, SubButton } from "@rilldata/web-common/components/button-group";
  import ExportMenu from "../../exports/ExportMenu.svelte";
  import { featureFlags } from "../../feature-flags";
  import { getStateManagers } from "../state-managers/state-managers";
  import { getPivotExportQuery } from "./pivot-export";
  import type { PivotChipData, PivotTableMode } from "./types";

  export let showPanels = true;
  export let isFetching = false;

  const { adminServer, exports } = featureFlags;

  const stateManagers = getStateManagers();
  const {
    exploreName,
    dashboardStore,
    timeRangeSummaryStore,
    selectors: {
      pivot: { rows, columns, isFlat },
    },
  } = stateManagers;

  $: expanded = $dashboardStore?.pivot?.expanded ?? {};
  $: exploreHasTimeDimension = !!$timeRangeSummaryStore.data;

  /**
   * This method stores the previous nest state and passes it to
   * dashboard store when toggling back from `flat` to `nest`
   */
  function togglePivotType(newJoinState: PivotTableMode) {
    if (newJoinState === "flat") {
      lastNestState.set($rows);
      metricsExplorerStore.setPivotTableMode(
        $exploreName,
        "flat",
        [],
        [...$columns.dimension, ...$rows, ...$columns.measure],
      );
      return;
    }

    // Handle nest state
    const updatedRows = $lastNestState ?? $columns.dimension;
    const rowDimensionIds = new Set(updatedRows.map((d) => d.id));

    metricsExplorerStore.setPivotTableMode($exploreName, "nest", updatedRows, [
      ...($lastNestState
        ? $columns.dimension.filter((d) => !rowDimensionIds.has(d.id))
        : []),
      ...$columns.measure,
    ]);
  }

  function blurCurrentTarget(e: MouseEvent) {
    (e.currentTarget as HTMLButtonElement | null)?.blur();
  }

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
</script>

<div class="flex items-center gap-x-4 select-none pointer-events-none">
  <Button
    square
    type="secondary"
    theme
    selected={showPanels}
    onClick={(e) => {
      showPanels = !showPanels;
      blurCurrentTarget(e);
    }}
  >
    <PivotPanel size="18px" open={showPanels} colorClass="fill-theme-800" />
  </Button>

  <div class="flex items-center gap-x-1 pointer-events-auto">
    <div class="toolbar-button-group">
      <ButtonGroup
        selected={$isFlat ? ["flat"] : ["pivot"]}
        on:subbutton-click={(event) => {
          togglePivotType(event.detail === "pivot" ? "nest" : "flat");
        }}
      >
        <SubButton
          value="pivot"
          tooltips={{
            selected: "Currently showing pivot view",
            unselected: "Switch to pivot view"
          }}
        >
          <Pivot size="16px" />
          Pivot
        </SubButton>
        <SubButton
          value="flat"
          tooltips={{
            selected: "Currently showing flat view", 
            unselected: "Switch to flat view"
          }}
        >
          <TableIcon size="16px" />
          Flat
        </SubButton>
      </ButtonGroup>
    </div>

    <!-- <Button
    compact
    type="text"
    onClick={() => {
      expandVisible();
    }}
  >
    Expand Visible
  </Button> -->
    <Button
      type="toolbar"
      onClick={() => {
        metricsExplorerStore.setPivotExpanded($exploreName, {});
      }}
      disabled={Object.keys(expanded).length === 0}
    >
      <Collapse size="16px" />
      Collapse All
    </Button>

    {#if $exports}
      <ExportMenu
        label="Export pivot data"
        includeScheduledReport={$adminServer && exploreHasTimeDimension}
        getQuery={(isScheduled) =>
          getPivotExportQuery(stateManagers, isScheduled)}
        exploreName={$exploreName}
      />
    {/if}

    {#if isFetching}
      <Spinner size="18px" status={EntityStatus.Running} />
    {/if}
  </div>
</div>

<style lang="postcss">
  /* Make ButtonGroup match toolbar button height (24px) */
  .toolbar-button-group :global(button) {
    @apply h-6 px-1.5 py-0;
    @apply text-sm font-normal;
    @apply flex items-center gap-x-1;
  }
</style>
