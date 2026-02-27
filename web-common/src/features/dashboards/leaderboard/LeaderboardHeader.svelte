<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { fly } from "svelte/transition";
  import DeltaChange from "../dimension-table/DeltaChange.svelte";
  import DeltaChangePercentage from "../dimension-table/DeltaChangePercentage.svelte";
  import PercentOfTotal from "../dimension-table/PercentOfTotal.svelte";
  import { SortType } from "../proto-state/derived-types";
  import DimensionCompareMenu from "./DimensionCompareMenu.svelte";

  export let dimensionName: string;
  export let isFetching: boolean;
  export let isValidPercentOfTotal: (measureName: string) => boolean;
  export let isTimeComparisonActive: boolean;
  export let isBeingCompared: boolean;
  export let sortedAscending: boolean;
  export let displayName: string;
  export let dimensionDescription: string;
  export let hovered: boolean;
  export let sortType: SortType;
  export let allowDimensionComparison: boolean;
  export let allowExpandTable: boolean;
  export let leaderboardMeasureNames: string[] = [];
  export let leaderboardSortByMeasureName: string | null;
  export let leaderboardShowContextForAllMeasures: boolean;
  export let toggleSort: (sortType: SortType, measureName?: string) => void;
  export let setPrimaryDimension: (dimensionName: string) => void;
  export let toggleComparisonDimension: (
    dimensionName: string | undefined,
  ) => void;
  export let measureLabel: (measureName: string) => string;

  function shouldShowContextColumns(measureName: string): boolean {
    return (
      leaderboardShowContextForAllMeasures ||
      measureName === leaderboardSortByMeasureName
    );
  }
</script>

<thead>
  <tr>
    <th aria-label="Comparison column" class="grid place-content-center">
      {#if isFetching}
        <DelayedSpinner isLoading={isFetching} size="16px" />
      {:else if allowDimensionComparison && (hovered || isBeingCompared)}
        <DimensionCompareMenu
          {dimensionName}
          {isBeingCompared}
          {toggleComparisonDimension}
        />
      {:else}
        <Spacer size="14px" />
      {/if}
    </th>

    <th data-dimension-header>
      <Tooltip location="top">
        <button
          disabled={!allowExpandTable}
          class="text-fg-muted text-left {allowExpandTable
            ? 'hover:text-theme-700'
            : ''}"
          aria-label="Open dimension details"
          on:click={() => setPrimaryDimension(dimensionName)}
        >
          <span class="line-clamp-2">{displayName}</span>
        </button>
        <TooltipContent slot="tooltip-content" maxWidth="280px">
          <div
            class="pointer-events-none items-baseline"
            aria-label="tooltip-name"
          >
            {displayName}
          </div>
          {#if dimensionDescription}
            <div
              class="grid gap-x-2 pointer-events-none pt-1 pb-1 items-baseline"
              style="grid-template-columns: auto max-content"
              style:min-width="200px"
            >
              <div
                class="text-fg-muted justify-self-start"
                style:max-width="280px"
                aria-label="tooltip-name-description"
              >
                {dimensionDescription}
              </div>
            </div>
          {/if}
        </TooltipContent>
      </Tooltip>
    </th>

    {#each leaderboardMeasureNames as measureName, index (index)}
      <th data-measure-header>
        <button
          aria-label="Toggle sort leaderboards by value"
          on:click={() => {
            toggleSort(SortType.VALUE, measureName);
          }}
          class="font-normal text-right"
        >
          <span
            class="measure-label line-clamp-2 text-fg-muted"
            title={measureLabel(measureName)}
          >
            {#if leaderboardMeasureNames.length > 1}
              {measureLabel(measureName)}
            {:else}
              #
            {/if}
          </span>
          {#if measureName === leaderboardSortByMeasureName && sortType === SortType.VALUE}
            <div class="text-fg-muted">
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

      {#if isValidPercentOfTotal(measureName) && shouldShowContextColumns(measureName)}
        <th data-percent-of-total-header>
          <button
            aria-label="Toggle sort leaderboards by percent of total"
            on:click={() => toggleSort(SortType.PERCENT, measureName)}
          >
            <PercentOfTotal />
            {#if sortType === SortType.PERCENT && measureName === leaderboardSortByMeasureName}
              <div class="text-fg-muted">
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

      {#if isTimeComparisonActive && shouldShowContextColumns(measureName)}
        <th data-absolute-change-header>
          <button
            aria-label="Toggle sort leaderboards by absolute change"
            on:click={() => toggleSort(SortType.DELTA_ABSOLUTE, measureName)}
          >
            <DeltaChange />
            {#if sortType === SortType.DELTA_ABSOLUTE && measureName === leaderboardSortByMeasureName}
              <div class="text-fg-muted">
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

      {#if isTimeComparisonActive && shouldShowContextColumns(measureName)}
        <th data-percent-change-header>
          <button
            aria-label="Toggle sort leaderboards by percent change"
            on:click={() => toggleSort(SortType.DELTA_PERCENT, measureName)}
          >
            <DeltaChangePercentage />
            {#if sortType === SortType.DELTA_PERCENT && measureName === leaderboardSortByMeasureName}
              <div class="text-fg-muted">
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
    @apply sticky left-0 z-30 bg-surface-background text-left;
  }

  th:not(:first-of-type) {
    @apply border-b;
  }

  th:not(:nth-of-type(2)) button {
    @apply justify-end;
  }
</style>
