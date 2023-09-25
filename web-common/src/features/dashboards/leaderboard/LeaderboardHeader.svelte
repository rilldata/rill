<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Spinner from "../../entity-management/Spinner.svelte";
  import DimensionCompareMenu from "./DimensionCompareMenu.svelte";
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import PieChart from "@rilldata/web-common/components/icons/PieChart.svelte";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import { createEventDispatcher } from "svelte";
  import { LeaderboardContextColumn } from "../leaderboard-context-column";
  import { contextColumnWidth } from "./leaderboard-utils";
  import { SortType } from "../proto-state/derived-types";

  export let displayName: string;
  export let isFetching: boolean;
  export let dimensionDescription: string;
  export let hovered: boolean;
  export let contextColumn: LeaderboardContextColumn;
  export let sortAscending: boolean;
  export let sortType: SortType;
  export let isBeingCompared: boolean;

  const dispatch = createEventDispatcher();
  $: contextColumnSortType = {
    [LeaderboardContextColumn.DELTA_PERCENT]: SortType.DELTA_PERCENT,
    [LeaderboardContextColumn.DELTA_ABSOLUTE]: SortType.DELTA_ABSOLUTE,
    [LeaderboardContextColumn.PERCENT]: SortType.PERCENT,
  }[contextColumn];

  $: arrowTransform = sortAscending ? "scale(1 -1)" : "scale(1 1)";
</script>

<div class="flex flex-row items-center">
  <div class="grid place-items-center" style:height="22px" style:width="22px">
    {#if isFetching}
      <Spinner size="16px" status={EntityStatus.Running} />
    {:else if hovered || isBeingCompared}
      <div style="position:relative; height:100%; width:100%; ">
        <div style="position: absolute; ">
          <DimensionCompareMenu
            {isBeingCompared}
            on:toggle-dimension-comparison
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
        on:click={() => dispatch("toggle-sort", SortType.VALUE)}
        class="shrink flex flex-row items-center justify-end"
        aria-label="Toggle sort leaderboards by value"
      >
        #{#if sortType === SortType.VALUE}
          <ArrowDown transform={arrowTransform} />
        {/if}
      </button>

      {#if contextColumn !== LeaderboardContextColumn.HIDDEN}
        <button
          on:click={() => dispatch("toggle-sort", contextColumnSortType)}
          class="shrink flex flex-row items-center justify-end"
          aria-label="Toggle sort leaderboards by context column"
          style:width={contextColumnWidth(contextColumn)}
        >
          {#if contextColumn === LeaderboardContextColumn.DELTA_PERCENT}
            <Delta /> %
          {:else if contextColumn === LeaderboardContextColumn.DELTA_ABSOLUTE}
            <Delta />
          {:else if contextColumn === LeaderboardContextColumn.PERCENT}
            <PieChart /> %
          {/if}{#if sortType !== SortType.VALUE}
            <ArrowDown transform={arrowTransform} />
          {/if}
        </button>
      {/if}
    </div>
  </div>
</div>
