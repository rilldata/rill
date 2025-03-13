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
  import { fly } from "svelte/transition";

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
  export let contextColumns: string[] = [];
  export let activeMeasureNames: string[] = [];
  export let sortBy: string | null;
  export let dimensionShowAllMeasures: boolean;
  export let activeMeasureName: string;
  export let toggleSort: (sortType: SortType, measureName?: string) => void;
  export let setPrimaryDimension: (dimensionName: string) => void;
  export let toggleComparisonDimension: (
    dimensionName: string | undefined,
  ) => void;
  export let measureLabel: (measureName: string) => string;

  $: showPercentOfTotal =
    isValidPercentOfTotal &&
    contextColumns.includes(LeaderboardContextColumn.PERCENT);

  $: showDeltaAbsolute =
    isTimeComparisonActive &&
    contextColumns.includes(LeaderboardContextColumn.DELTA_ABSOLUTE);

  $: showDeltaPercent =
    isTimeComparisonActive &&
    contextColumns.includes(LeaderboardContextColumn.DELTA_PERCENT);

  function shouldShowComparisonForMeasure(measureName: string): boolean {
    return dimensionShowAllMeasures || measureName === activeMeasureName;
  }
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

    <th data-dimension-header>
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

    {#each activeMeasureNames as measureName, index (index)}
      <th>
        <button
          aria-label="Toggle sort leaderboards by value"
          on:click={() => {
            toggleSort(SortType.VALUE, measureName);
          }}
          class="font-normal text-right"
        >
          <span
            class="measure-label line-clamp-2"
            title={measureLabel(measureName)}
          >
            {measureLabel(measureName)}
          </span>
          {#if measureName === sortBy && sortType === SortType.VALUE}
            <div class="ui-copy-icon">
              {#if sortedAscending}
                <div in:fly|global={{ duration: 200, y: 8 }} style:opacity={1}>
                  <ArrowDown flip />
                </div>
              {:else}
                <div in:fly|global={{ duration: 200, y: -8 }} style:opacity={1}>
                  <ArrowDown />
                </div>
              {/if}
            </div>
          {/if}
        </button>
      </th>

      {#if shouldShowComparisonForMeasure(measureName)}
        {#if showPercentOfTotal}
          <th>
            <button
              aria-label="Toggle sort leaderboards by percent of total"
              on:click={() => toggleSort(SortType.PERCENT, measureName)}
            >
              <PercentOfTotal />
              {#if sortType === SortType.PERCENT && measureName === sortBy}
                <div class="ui-copy-icon">
                  {#if sortedAscending}
                    <div
                      in:fly|global={{ duration: 200, y: 8 }}
                      style:opacity={1}
                    >
                      <ArrowDown flip />
                    </div>
                  {:else}
                    <div
                      in:fly|global={{ duration: 200, y: -8 }}
                      style:opacity={1}
                    >
                      <ArrowDown />
                    </div>
                  {/if}
                </div>
              {/if}
            </button>
          </th>
        {/if}

        {#if showDeltaAbsolute}
          <th>
            <button
              aria-label="Toggle sort leaderboards by absolute change"
              on:click={() => toggleSort(SortType.DELTA_ABSOLUTE, measureName)}
            >
              <DeltaChange />
              {#if sortType === SortType.DELTA_ABSOLUTE && measureName === sortBy}
                <div class="ui-copy-icon">
                  {#if sortedAscending}
                    <div
                      in:fly|global={{ duration: 200, y: 8 }}
                      style:opacity={1}
                    >
                      <ArrowDown flip />
                    </div>
                  {:else}
                    <div
                      in:fly|global={{ duration: 200, y: -8 }}
                      style:opacity={1}
                    >
                      <ArrowDown />
                    </div>
                  {/if}
                </div>
              {/if}
            </button>
          </th>
        {/if}

        {#if showDeltaPercent}
          <th>
            <button
              aria-label="Toggle sort leaderboards by percent change"
              on:click={() => toggleSort(SortType.DELTA_PERCENT, measureName)}
            >
              <DeltaChangePercentage />
              {#if sortType === SortType.DELTA_PERCENT && measureName === sortBy}
                <div class="ui-copy-icon">
                  {#if sortedAscending}
                    <div
                      in:fly|global={{ duration: 200, y: 8 }}
                      style:opacity={1}
                    >
                      <ArrowDown flip />
                    </div>
                  {:else}
                    <div
                      in:fly|global={{ duration: 200, y: -8 }}
                      style:opacity={1}
                    >
                      <ArrowDown />
                    </div>
                  {/if}
                </div>
              {/if}
            </button>
          </th>
        {/if}
      {/if}
    {/each}
  </tr>
</thead>

<style lang="postcss">
  button {
    @apply px-2 flex items-center justify-start size-full;
  }

  th {
    @apply p-0 text-right h-8;
  }

  th[data-dimension-header] {
    @apply sticky left-0 z-30 bg-surface text-left;
  }

  th:not(:first-of-type) {
    @apply border-b;
  }

  th:not(:nth-of-type(2)) button {
    @apply justify-end;
  }
</style>
