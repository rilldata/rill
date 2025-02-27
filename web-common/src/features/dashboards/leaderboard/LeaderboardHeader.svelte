<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { SortType } from "../proto-state/derived-types";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import DimensionCompareMenu from "./DimensionCompareMenu.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import DeltaChangePercentage from "../dimension-table/DeltaChangePercentage.svelte";
  import DeltaChange from "../dimension-table/DeltaChange.svelte";
  import PercentOfTotal from "../dimension-table/PercentOfTotal.svelte";
  export let dimensionName: string;
  export let isFetching: boolean;
  export let isTimeComparisonActive: boolean;
  export let isValidPercentOfTotal: boolean;
  export let dimensionDescription: string;
  export let isBeingCompared: boolean;
  export let sortedAscending: boolean;
  export let displayName: string;
  export let hovered: boolean;
  export let sortType: SortType;
  export let contextColumnFilters: LeaderboardContextColumn[] = [];
  export let toggleSort: (sortType: SortType) => void;
  export let setPrimaryDimension: (dimensionName: string) => void;
  export let toggleComparisonDimension: (
    dimensionName: string | undefined,
  ) => void;

  $: showDeltaAbsolute =
    isTimeComparisonActive &&
    contextColumnFilters.includes(LeaderboardContextColumn.DELTA_ABSOLUTE);

  $: showDeltaPercent =
    isTimeComparisonActive &&
    contextColumnFilters.includes(LeaderboardContextColumn.DELTA_PERCENT);

  $: showPercentOfTotal =
    !isTimeComparisonActive &&
    isValidPercentOfTotal &&
    contextColumnFilters.includes(LeaderboardContextColumn.PERCENT);
</script>

<thead>
  <tr>
    <th aria-label="Comparison column">
      {#if isFetching}
        <DelayedSpinner isLoading={isFetching} size="16px" />
      {:else if hovered || isBeingCompared}
        <DimensionCompareMenu
          {dimensionName}
          {isBeingCompared}
          {toggleComparisonDimension}
        />
      {/if}
    </th>

    <th>
      <Tooltip distance={16} location="top">
        <button
          class="ui-header-primary"
          aria-label="Open dimension details"
          on:click={() => setPrimaryDimension(dimensionName)}
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
    </th>

    <th>
      <button
        aria-label="Toggle sort leaderboards by value"
        on:click={() => toggleSort(SortType.VALUE)}
      >
        # {#if sortType === SortType.VALUE}
          <ArrowDown flip={sortedAscending} />
        {/if}
      </button>
    </th>

    {#if showDeltaAbsolute}
      <th>
        <button
          aria-label="Toggle sort leaderboards by absolute change"
          on:click={() => toggleSort(SortType.DELTA_ABSOLUTE)}
        >
          <DeltaChange />
          {#if sortType === SortType.DELTA_ABSOLUTE}
            <ArrowDown flip={sortedAscending} />
          {/if}
        </button>
      </th>
    {/if}

    {#if showDeltaPercent}
      <th>
        <button
          aria-label="Toggle sort leaderboards by percent change"
          on:click={() => toggleSort(SortType.DELTA_PERCENT)}
        >
          <DeltaChangePercentage />
          {#if sortType === SortType.DELTA_PERCENT}
            <ArrowDown flip={sortedAscending} />
          {/if}
        </button>
      </th>
    {/if}

    {#if showPercentOfTotal}
      <th>
        <button
          aria-label="Toggle sort leaderboards by percent of total"
          on:click={() => toggleSort(SortType.PERCENT)}
        >
          <PercentOfTotal />
          {#if sortType === SortType.PERCENT}
            <ArrowDown flip={sortedAscending} />
          {/if}
        </button>
      </th>
    {/if}

    <!-- TODO: support new measure columns -->
  </tr>
</thead>

<style lang="postcss">
  button {
    @apply px-2 flex items-center justify-start size-full;
  }

  th {
    @apply p-0 text-right h-8;
  }

  th:first-of-type {
    @apply text-left;
  }

  th:not(:first-of-type) {
    @apply border-b;
  }

  th:not(:nth-of-type(2)) button {
    @apply justify-end;
  }
</style>
