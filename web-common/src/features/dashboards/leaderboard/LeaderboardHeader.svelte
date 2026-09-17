<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { fly } from "svelte/transition";
  import DeltaChange from "../dimension-table/DeltaChange.svelte";
  import DeltaChangePercentage from "../dimension-table/DeltaChangePercentage.svelte";
  import PercentOfTotal from "../dimension-table/PercentOfTotal.svelte";
  import { SortType } from "../proto-state/derived-types";
  import DimensionCompareMenu from "./DimensionCompareMenu.svelte";
  import {
    DEFAULT_DIMENSION_COLUMN_WIDTH,
    MAX_DIMENSION_COLUMN_WIDTH,
    MIN_DIMENSION_COLUMN_WIDTH,
  } from "./leaderboard-widths";

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
  // Measures that render context columns; see Leaderboard.svelte.
  export let measuresWithContext: Set<string>;
  export let toggleSort: (sortType: SortType, measureName?: string) => void;
  export let setPrimaryDimension: (dimensionName: string) => void;
  export let toggleComparisonDimension: (
    dimensionName: string | undefined,
  ) => void;
  export let measureLabel: (measureName: string) => string;
  export let dimensionColumnWidth: number;
  export let onDimensionColumnResize: ((width: number) => void) | null = null;
  // Height of the whole leaderboard table, so the resize handle can span all
  // rows instead of just the header cell.
  export let tableHeight = 0;
</script>

<thead>
  <tr>
    <th
      aria-label={m.dashboard_comparison_column_aria()}
      class="grid place-content-center"
    >
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

    <th data-dimension-header class:resizable={!!onDimensionColumnResize}>
      <Tooltip location="top">
        <button
          disabled={!allowExpandTable}
          class="text-fg-muted text-left {allowExpandTable
            ? 'hover:text-theme-700'
            : ''}"
          aria-label={m.dashboard_open_dimension_details_aria()}
          onclick={() => setPrimaryDimension(dimensionName)}
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
                class="text-fg-inverse/70 justify-self-start"
                style:max-width="280px"
                aria-label="tooltip-name-description"
              >
                {dimensionDescription}
              </div>
            </div>
          {/if}
        </TooltipContent>
      </Tooltip>

      {#if onDimensionColumnResize}
        <div class="resizer-container" style:height="{tableHeight}px">
          <Resizer
            side="right"
            direction="EW"
            min={MIN_DIMENSION_COLUMN_WIDTH}
            max={MAX_DIMENSION_COLUMN_WIDTH}
            basis={DEFAULT_DIMENSION_COLUMN_WIDTH}
            dimension={dimensionColumnWidth}
            onUpdate={onDimensionColumnResize}
          >
            <div class="resize-bar"></div>
          </Resizer>
        </div>
      {/if}
    </th>

    {#each leaderboardMeasureNames as measureName, index (index)}
      <th data-measure-header>
        <button
          aria-label={m.dashboard_sort_by_value_aria()}
          onclick={() => {
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

      {#if isValidPercentOfTotal(measureName) && measuresWithContext.has(measureName)}
        <th data-percent-of-total-header>
          <button
            aria-label={m.dashboard_sort_by_percent_total_aria()}
            onclick={() => toggleSort(SortType.PERCENT, measureName)}
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

      {#if isTimeComparisonActive && measuresWithContext.has(measureName)}
        <th data-absolute-change-header>
          <button
            aria-label={m.dashboard_sort_by_absolute_change_aria()}
            onclick={() => toggleSort(SortType.DELTA_ABSOLUTE, measureName)}
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

      {#if isTimeComparisonActive && measuresWithContext.has(measureName)}
        <th data-percent-change-header>
          <button
            aria-label={m.dashboard_sort_by_percent_change_aria()}
            onclick={() => toggleSort(SortType.DELTA_PERCENT, measureName)}
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
    /* z-40 keeps the resize handle (which hangs below the header, down the
       whole column) above the rows' sticky dimension cells (z-30). */
    @apply sticky left-0 z-40 bg-surface-background text-left;
  }

  /* Visible separator at the column boundary, marking where to pick up the
     resize handle (same as the expanded dimension table). */
  th[data-dimension-header].resizable {
    @apply border-r;
  }

  .resizer-container {
    @apply absolute top-0 right-0 w-0 z-50;
  }

  .resize-bar {
    @apply bg-primary-500 w-1 h-full;
  }

  th:not(:first-of-type) {
    @apply border-b;
  }

  th:not(:nth-of-type(2)) button {
    @apply justify-end;
  }
</style>
