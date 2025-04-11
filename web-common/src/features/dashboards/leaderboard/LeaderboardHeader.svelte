<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
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
  export let dimensionDescription: string;
  export let isBeingCompared: boolean;
  export let sortedAscending: boolean;
  export let displayName: string;
  export let hovered: boolean;
  export let sortType: SortType;
  export let allowDimensionComparison: boolean;
  export let allowExpandTable: boolean;
  export let leaderboardMeasureNames: string[] = [];
  export let sortBy: string | null;
  export let leaderboardMeasureCountFeatureFlag: boolean;
  export let toggleSort: (sortType: SortType, measureName?: string) => void;
  export let setPrimaryDimension: (dimensionName: string) => void;
  export let toggleComparisonDimension: (
    dimensionName: string | undefined,
  ) => void;
  export let measureLabel: (measureName: string) => string;
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
      <Tooltip distance={16} location="top">
        <button
          disabled={!allowExpandTable}
          class="text-slate-600 {allowExpandTable
            ? 'hover:text-primary-700'
            : ''}"
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
            {#if allowExpandTable}
              <div>Expand leaderboard</div>
              <Shortcut>Click</Shortcut>
            {/if}
          </TooltipShortcutContainer>
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
            class="measure-label line-clamp-2"
            title={measureLabel(measureName)}
          >
            {#if leaderboardMeasureCountFeatureFlag}
              {measureLabel(measureName)}
            {:else}
              #
            {/if}
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

      {#if isValidPercentOfTotal(measureName)}
        <th data-percent-of-total-header>
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

      {#if isTimeComparisonActive}
        <th data-absolute-change-header>
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

      {#if isTimeComparisonActive}
        <th data-percent-change-header>
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
