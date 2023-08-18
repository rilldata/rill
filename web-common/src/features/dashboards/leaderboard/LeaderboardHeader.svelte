<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Spinner from "../../entity-management/Spinner.svelte";
  import LeaderboardOptionsMenu from "../leaderboard/LeaderboardOptionsMenu.svelte";
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import PieChart from "@rilldata/web-common/components/icons/PieChart.svelte";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import { CONTEXT_COLUMN_WIDTH } from "./leaderboard-utils";
  import { createEventDispatcher } from "svelte";
  import { LeaderboardContextColumn } from "../leaderboard-context-column";

  export let displayName: string;
  export let isFetching: boolean;
  export let dimensionDescription: string;
  export let hovered: boolean;
  export let contextColumn: LeaderboardContextColumn;
  export let sortAscending: boolean;

  export let filterExcludeMode: boolean;

  let optionsMenuActive = false;
  const dispatch = createEventDispatcher();

  $: arrowTransform = sortAscending ? "scale(1 -1)" : "scale(1 1)";
</script>

<div class="flex flex-row items-center">
  <div class="grid place-items-center" style:height="22px" style:width="22px">
    {#if isFetching}
      <Spinner size="16px" status={EntityStatus.Running} />
    {:else if hovered || optionsMenuActive}
      <div style="position:relative; height:100%; width:100%; ">
        <div style="position: absolute; ">
          <LeaderboardOptionsMenu
            bind:optionsMenuActive
            on:toggle-filter-mode
            {filterExcludeMode}
          />
        </div>
      </div>
    {/if}
  </div>

  <div
    class="
        pr-2
        grid justify-between items-center
        w-full
        border-b
        border-gray-200
        rounded-t
        surface
        ui-copy-muted
        font-semibold
        truncate
    "
    style="max-width: calc(100% - 22px);"
    style:flex="1"
    style:grid-template-columns="auto max-content"
    style:height="32px"
  >
    <div>
      <Tooltip distance={16} location="top">
        <button
          on:click={() => dispatch("open-dimension-details")}
          class="pl-2 truncate"
          style="max-width: calc(315px - 60px);"
          aria-label="Open dimension details"
        >
          {displayName}
        </button>
        <TooltipContent slot="tooltip-content">
          <TooltipTitle>
            <svelte:fragment slot="name">
              {displayName}
            </svelte:fragment>
            <svelte:fragment slot="description" />
          </TooltipTitle>
          <TooltipShortcutContainer>
            <div>
              {#if dimensionDescription}
                {dimensionDescription}
              {:else}
                The leaderboard metrics for {displayName}
              {/if}
            </div>
            <Shortcut />
            <div>Expand leaderboard</div>
            <Shortcut>Click</Shortcut>
          </TooltipShortcutContainer>
        </TooltipContent>
      </Tooltip>
    </div>
    <div class="shrink flex flex-row items-center gap-x-4">
      <button
        on:click={() => dispatch("toggle-sort-direction")}
        class="shrink flex flex-row items-center"
        aria-label="Toggle sort order for all leaderboards"
      >
        # <ArrowDown transform={arrowTransform} />
      </button>

      {#if contextColumn === LeaderboardContextColumn.DELTA_CHANGE}
        <Delta /> %
      {:else if contextColumn === LeaderboardContextColumn.DELTA_ABSOLUTE}
        <Delta />
      {:else if contextColumn === LeaderboardContextColumn.PERCENT}
        <PieChart /> %
      {/if}
    </div>
  </div>
</div>
