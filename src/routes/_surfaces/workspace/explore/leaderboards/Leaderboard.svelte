<script lang="ts">
  /**
   * Leaderboard.svelte
   * -------------------------
   * This is the "implemented" feature of the leaderboard, meant to be used
   * in the application itself.
   */
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "$lib/application-state-stores/explorer-stores";
  import LeaderboardContainer from "$lib/components/leaderboard/LeaderboardContainer.svelte";
  import LeaderboardHeader from "$lib/components/leaderboard/LeaderboardHeader.svelte";
  import LeaderboardList from "$lib/components/leaderboard/LeaderboardList.svelte";
  import LeaderboardListItem from "$lib/components/leaderboard/LeaderboardListItem.svelte";
  import Spinner from "$lib/components/Spinner.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import {
    selectDimensionFromMeta,
    selectMeasureFromMeta,
    useMetaQuery,
    useTopListQuery,
  } from "$lib/svelte-query/queries/metrics-view";
  import { slideRight } from "$lib/transitions";
  import {
    humanizeGroupValues,
    NicelyFormattedTypes,
    ShortHandSymbols,
  } from "$lib/util/humanize-numbers";
  import { createEventDispatcher } from "svelte";
  import { getDisplayName } from "../utils";
  import LeaderboardEntrySet from "./DimensionLeaderboardEntrySet.svelte";

  export let metricsDefId: string;
  export let dimensionId: string;
  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */
  export let referenceValue: number;

  export let formatPreset: NicelyFormattedTypes;
  export let leaderboardFormatScale: ShortHandSymbols;
  export let isSummableMeasure = false;

  export let slice = 7;
  export let seeMoreSlice = 50;
  let seeMore = false;

  const dispatch = createEventDispatcher();

  $: metaQuery = useMetaQuery(metricsDefId);

  let dimension: DimensionDefinitionEntity;
  $: dimension = selectDimensionFromMeta($metaQuery.data, dimensionId);
  let displayName: string;
  // TODO: select based on label?
  $: displayName = getDisplayName(dimension);

  let measure: MeasureDefinitionEntity;
  $: measure = selectMeasureFromMeta(
    $metaQuery.data,
    metricsExplorer?.leaderboardMeasureId
  );

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  let activeValues: Array<unknown>;
  $: activeValues =
    metricsExplorer?.filters.include.find((d) => d.name === dimension?.id)
      ?.values ?? [];
  $: atLeastOneActive = !!activeValues?.length;

  function setLeaderboardValues(values) {
    dispatch("leaderboard-value", {
      dimensionId,
      values,
    });
  }

  let topListQuery;

  $: if (
    measure?.id &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    topListQuery = useTopListQuery(metricsDefId, dimensionId, {
      measures: [measure?.id],
      limit: 15,
      offset: 0,
      sort: [
        {
          name: measure?.sqlName,
          direction: "desc",
        },
      ],
      time: {
        start: metricsExplorer?.selectedTimeRange?.start,
        end: metricsExplorer?.selectedTimeRange?.end,
      },
      filter: metricsExplorer?.filters,
    });
  }

  let values = [];

  /** replace data after fetched. */
  $: if (!$topListQuery?.isFetching) {
    values =
      $topListQuery?.data?.data.map((val) => ({
        value: val[measure?.sqlName],
        label: val[dimension?.dimensionColumn],
      })) ?? [];
    setLeaderboardValues(values);
  }
  /** figure out how many selected values are currently hidden */
  // $: hiddenSelectedValues = values.filter((di, i) => {
  //   return activeValues.includes(di.label) && i > slice - 1 && !seeMore;
  // });

  $: if (values) {
    values = formatPreset
      ? humanizeGroupValues(values, formatPreset, {
          scale: leaderboardFormatScale,
        })
      : humanizeGroupValues(values, NicelyFormattedTypes.HUMANIZE, {
          scale: leaderboardFormatScale,
        });
  }

  // get all values that are selected but not visible.
  // we'll put these at the bottom w/ a divider.
  $: selectedValuesThatAreBelowTheFold = activeValues
    ?.filter((label) => {
      return (
        // the value is visible within the fold.
        !values.slice(0, !seeMore ? slice : seeMoreSlice).some((value) => {
          return value.label === label;
        })
      );
    })
    .map((label) => {
      const existingValue = values.find((value) => value.label === label);
      // return the existing value, or if it does not exist, just return the label.
      return existingValue ? { ...existingValue } : { label };
    })
    .sort((a, b) => {
      return b.value - a.value;
    });
</script>

{#if topListQuery}
  <LeaderboardContainer focused={atLeastOneActive}>
    <LeaderboardHeader isActive={atLeastOneActive}>
      <div
        slot="title"
        class:text-gray-500={atLeastOneActive}
        class:italic={atLeastOneActive}
      >
        <Tooltip location="top" distance={16}>
          <div class="flex flex-row gap-x-2 items-center">
            {#if $topListQuery?.isFetching}
              <div transition:slideRight|local={{ leftOffset: 8 }}>
                <Spinner size="16px" status={EntityStatus.Running} />
              </div>
            {/if}
            {displayName}
          </div>
          <TooltipContent slot="tooltip-content">
            <TooltipTitle>
              <svelte:fragment slot="name">
                {displayName}
              </svelte:fragment>
              <svelte:fragment slot="description">dimension</svelte:fragment>
            </TooltipTitle>
            <TooltipShortcutContainer>
              <div>
                {#if dimension?.description}
                  {dimension.description}
                {:else}
                  the leaderboard metrics for {displayName}
                {/if}
              </div>
            </TooltipShortcutContainer>
          </TooltipContent>
        </Tooltip>
      </div>
    </LeaderboardHeader>

    {#if values}
      <LeaderboardList>
        <!-- place the leaderboard entries that are above the fold here -->
        <LeaderboardEntrySet
          loading={$topListQuery?.isFetching}
          values={values.slice(0, !seeMore ? slice : seeMoreSlice)}
          {activeValues}
          {atLeastOneActive}
          {referenceValue}
          {isSummableMeasure}
          on:select-item
        />
        <!-- place the selected values that are not above the fold here -->
        {#if selectedValuesThatAreBelowTheFold?.length}
          <hr />
          <LeaderboardEntrySet
            loading={$topListQuery?.isFetching}
            values={selectedValuesThatAreBelowTheFold}
            {activeValues}
            {atLeastOneActive}
            {referenceValue}
            {isSummableMeasure}
            on:select-item
          />
          <hr />
        {/if}
        {#if $topListQuery?.isError}
          <div class="text-red-500">
            {$topListQuery?.error}
          </div>
        {:else if values.length === 0}
          <div class="p-1 italic text-gray-500">no available values</div>
        {/if}

        {#if values.length > slice}
          <Tooltip location="right">
            <LeaderboardListItem
              value={0}
              color="bg-gray-100"
              on:click={() => {
                seeMore = !seeMore;
              }}
            >
              <div class="italic text-gray-500" slot="title">
                See {#if seeMore}Less{:else}More{/if}
              </div>
            </LeaderboardListItem>
            <TooltipContent slot="tooltip-content"
              >See More Items</TooltipContent
            >
          </Tooltip>
        {/if}
      </LeaderboardList>
    {/if}
  </LeaderboardContainer>
{/if}
