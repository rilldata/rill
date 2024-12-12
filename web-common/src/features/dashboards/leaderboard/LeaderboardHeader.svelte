<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { SortType } from "../proto-state/derived-types";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import PieChart from "@rilldata/web-common/components/icons/PieChart.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import DimensionCompareMenu from "./DimensionCompareMenu.svelte";

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
  export let toggleSort: (sortType: SortType) => void;
  export let setPrimaryDimension: (dimensionName: string) => void;
  export let toggleComparisonDimension: (
    dimensionName: string | undefined,
  ) => void;
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
        #{#if sortType === SortType.VALUE}
          <ArrowDown flip={sortedAscending} />
        {/if}
      </button>
    </th>

    {#if isTimeComparisonActive}
      <th>
        <button
          aria-label="Toggle sort leaderboards by absolute change"
          on:click={() => toggleSort(SortType.DELTA_ABSOLUTE)}
        >
          <Delta />
          {#if sortType === SortType.DELTA_ABSOLUTE}
            <ArrowDown flip={sortedAscending} />
          {/if}
        </button>
      </th>

      <th>
        <button
          aria-label="Toggle sort leaderboards by percent change"
          on:click={() => toggleSort(SortType.DELTA_PERCENT)}
        >
          <Delta /> %
          {#if sortType === SortType.DELTA_PERCENT}
            <ArrowDown flip={sortedAscending} />
          {/if}
        </button>
      </th>
    {:else if isValidPercentOfTotal}
      <th>
        <button
          aria-label="Toggle sort leaderboards by percent of total"
          on:click={() => toggleSort(SortType.PERCENT)}
        >
          <PieChart /> %
          {#if sortType === SortType.PERCENT}
            <ArrowDown flip={sortedAscending} />
          {/if}
        </button>
      </th>
    {/if}
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
